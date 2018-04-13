package mocker

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/loader"
)

type mocker struct {
	w       io.Writer
	src     *string
	pkg     *string
	iface   *[]string
	prefix  *string
	suffix  *string
	imports imports
}

func New(src *string, pkg *string, iface *[]string, prefix, suffix *string, w io.Writer) (*mocker, error) {
	return &mocker{w, src, pkg, iface, prefix, suffix, imports{make(map[string]iimport), make(map[string]iimport)}}, nil
}

func (m *mocker) Mock() error {
	fset := token.NewFileSet()
	testFilter := func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}
	pkgs, err := parser.ParseDir(fset, *m.src, testFilter, parser.SpuriousErrors)
	if err != nil {
		return errors.Wrap(err, "mock: failed to parse src dir")
	}
	if *m.pkg == "" {
		for pkg := range pkgs {
			if strings.Contains(pkg, "_test") {
				continue
			}
			m.pkg = &pkg
			break
		}
	}
	tmpl, err := template.New("mocker").Funcs(tmplFns).Parse(tmpl)
	if err != nil {
		return errors.Wrap(err, "failed to parse template")
	}
	f := file{Pkg: *m.pkg, Imports: []iimport{{Path: "sync"}}}
	pkgInfo, err := m.pkgInfo(*m.src)
	if err != nil {
		return errors.Wrap(err, "failed to get pkg info")
	}
	for _, pkg := range pkgs {
		i := 0
		files := make([]*ast.File, len(pkg.Files))
		for _, f := range pkg.Files {
			files[i] = f
			i++
		}
		for _, f := range files {
			for _, d := range f.Decls {
				gd, ok := d.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, s := range gd.Specs {
					is, ok := s.(*ast.ImportSpec)
					if !ok {
						continue
					}
					if is.Name != nil {
						i := iimport{Name: is.Name.Name, Path: strings.Replace(is.Path.Value, `"`, "", -1)}
						m.imports.named[i.Path] = i
					}
				}
			}
		}
	}
	for _, n := range *m.iface {
		ifaceobj := pkgInfo.Pkg.Scope().Lookup(n)
		if ifaceobj == nil {
			return fmt.Errorf("failed to find interface: %s", n)
		}
		if !types.IsInterface(ifaceobj.Type()) {
			return errors.Wrap(err, fmt.Sprintf("%s (%s) is not an interface", n, ifaceobj.Type().String()))
		}
		iiface := ifaceobj.Type().Underlying().(*types.Interface).Complete()
		iface := iface{Name: n, Suffix: *m.suffix, Prefix: *m.prefix}
		for i := 0; i < iiface.NumMethods(); i++ {
			met := iiface.Method(i)
			sig := met.Type().(*types.Signature)
			m := method{Name: met.Name(), Params: m.params(sig, sig.Params(), "in%d"), Returns: m.params(sig, sig.Results(), "out%d")}
			iface.Methods = append(iface.Methods, m)
		}
		f.Ifaces = append(f.Ifaces, iface)
	}
	for p, n := range m.imports.named {
		if _, ok := m.imports.all[p]; ok {
			m.imports.all[p] = n
		}
	}
	for _, pkg := range m.imports.all {
		f.Imports = append(f.Imports, pkg)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, f); err != nil {
		return errors.Wrap(err, "failed to execute template")
	}
	fmted, err := format.Source(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to format file")
	}
	if _, err := m.w.Write(fmted); err != nil {
		return errors.Wrap(err, "failed to write file")
	}
	return nil
}

type file struct {
	Pkg     string
	Ifaces  []iface
	Imports []iimport
}

type imports struct {
	all   map[string]iimport
	named map[string]iimport
}

type iimport struct {
	Path string
	Name string
}

type iface struct {
	Name    string
	Prefix  string
	Suffix  string
	Methods []method
}

type method struct {
	Name    string
	Params  []param
	Returns []param
}

func (m method) ParamStr() string {
	var params []string
	for _, p := range m.Params {
		params = append(params, p.String())
	}
	return strings.Join(params, ",")
}

func (m method) CallStr() string {
	var params []string
	for _, p := range m.Params {
		params = append(params, p.CallStr())
	}
	return strings.Join(params, ",")
}

func (m method) ReturnStr() string {
	var returns []string
	for _, r := range m.Returns {
		returns = append(returns, r.ReturnStr())
	}
	if len(m.Returns) > 1 {
		return fmt.Sprintf("(%s)", strings.Join(returns, ", "))
	}
	return strings.Join(returns, ",")
}

type param struct {
	Name     string
	Type     string
	Variadic bool
}

func (p param) String() string {
	return fmt.Sprintf("%s %s", p.Name, p.TypeStr())
}

func (p param) CallStr() string {
	if p.Variadic {
		return p.Name + "..."
	}
	return p.Name
}

func (p param) TypeStr() string {
	if p.Variadic {
		return "..." + p.Type[2:]
	}
	return p.Type
}

func (p param) ReturnStr() string {
	return p.Type
}

func (m *mocker) params(sig *types.Signature, tuple *types.Tuple, format string) []param {
	var params []param
	typeq := func(pkg *types.Package) string {
		if *m.pkg == pkg.Name() {
			return ""
		}
		path := pkg.Path()
		wd, err := os.Getwd()
		if err != nil {
			return ""
		}
		if path == "." {
			path = strings.TrimPrefix(wd, os.Getenv("GOPATH")+"/src/")
		} else {
			path = strings.TrimPrefix(path, strings.TrimPrefix(wd, os.Getenv("GOPATH")+"/src/")+"/vendor/")
		}
		name := pkg.Name()
		if i, ok := m.imports.named[path]; ok {
			name = i.Name
			m.imports.all[path] = iimport{Name: name, Path: path}
		} else {
			m.imports.all[path] = iimport{Path: path}
		}
		return name
	}
	for i := 0; i < tuple.Len(); i++ {
		v := tuple.At(i)
		name := v.Name()
		if name == "" {
			name = fmt.Sprintf(format, i+1)
		}
		tname := types.TypeString(v.Type(), typeq)
		variadic := sig.Variadic() && i == tuple.Len()-1 && tname[0:2] == "[]"
		params = append(params, param{Name: name, Type: tname, Variadic: variadic})

	}
	return params
}

func (m *mocker) pkgInfo(src string) (*loader.PackageInfo, error) {
	abs, err := filepath.Abs(src)
	if err != nil {
		return nil, errors.Wrap(err, "faild to get abs src path")
	}
	pkgPath := m.strip(abs)
	conf := loader.Config{
		ParserMode: parser.SpuriousErrors,
		Cwd:        src,
	}
	conf.Import(pkgPath)
	loader, err := conf.Load()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load program")
	}
	pkgInfo := loader.Package(pkgPath)
	if pkgInfo == nil {
		return nil, errors.New("unable to load package")
	}
	return pkgInfo, nil
}

func (m *mocker) strip(pkg string) string {
	for _, path := range strings.Split(os.Getenv("GOPATH"), string(filepath.ListSeparator)) {
		pkg = strings.TrimPrefix(pkg, filepath.Join(path, "src")+"/")
	}
	return pkg
}
