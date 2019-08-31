// Copyright 2019 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package testpkg

func tested() string {
	return "a"
}

func untested() string {
	return "b"
}

func partlytested(b bool) string {
	r := ""
	if b {
		r = "c"
	} else {
		r = "d"
	}
	return r
}
