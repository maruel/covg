// Copyright 2019 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func TestSmokePkg(t *testing.T) {
	defer mockArgs([]string{"covg", "-v", "./testpkg"})()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	if err := mainImpl(); err != nil {
		t.Fatal(err)
	}
	expected := "testpkg.go:11:\tuntested\t  0.0% 11-13\ntestpkg.go:15:\tpartlytested\t 80.0% 17-19\ntotal:\t\t(statements)\t 71.4%\n"
	parts := strings.SplitN(stdout.(*bytes.Buffer).String(), "\n", 2)
	if parts[1] != expected {
		t.Fatalf("Output mismatch\nGot      %q\nExpected %q", parts[1], expected)
	}
}

func TestSmokePkgAll(t *testing.T) {
	defer mockArgs([]string{"covg", "-a", "./testpkg"})()
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	if err := mainImpl(); err != nil {
		t.Fatal(err)
	}
	expected := "github.com/maruel/covg/testpkg/testpkg.go:7:\ttested\t\t100.0% 7-9\ngithub.com/maruel/covg/testpkg/testpkg.go:11:\tuntested\t  0.0% 11-13\ngithub.com/maruel/covg/testpkg/testpkg.go:15:\tpartlytested\t 80.0% 17-19\ntotal:\t\t\t\t\t\t(statements)\t 71.4%\n"
	parts := strings.SplitN(stdout.(*bytes.Buffer).String(), "\n", 2)
	if parts[1] != expected {
		t.Fatalf("Output mismatch\nGot      %q\nExpected %q", parts[1], expected)
	}
}

func mockArgs(args []string) func() {
	oldcmd := flag.CommandLine
	oldargs := os.Args
	os.Args = args
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
	stdout = &bytes.Buffer{}
	return func() {
		flag.CommandLine = oldcmd
		os.Args = oldargs
		log.SetOutput(os.Stderr)
		stdout = os.Stdout
	}
}
