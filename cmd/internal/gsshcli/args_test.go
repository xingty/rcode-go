package gsshcli

import "testing"

func TestParseArgs_HelpVersionAreInternalOnly(t *testing.T) {
	opts, sshArgs, err := ParseArgs([]string{"--help"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.Help || opts.Version {
		t.Fatalf("opts=%+v", opts)
	}
	if len(sshArgs) != 0 {
		t.Fatalf("sshArgs=%v want empty", sshArgs)
	}

	opts, sshArgs, err = ParseArgs([]string{"--version"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Help || !opts.Version {
		t.Fatalf("opts=%+v", opts)
	}
	if len(sshArgs) != 0 {
		t.Fatalf("sshArgs=%v want empty", sshArgs)
	}
}

func TestParseArgs_HostPortAndPassthrough(t *testing.T) {
	opts, sshArgs, err := ParseArgs([]string{"--host", "10.0.0.1", "--port=2200", "user@host", "-p", "22"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Host != "10.0.0.1" || opts.Port != 2200 {
		t.Fatalf("opts=%+v", opts)
	}
	if got, want := len(sshArgs), 3; got != want {
		t.Fatalf("len(sshArgs)=%d want=%d sshArgs=%v", got, want, sshArgs)
	}
}

func TestParseArgs_DoubleDashStopsParsing(t *testing.T) {
	opts, sshArgs, err := ParseArgs([]string{"--host", "10.0.0.1", "--", "--host", "evil", "-h"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if opts.Host != "10.0.0.1" {
		t.Fatalf("opts=%+v", opts)
	}
	if got, want := sshArgs, []string{"--host", "evil", "-h"}; !equalStrings(got, want) {
		t.Fatalf("sshArgs=%v want=%v", got, want)
	}
}

func TestParseArgs_ShortHIsPassthrough(t *testing.T) {
	_, sshArgs, err := ParseArgs([]string{"-h"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := sshArgs, []string{"-h"}; !equalStrings(got, want) {
		t.Fatalf("sshArgs=%v want=%v", got, want)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
