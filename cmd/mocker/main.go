package main

import (
	"log"

	"github.com/travisjeffery/mocker/pkg/mocker"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	src     = kingpin.Arg("src", "File to find interfaces.").String()
	dst     = kingpin.Flag("dst", "File write mocks. Leave blank to write to Stdout.").String()
	pkg     = kingpin.Flag("pkg", "Name of package for mocks. Inferred by default.").String()
	prefix  = kingpin.Flag("prefix", "Prefix of mock names.").Default("Mock").String()
	suffix  = kingpin.Flag("suffix", "Suffix of mock names.").String()
	selfpkg = kingpin.Flag("selfpkg", "The full package import path for the generated code. The purpose of this flag is to prevent import cycles in the generated code by trying to include its own package. This can happen if the mock's package is set to one of its inputs (usually the main one) and the output is stdio so mocker cannot detect the final output package. Setting this flag will then tell mocker which import to exclude.").String()
	intfs   = kingpin.Arg("ifaces", "Interfaces to mock. Leave empty to mock every interface in the file.").Strings()
)

func main() {
	kingpin.Parse()

	m, err := mocker.New(*src, *dst, *pkg, *prefix, *suffix, *selfpkg, *intfs)
	if err != nil {
		log.Fatal("mocker: failed to instantiate")
	}

	if err = m.Mock(); err != nil {
		log.Fatalf("mocker: failed to mock: %v", err)
	}
}
