package gsshcli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Options struct {
	Host    string
	Port    int
	Help    bool
	Version bool
}

func DefaultOptions() Options {
	return Options{
		Host: "127.0.0.1",
		Port: 7532,
	}
}

func ParseArgs(argv []string) (Options, []string, error) {
	opts := DefaultOptions()
	sshArgs := make([]string, 0, len(argv))

	for i := 0; i < len(argv); i++ {
		arg := argv[i]

		if arg == "--" {
			sshArgs = append(sshArgs, argv[i+1:]...)
			break
		}

		if arg == "--help" {
			opts.Help = true
			continue
		}
		if arg == "--version" {
			opts.Version = true
			continue
		}

		if v, ok := cutKV(arg, "--host"); ok {
			opts.Host = v
			continue
		}
		if arg == "--host" || arg == "-host" {
			i++
			if i >= len(argv) {
				return Options{}, nil, errors.New("missing value for --host")
			}
			opts.Host = argv[i]
			continue
		}

		if v, ok := cutKV(arg, "--port"); ok {
			port, err := parsePort(v)
			if err != nil {
				return Options{}, nil, err
			}
			opts.Port = port
			continue
		}
		if arg == "--port" || arg == "-port" {
			i++
			if i >= len(argv) {
				return Options{}, nil, errors.New("missing value for --port")
			}
			port, err := parsePort(argv[i])
			if err != nil {
				return Options{}, nil, err
			}
			opts.Port = port
			continue
		}

		sshArgs = append(sshArgs, arg)
	}

	return opts, sshArgs, nil
}

func Usage(program string) string {
	return strings.TrimSpace(fmt.Sprintf(`
Usage:
  %s [--host <host>] [--port <port>] [--] <ssh-args...>
  %s --help
  %s --version
`, program, program, program))
}

func cutKV(arg string, key string) (string, bool) {
	if !strings.HasPrefix(arg, key+"=") {
		return "", false
	}
	return strings.TrimPrefix(arg, key+"="), true
}

func parsePort(s string) (int, error) {
	port, err := strconv.Atoi(s)
	if err != nil || port <= 0 || port > 65535 {
		return 0, fmt.Errorf("invalid --port: %q", s)
	}
	return port, nil
}
