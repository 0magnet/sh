package interp_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/0magnet/sh/v3/interp"
	"github.com/0magnet/sh/v3/syntax"
)

func TestHelpBuiltin(t *testing.T) {
	all := runSrc(t, "help")
	for _, want := range []string{"defined internally", "A star (*)", "cd [-L|[-P [-e]]] [dir]", "*ulimit"} {
		if !strings.Contains(all, want) {
			t.Fatalf("help output missing %q:\n%s", want, all[:min(400, len(all))])
		}
	}
	one := runSrc(t, "help cd")
	if !strings.Contains(one, "cd: cd [-L") || !strings.Contains(one, "working directory") {
		t.Fatalf("help cd = %q", one)
	}
	if s := runSrc(t, "help -s pwd"); strings.TrimSpace(s) != "pwd: pwd [-LP]" {
		t.Fatalf("help -s pwd = %q", s)
	}
	if d := runSrc(t, "help -d true"); !strings.Contains(d, "true - Return a successful result.") {
		t.Fatalf("help -d true = %q", d)
	}
	if g := runSrc(t, "help 'ex*'"); !strings.Contains(g, "exec:") || !strings.Contains(g, "exit:") || !strings.Contains(g, "export:") {
		t.Fatalf("glob match = %q", g)
	}
	if miss := runSrc(t, "help nosuchthing"); !strings.Contains(miss, "no help topics match") {
		t.Fatalf("miss = %q", miss)
	}
	// an unimplemented builtin is disclosed, not silently listed
	if u := runSrc(t, "help ulimit"); !strings.Contains(u, "not implemented") {
		t.Fatalf("help ulimit = %q", u)
	}
	// one that works differently here says so
	if j := runSrc(t, "help jobs"); !strings.Contains(j, "controlling terminal") {
		t.Fatalf("help jobs = %q", j)
	}
}

func TestUmaskAndTimes(t *testing.T) {
	if s := runSrc(t, "umask"); strings.TrimSpace(s) != "0022" {
		t.Fatalf("umask = %q", s)
	}
	if s := runSrc(t, "umask -S"); strings.TrimSpace(s) != "u=rwx,g=rx,o=rx" {
		t.Fatalf("umask -S = %q", s)
	}
	if s := runSrc(t, "umask 077; umask"); strings.TrimSpace(s) != "0077" {
		t.Fatalf("umask set = %q", s)
	}
	if s := runSrc(t, "umask 077; umask -S"); strings.TrimSpace(s) != "u=rwx,g=,o=" {
		t.Fatalf("umask -S after set = %q", s)
	}
	if s := runSrc(t, "times"); !strings.Contains(s, "0m0.000s") {
		t.Fatalf("times = %q", s)
	}
}

func TestUmaskAppliesToRedirections(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want os.FileMode
	}{
		{"echo hi >f", 0o644},
		{"umask 077; echo hi >f", 0o600},
		{"umask 002; echo hi >>f", 0o664},
	} {
		var got os.FileMode
		open := func(ctx context.Context, path string, flag int, perm os.FileMode) (io.ReadWriteCloser, error) {
			got = perm
			return nopFile{}, nil
		}
		r, err := interp.New(interp.OpenHandler(open))
		if err != nil {
			t.Fatal(err)
		}
		f, err := syntax.NewParser().Parse(strings.NewReader(tc.src), "")
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Run(context.Background(), f); err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("%q created mode %v, want %v", tc.src, got, tc.want)
		}
	}
}

type nopFile struct{}

func (nopFile) Read([]byte) (int, error)    { return 0, io.EOF }
func (nopFile) Write(p []byte) (int, error) { return len(p), nil }
func (nopFile) Close() error                { return nil }
