package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	distDir  = flag.String("dist", "dist", "output dist directory")
	platform = flag.String("platform", "", "target platform (windows|linux|darwin)")
	arch     = flag.String("arch", "", "target architecture (amd64|386|arm64)")
	all      = flag.Bool("all", false, "build all supported targets")
	clean    = flag.Bool("clean", false, "remove dist directory")
	version  = flag.String("version", "", "version string to embed (default: git describe/dev)")
)

var supportedPlatforms = []string{"windows", "linux", "darwin"}
var supportedArchs = []string{"amd64", "386", "arm64"}

var ideAliases = map[string]string{
	"gcursor":   "cursor",
	"gwindsurf": "windsurf",
	"gzed":      "zed",
	"gtrae":     "trae",
}

func main() {
	flag.Parse()

	repoRoot, err := findRepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	if !*all && !*clean {
		if *platform == "" {
			*platform = runtime.GOOS
		}
		if *arch == "" {
			*arch = runtime.GOARCH
		}
	}

	if os.Getenv("GOCACHE") == "" {
		_ = os.Setenv("GOCACHE", filepath.Join(repoRoot, ".gocache"))
	}

	if *clean {
		if err := os.RemoveAll(filepath.Join(repoRoot, *distDir)); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		return
	}

	v := strings.TrimSpace(*version)
	if v == "" {
		v = detectVersion(repoRoot)
	}

	if *all {
		for _, p := range supportedPlatforms {
			archs := supportedArchs
			if p == "darwin" {
				archs = []string{"arm64", "amd64"}
			}
			for _, a := range archs {
				if err := buildOne(repoRoot, p, a, v); err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					os.Exit(1)
				}
			}
		}
		return
	}

	if err := validateTarget(*platform, *arch); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
	if err := buildOne(repoRoot, *platform, *arch, v); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func validateTarget(p, a string) error {
	if p == "" || a == "" {
		return errors.New("platform and arch are required (e.g. -platform=linux -arch=amd64)")
	}
	if !contains(supportedPlatforms, p) {
		return fmt.Errorf("unsupported platform %q (supported: %s)", p, strings.Join(supportedPlatforms, ", "))
	}
	if !contains(supportedArchs, a) {
		return fmt.Errorf("unsupported arch %q (supported: %s)", a, strings.Join(supportedArchs, ", "))
	}
	if p == "darwin" && a == "386" {
		return fmt.Errorf("unsupported target %s/%s", p, a)
	}
	return nil
}

func buildOne(repoRoot, goos, goarch, v string) error {
	outDir := filepath.Join(repoRoot, *distDir, fmt.Sprintf("%s-%s", goos, goarch))
	binDir := filepath.Join(outDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return err
	}

	exeSuffix := ""
	if goos == "windows" {
		exeSuffix = ".exe"
	}

	gsshName := "gssh" + exeSuffix
	if goos == "windows" {
		gsshName = "gssh-core" + exeSuffix
	}

	if err := goBuild(repoRoot, goos, goarch, v, filepath.Join(binDir, gsshName), "./cmd/gssh"); err != nil {
		return err
	}
	if err := goBuild(repoRoot, goos, goarch, v, filepath.Join(binDir, "gssh-ipc"+exeSuffix), "./cmd/ipc"); err != nil {
		return err
	}
	if err := goBuild(repoRoot, goos, goarch, v, filepath.Join(binDir, "gcode"+exeSuffix), "./cmd/gcode"); err != nil {
		return err
	}

	if goos == "windows" {
		if err := copyFile(
			filepath.Join(repoRoot, "cmd", "gssh", "bat", "gssh.cmd"),
			filepath.Join(binDir, "gssh.cmd"),
			0,
		); err != nil {
			return err
		}
		if err := copyFile(
			filepath.Join(repoRoot, "cmd", "gssh", "bat", "gssh.ps1"),
			filepath.Join(binDir, "gssh.ps1"),
			0,
		); err != nil {
			return err
		}

		if err := copyFile(
			filepath.Join(repoRoot, "cmd", "gcode", "bat", "ssh-wrapper.bat"),
			filepath.Join(binDir, "ssh-wrapper.bat"),
			0,
		); err != nil {
			return err
		}
		if err := copyFile(
			filepath.Join(repoRoot, "cmd", "gcode", "bat", "ssh-wrapper.ps1"),
			filepath.Join(binDir, "ssh-wrapper.ps1"),
			0,
		); err != nil {
			return err
		}

		for alias, ide := range ideAliases {
			path := filepath.Join(binDir, alias+".cmd")
			content := fmt.Sprintf("@echo off\r\n\"%%~dp0gcode.exe\" --ide %s %%*\r\n", ide)
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return err
			}
		}
	} else {
		if err := copyFile(
			filepath.Join(repoRoot, "cmd", "gcode", "sh", "ssh-wrapper"),
			filepath.Join(binDir, "ssh-wrapper"),
			0o755,
		); err != nil {
			return err
		}

		for alias := range ideAliases {
			dst := filepath.Join(binDir, alias)
			_ = os.Remove(dst)
			if err := os.Symlink("gcode", dst); err != nil {
				return err
			}
		}
	}

	return nil
}

func goBuild(repoRoot, goos, goarch, v, out, pkg string) error {
	ldflags := fmt.Sprintf("-X main.version=%s", v)
	args := []string{
		"build",
		"-ldflags",
		ldflags,
		"-o",
		out,
		pkg,
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(),
		"GOOS="+goos,
		"GOARCH="+goarch,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}

	if mode != 0 {
		return os.Chmod(dst, mode)
	}
	return nil
}

func detectVersion(repoRoot string) string {
	cmd := exec.Command("git", "describe", "--tags", "--always", "--dirty")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err == nil {
		v := strings.TrimSpace(string(out))
		if v != "" {
			return v
		}
	}

	// Fallback: if VERSION file exists, use v<content>, else dev.
	data, err := os.ReadFile(filepath.Join(repoRoot, "VERSION"))
	if err == nil {
		t := strings.TrimSpace(string(data))
		if t != "" {
			return "v" + t
		}
	}
	return "dev"
}

func findRepoRoot() (string, error) {
	start, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod from %s", start)
		}
		dir = parent
	}
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
