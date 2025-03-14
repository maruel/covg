// Copyright 2015 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// covg: yet another coverage tool.
//
// Use with:
//
//	covg <packages> [--] <go test args>
package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"golang.org/x/tools/cover"
)

// stdout is mocked in test.
var stdout io.Writer = os.Stdout

// errSilent means that the process exit code must be 1.
var errSilent = errors.New("silent error")

// command runs a command and optionally logs it.
func command(ctx context.Context, args ...string) *exec.Cmd {
	log.Printf("%v", args)
	c := exec.CommandContext(ctx, args[0], args[1:]...)
	c.Stdout = stdout
	c.Stderr = os.Stderr
	return c
}

// profile returns the blocks for this function implementation.
func (f *FuncExtent) profile(profile *cover.Profile) []cover.ProfileBlock {
	start := -1
	for i, b := range profile.Blocks {
		if b.StartLine > f.endLine || (b.StartLine == f.endLine && b.StartCol >= f.endCol) {
			// Past the end of the function.
			if start == -1 {
				// TODO(maruel) Temporary to find bugs, if any.
				panic(f.name)
			}
			return profile.Blocks[start:i]
		}
		if b.EndLine < f.startLine || (b.EndLine == f.startLine && b.EndCol <= f.startCol) {
			// Before the beginning of the function
			continue
		}
		if start == -1 {
			start = i
		}
	}
	if start == -1 {
		panic(f.name)
	}
	return profile.Blocks[start:]
}

// formatBlock converts the line numbers of a block into a string.
func formatBlock(b cover.ProfileBlock) string {
	if b.StartLine == b.EndLine {
		return fmt.Sprintf("%d", b.StartLine)
	}
	return fmt.Sprintf("%d-%d", b.StartLine, b.EndLine)
}

// coverageBlocks returns number of statement covered versus the total number
// of statements.
func coverageBlocks(blocks []cover.ProfileBlock) (int64, int64) {
	var covered, total int64
	for _, b := range blocks {
		total += int64(b.NumStmt)
		if b.Count > 0 {
			covered += int64(b.NumStmt)
		}
	}
	return covered, total
}

// extentsBlocks returns a strring representing all the blocks.
func extentsBlocks(blocks []cover.ProfileBlock) string {
	if len(blocks) == 0 {
		return ""
	}
	b := blocks[0]
	b.EndLine = blocks[len(blocks)-1].EndLine
	return formatBlock(b)
}

// allBlocks returns a string representing all the blocks.
func allBlocks(blocks []cover.ProfileBlock) string {
	var out []string
	for _, b := range blocks {
		out = append(out, formatBlock(b))
	}
	return strings.Join(out, ",")
}

// missingBlocks returns a strring representing the lines missing coverage.
func missingBlocks(blocks []cover.ProfileBlock) string {
	var out []string
	var accumul cover.ProfileBlock
	for _, b := range blocks {
		if b.Count > 0 {
			if accumul.StartLine != 0 {
				out = append(out, formatBlock(accumul))
				accumul.StartLine = 0
			}
			continue
		}
		if accumul.StartLine == 0 {
			accumul.StartLine = b.StartLine
		}
		accumul.EndLine = b.EndLine
	}
	if accumul.StartLine != 0 {
		out = append(out, formatBlock(accumul))
	}
	return strings.Join(out, ",")
}

// printCoverageOld is the implementation using go tool cover.
func printCoverageOld(ctx context.Context, name string, all bool) error {
	return command(ctx, "go", "tool", "cover", "-func", name).Run()
}

func commonPrefix(a, b string) string {
	l := len(a)
	if len(b) < l {
		l = len(b)
	}
	i := 0
	for ; i < l && a[i] == b[i]; i++ {
	}
	return a[:i]
}

