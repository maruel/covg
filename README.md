# covg

Yet another CLI coverage tool.

It is essentially `go tool cover -func` but with much more fine tuned output.

It is particularly useful when you have a large codebase and you want to see
the coverage of a single package. Many IDEs will run *all tests* from *all
packages* and this may be too slow.

## Install

    go install github.com/maruel/covg@latest

## Usage

    covg <packages> [--] <go test args>


Here's how it looks like:

```
$ covg .
ok      github.com/maruel/covg  0.755s  coverage: 62.5% of statements
func.go:42:     funcOutput                0.0% 42-90
func.go:94:     findFuncs                85.7% 97-99
func.go:127:    Visit                    87.5% 130-132
func.go:149:    coverage                  0.0% 149-168
func.go:180:    findPkgs                 88.5% 185-187,203-205,213-215
func.go:222:    findFile                 55.6% 223-226,232-236
func.go:239:    percent                  66.7% 240-242
main.go:49:     profile                  78.6% 54-57,69-70
main.go:76:     formatBlock              66.7% 77-79
main.go:97:     extentsBlocks            80.0% 98-100
main.go:107:    allBlocks                 0.0% 107-112
main.go:139:    printCoverageOld          0.0% 139-141
main.go:143:    commonPrefix              0.0% 143-151
main.go:155:    printCoverage            76.6% 157-159,162-164,182-190,198-200,202-204,210-211
main.go:229:    runCover                 86.7% 232-234,242-244
main.go:252:    getPackages              83.3% 256-258
main.go:263:    mainImpl                 68.6% 266-268,278-281,283-287,289-291,302-304,310-313
main.go:316:    main                      0.0% 316-321
total:          (statements)             62.5%
```

Pass `-a` to also print the function covered at 100%.

## Imports

`func.go` was copied from
https://go.googlesource.com/go/+/refs/tags/go1.13rc2/src/cmd/cover/func.go with
a minor import fixes.
