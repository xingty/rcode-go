package ssh

import "testing"

func TestAnalyzeSSHArgsDestinationIndex(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantDest int
		wantTTY  bool
	}{
		{name: "bare", args: []string{"user@host"}, wantDest: 0},
		{name: "identitySeparate", args: []string{"-i", "id_rsa", "user@host"}, wantDest: 2},
		{name: "identityAttached", args: []string{"-i./id_rsa", "user@host"}, wantDest: 1},
		{name: "portSeparate", args: []string{"-p", "22", "user@host"}, wantDest: 2},
		{name: "portAttached", args: []string{"-p22", "user@host"}, wantDest: 1},
		{name: "optionOSeparate", args: []string{"-o", "StrictHostKeyChecking=no", "user@host"}, wantDest: 2},
		{name: "optionOAttached", args: []string{"-oStrictHostKeyChecking=no", "user@host"}, wantDest: 1},
		{name: "ttyCombined", args: []string{"-tt", "user@host"}, wantDest: 1, wantTTY: true},
		{name: "endOfOptions", args: []string{"--", "-weirdhost"}, wantDest: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, err := analyzeSSHArgs(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if inv.destinationIndex != tt.wantDest {
				t.Fatalf("destinationIndex=%d want=%d", inv.destinationIndex, tt.wantDest)
			}
			if inv.hasPseudoTTY != tt.wantTTY {
				t.Fatalf("hasPseudoTTY=%v want=%v", inv.hasPseudoTTY, tt.wantTTY)
			}
			if inv.disableReason != "" {
				t.Fatalf("disableReason=%q want empty", inv.disableReason)
			}
		})
	}
}

func TestAnalyzeSSHArgsDisableReasons(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantReason  string
		wantHasDest bool
	}{
		{name: "disableT", args: []string{"-vT", "user@host"}, wantReason: "-T", wantHasDest: true},
		{name: "disableR", args: []string{"-R", "/tmp/sock:127.0.0.1:7532", "user@host"}, wantReason: "-R", wantHasDest: true},
		{name: "remoteCommand", args: []string{"user@host", "ls", "-l"}, wantReason: "remote command", wantHasDest: true},
		{name: "helpH", args: []string{"-h"}, wantReason: "-h", wantHasDest: false},
		{name: "helpQuestion", args: []string{"-?"}, wantReason: "-h", wantHasDest: false},
		{name: "versionV", args: []string{"-V"}, wantReason: "-V", wantHasDest: false},
		{name: "queryQ", args: []string{"-Q", "cipher"}, wantReason: "-Q", wantHasDest: false},
		{name: "configG", args: []string{"-G", "user@host"}, wantReason: "-G", wantHasDest: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv, err := analyzeSSHArgs(tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if inv.disableReason != tt.wantReason {
				t.Fatalf("disableReason=%q want=%q", inv.disableReason, tt.wantReason)
			}
			if tt.wantHasDest && inv.destinationIndex < 0 {
				t.Fatalf("expected destinationIndex to be set, got %d", inv.destinationIndex)
			}
		})
	}
}

func TestAnalyzeSSHArgsNoHost(t *testing.T) {
	_, err := analyzeSSHArgs([]string{"-v"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestBuildSSHArgs_PassthroughOnHelpAndVersion(t *testing.T) {
	tests := [][]string{
		{"-h"},
		{"-?"},
		{"-V"},
		{"-Q", "cipher"},
	}

	for _, sshArgs := range tests {
		got, err := BuildSSHArgs("127.0.0.1", 7532, 123, sshArgs)
		if err != nil {
			t.Fatalf("sshArgs=%v unexpected error: %v", sshArgs, err)
		}
		if len(got) != len(sshArgs) {
			t.Fatalf("sshArgs=%v got=%v", sshArgs, got)
		}
		for i := range got {
			if got[i] != sshArgs[i] {
				t.Fatalf("sshArgs=%v got=%v", sshArgs, got)
			}
		}
	}
}
