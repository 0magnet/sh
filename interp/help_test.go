package interp_test

import (
	"context"
	"strings"
	"testing"

	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func runSrc(t *testing.T, src string) string {
	t.Helper()
	var out strings.Builder
	r, err := interp.New(interp.StdIO(strings.NewReader(""), &out, &out))
	if err != nil {
		t.Fatal(err)
	}
	f, err := syntax.NewParser().Parse(strings.NewReader(src), "")
	if err != nil {
		t.Fatal(err)
	}
	_ = r.Run(context.Background(), f)
	return out.String()
}

func TestHelpBuiltin(t *testing.T) {
	all := runSrc(t, "help")
	for _, want := range []string{"defined internally", "A star (*)", "cd [-L|[-P [-e]]] [dir]", "*jobs"} {
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
	if j := runSrc(t, "help jobs"); !strings.Contains(j, "not implemented") {
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
