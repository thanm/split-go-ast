// Copyright 2023 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestBasic(t *testing.T) {
	tdir := t.TempDir()

	if false {
		tdir = "/tmp/qqq"
		os.RemoveAll(tdir)
		os.Mkdir(tdir, 0777)
	}

	// Do a build of . into <tmpdir>/out.exe with -W=2
	exe := filepath.Join(tdir, "out.exe")
	gotoolpath := filepath.Join(runtime.GOROOT(), "bin", "go")
	cmd := exec.Command(gotoolpath, "build", "-o", exe, ".")
	//t.Logf("cmd: %+v\n", cmd)
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Logf("build: %s\n", b)
		t.Fatalf("build error: %v", err)
	}

	// Create tiny package.
	const tiny = `package tiny
func foo() int { return 42 }
func bar() int { return -42 }
`
	tp := filepath.Join(tdir, "tiny.go")
	if err := os.WriteFile(tp, []byte(tiny), 0666); err != nil {
		t.Fatalf("writing tiny.go: %v", err)
	}

	// Compile tiny package with -W=2.
	errpath := filepath.Join(tdir, "err.txt")
	errf, oerr := os.OpenFile(errpath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if oerr != nil {
		t.Fatalf("opening err file: %v", oerr)
	}
	cmd = exec.Command(gotoolpath, "build", "-gcflags=-W=2", tp)
	cmd.Stdout = errf
	cmd.Stderr = errf
	if err := cmd.Run(); err != nil {
		t.Fatalf("tiny build -W=2 failed: %v", err)
	}
	if err := errf.Close(); err != nil {
		t.Fatalf("closing err file: %v", err)
	}

	// Now run this prog on errpath.
	cmd = exec.Command(exe, "-func=bar", "-phase=escape", "-i="+errpath)
	t.Logf("cmd: %+v\n", cmd)
	var output string
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Logf("run: %s\n", b)
		t.Fatalf("run error: %v", err)
	} else {
		output = string(b)
	}

	lines := strings.Split(output, "\n")
	want := "before escape bar"
	bad := false
	if lines[0] != want {
		t.Errorf("lines[0] wanted %s got %s", want, lines[0])
		bad = true
	}
	wantlen := 10
	if len(lines) != wantlen {
		t.Errorf("len(lines) wanted %d got %d", wantlen, len(lines))
		bad = true
	}

	if bad {
		t.Logf("output is %s\n", output)
	}
}
