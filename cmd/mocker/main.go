package main

import (
	"log"

	"github.com/travisjeffery/mocker/pkg/mocker"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	src    = kingpin.Arg("src", "File to find interfaces.").String()
	dst    = kingpin.Flag("dst", "File write mocks. Leave blank to write to Stdout.").String()
	pkg    = kingpin.Flag("pkg", "Name of package for mocks. Inferred by default.").String()
	prefix = kingpin.Flag("prefix", "Prefix of mock names.").Default("Mock").String()
	suffix = kingpin.Flag("suffix", "Suffix of mock names.").String()
	intfs  = kingpin.Arg("ifaces", "Interfaces to mock.").Strings()
)

func main() {
	kingpin.Parse()

	m, err := mocker.New(*src, *dst, *pkg, *prefix, *suffix, *intfs)
	if err != nil {
		log.Fatal("mocker: failed to instantiate")
	}

	if err = m.Mock(); err != nil {
		log.Fatalf("mocker: failed to mock: %v", err)
	}
}
