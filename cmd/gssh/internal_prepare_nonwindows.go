//go:build !windows

package main

func maybeHandleInternalPrepare(argv []string) (handled bool, exitCode int) {
	return false, 0
}