// printCoverage is our implementation using algorithms!
func printCoverage(ctx context.Context, name string, all bool) error {
	profiles, err := cover.ParseProfiles(name)
	if err != nil {
		return err
	}

	dirs, err := findPkgs(profiles)
	if err != nil {
		return err
	}

	out := bufio.NewWriter(stdout)
	defer out.Flush()
	tabber := tabwriter.NewWriter(out, 1, 8, 1, '\t', 0)
	defer tabber.Flush()

	offset := 0
	if !all {
		var d []string
		for _, profile := range profiles {
			d = append(d, filepath.Dir(profile.FileName))
		}
		if len(d) == 1 {
			if offset = len(d[0]); offset != 0 {
				// Also trim trailing path separator.
				offset++
			}
		} else if len(d) != 0 {
			prefix := d[0]
			for i := 1; i < len(d); i++ {
				prefix = commonPrefix(prefix, d[i])
			}
			if offset = len(prefix); offset != 0 {
				// Also trim trailing path separator.
				offset++
			}
		}
	}

	var total, covered int64
	for _, profile := range profiles {
		fn := profile.FileName
		file, err := findFile(dirs, fn)
		if err != nil {
			return err
		}
		funcs, err := findFuncs(file)
		if err != nil {
			return err
		}
		// Now match up functions and profile blocks.
		for _, f := range funcs {
			blocks := f.profile(profile)
			c, t := coverageBlocks(blocks)
			// Discount lack of coverage of function "_".
			if f.name == "_" && !all {
				continue
			}
			total += t
			covered += c
			if c == t {
				if all {
					fmt.Fprintf(tabber, "%s:%d:\t%s\t%5.1f%% %s\n", fn[offset:], f.startLine, f.name, percent(c, t), extentsBlocks(blocks))
				}
				continue
			}
			fmt.Fprintf(tabber, "%s:%d:\t%s\t%5.1f%% %s\n", fn[offset:], f.startLine, f.name, percent(c, t), missingBlocks(blocks))
		}
	}
	fmt.Fprintf(tabber, "total:\t(statements)\t%5.1f%%\n", percent(covered, total))
	return nil
}

// runCover runs the test under coverage and prints the coverage.
func runCover(ctx context.Context, pkgs, extraArgs []string, all bool) error {
	log.Printf("runCover(%#v, %s, %t)", pkgs, extraArgs, all)
	f, err := os.CreateTemp("", "covg")
	if err != nil {
		return err
	}
	name := f.Name()
	_ = f.Close()
	args := append([]string{"go", "test", "-covermode=count", "-coverprofile", name}, extraArgs...)
	c := command(ctx, append(args, pkgs...)...)
	if err = c.Run(); err == nil {
		err = printCoverage(ctx, name, all)
	}
	if err != nil {
		err = errSilent
	}
	if err2 := os.Remove(name); err == nil {
		err = err2
	}
	return err
}

// getPackages resolves the provided packages into canonical format.
func getPackages(ctx context.Context, pkgs []string) []string {
	c := command(ctx, append([]string{"go", "list", "-f", "{{.ImportPath}}"}, pkgs...)...)
	b := bytes.Buffer{}
	c.Stdout = &b
	if c.Run() != nil {
		return nil
	}
	// TODO(maruel): Windows.
	return strings.Split(strings.TrimSpace(b.String()), "\n")
}

func mainImpl() error {
	all := flag.Bool("a", false, "show functions with 100% coverage")
	verboseFlag := flag.Bool("v", false, "enable logging")
	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return errSilent
	}

	log.SetFlags(log.Lmicroseconds)
	if !*verboseFlag {
		log.SetOutput(io.Discard)
	}

	args := flag.Args()
	var extraArgs []string
	for i, a := range args {
		if a == "--" {
			extraArgs = args[i+1:]
			args = args[:i]
			break
		}
		if strings.HasPrefix(a, "-") {
			// For the case "covg -- -v", flag.Parse() strips the '--' argument.
			extraArgs = args
			args = nil
		}
	}
	if len(args) == 0 {
		args = []string{"."}
	}

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	go func() {
		<-c
		cancel()
	}()
	signal.Notify(c, os.Interrupt)

	pkgs := getPackages(ctx, args)
	if len(pkgs) == 0 {
		return errors.New("invalid path")
	}
	sort.Strings(pkgs)
	err := runCover(ctx, pkgs, extraArgs, *all)
	if err == nil {
		return err
	}
	if ctx.Err() != nil {
		return errSilent
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		if err != errSilent {
			fmt.Fprintf(os.Stderr, "covg: %s\n", err)
		}
		os.Exit(1)
	}
}
