//go:build !windows

package main

import "testing"

func TestInternalPrepareIsUnavailableOnNonWindows(t *testing.T) {
	handled, exitCode := maybeHandleInternalPrepare([]string{"__prepare", "user@host"})
	if handled || exitCode != 0 {
		t.Fatalf("handled=%v exitCode=%d want handled=false exitCode=0", handled, exitCode)
	}
}
