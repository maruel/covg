// Copyright 2019 Marc-Antoine Ruel. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package testpkg

import "testing"

func TestTested(t *testing.T) {
	if s := tested(); s != "a" {
		t.Fatal(s)
	}
}

func TestPartlyTested(t *testing.T) {
	if s := partlytested(false); s != "d" {
		t.Fatal(s)
	}
}
