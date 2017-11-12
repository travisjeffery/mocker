package mocker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/errors"
)

type importer struct {
	src  string
	base types.Importer
	pkgs map[string]*types.Package
}

func (i *importer) Import(path string) (*types.Package, error) {
	var err error
	if path == "" || path[0] == '.' {
		path, err = filepath.Abs(filepath.Clean(path))
		if err != nil {
			return nil, errors.Wrap(err, "importer: failed to get path")
		}
	}
	if pkg, ok := i.pkgs[path]; ok {
		return pkg, nil
	}
	pkg, err := i.pkg(path)
	if err != nil {
		return nil, errors.Wrap(err, "importer: failed to read pkg")
	}
	i.pkgs[path] = pkg
	return pkg, nil
}

func (i *importer) pkg(pkg string) (*types.Package, error) {
	paths := []string{
		filepath.Join(i.src, "vendor", pkg),
		filepath.Join(os.Getenv("GOPATH"), "src", pkg),
		filepath.Join(os.Getenv("GOROOT"), "src", pkg),
	}
	var fpath string
	var errs []error
	for _, p := range paths {
		abs, err := filepath.Abs(p)
		if err != nil {
			errs = append(errs, errors.Wrap(err, "importer: failed to get abs path"))
			continue
		}
		if fi, err := os.Stat(abs); err != nil {
			errs = append(errs, errors.Wrap(err, "importer: failed stat'ing path"))
			continue
		} else if !fi.IsDir() {
			errs = append(errs, errors.Wrap(err, "importer: path not dir"))
			continue
		}
		fpath = abs
	}
	if len(errs) == 3 {
		return nil, fmt.Errorf("importer: failed to find pkg in vendor, GOPATH, or GOROOT:\n\t%v", errs)
	}
	f, err := ioutil.ReadDir(fpath)
	if err != nil {
		return nil, errors.Wrap(err, "importer: failed to read pkg dir")
	}
	fset := token.NewFileSet()
	var files []*ast.File
	for _, fi := range f {
		if fi.IsDir() {
			continue
		}
		n := fi.Name()
		if path.Ext(n) != ".go" {
			continue
		}
		p := path.Join(fpath, n)
		src, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, errors.Wrap(err, "importer: failed to read file")
		}
		f, err := parser.ParseFile(fset, p, src, 0)
		if err != nil {
			return nil, errors.Wrap(err, "importer: failed to parse file")
		}
		files = append(files, f)
	}
	cfg := types.Config{Importer: i}
	p, err := cfg.Check(pkg, fset, files, nil)
	if err != nil {
		if p, err = i.base.Import(pkg); err != nil {
			return nil, errors.Wrap(err, "importer: failed to import pkg")
		}
	}
	return p, nil
}
