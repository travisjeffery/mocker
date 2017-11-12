package mocker

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	goimporter "go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/pkg/errors"
)

type mocker struct {
	src     *string
	pkg     *string
	iface   *[]string
	imports map[string]bool
	w       io.Writer
}

func New(src *string, pkg *string, iface *[]string, w io.Writer) (*mocker, error) {
	return &mocker{src, pkg, iface, make(map[string]bool), w}, nil
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
	if m.pkg != nil {
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
		return errors.Wrap(err, "mocker: failed to parse template")
	}
	// TODO imports
	f := file{Pkg: *m.pkg, Imports: []string{"sync"}}
	for _, pkg := range pkgs {
		i := 0
		files := make([]*ast.File, len(pkg.Files))
		for _, f := range pkg.Files {
			files[i] = f
			i++
		}
		cfg := types.Config{Error: func(err error) { fmt.Println("err", err) }, Importer: &importer{src: *m.src, pkgs: make(map[string]*types.Package), base: goimporter.Default()}}
		tpkg, err := cfg.Check(*m.src, fset, files, nil)
		if err != nil {
			return errors.Wrap(err, "mocker: failed to type check pkg")
		}
		for _, i := range *m.iface {
			ifaceobj := tpkg.Scope().Lookup(i)
			if ifaceobj == nil {
				return fmt.Errorf("mocker: failed to find interface %s", i)
			}
			if !types.IsInterface(ifaceobj.Type()) {
				return fmt.Errorf("mocker: not an interface %s", i)
			}
			tiface := ifaceobj.Type().Underlying().(*types.Interface).Complete()
			iface := iface{Name: i}
			for i := 0; i < tiface.NumMethods(); i++ {
				met := tiface.Method(i)
				sig := met.Type().(*types.Signature)
				m := method{Name: met.Name(), Params: m.params(sig, sig.Params(), "in%d"), Returns: m.params(sig, sig.Results(), "out%d")}
				iface.Methods = append(iface.Methods, m)
			}
			f.Ifaces = append(f.Ifaces, iface)
		}
	}
	for pkg := range m.imports {
		f.Imports = append(f.Imports, pkg)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, f); err != nil {
		return errors.Wrap(err, "mocker: failed to execute template")
	}
	fmt.Println(buf.String())
	fmted, err := format.Source(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "mocker: failed to format file")
	}
	if _, err := m.w.Write(fmted); err != nil {
		return errors.Wrap(err, "mocker: failed to write file")
	}
	return nil
}

type file struct {
	Pkg     string
	Ifaces  []iface
	Imports []string
}

type iface struct {
	Name    string
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
		path := pkg.Path()
		if path == "." {
			// TODO
		}
		m.imports[pkg.Name()] = true
		return pkg.Name()
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
