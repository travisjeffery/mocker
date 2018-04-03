package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/travisjeffery/mocker/pkg/mocker"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	out    = kingpin.Flag("out", "File to write mocks to. Stdout by default.").String()
	pkg    = kingpin.Flag("pkg", "Name of package for mocks. Inferred by default.").String()
	src    = kingpin.Arg("src", "Directory to find interfaces.").Required().String()
	iface  = kingpin.Arg("ifaces", "Interfaces to mock.").Required().Strings()
	prefix = kingpin.Flag("prefix", "Prefix for mocks.").Default("Mock").String()
	suffix = kingpin.Flag("suffix", "Suffix for mocks.").String()
)

func main() {
	kingpin.Parse()

	var buf bytes.Buffer
	var w io.Writer

	w = os.Stdout
	if out != nil {
		w = &buf
	}

	m, err := mocker.New(src, pkg, iface, prefix, suffix, w)
	if err != nil {
		log.Fatal("mocker: failed to instantiate")
	}

	if err = m.Mock(); err != nil {
		log.Fatalf("mocker: failed to mock: %v", err)
	}

	if out != nil {
		ioutil.WriteFile(*out, buf.Bytes(), 0777)
	}
}
