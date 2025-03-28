# gcode and gssh

Gcode is a Go implementation of rcode, inspired by [yihong's rcode project](https://github.com/yihong0618/rcode). Thanks to yihong.

## Introduction

**gcode** and **gssh** are tools designed to enhance remote development workflows by integrating SSH connections with local development environments. They allow developers to seamlessly work on remote projects as if they were local, improving productivity and convenience.

## Overview

https://github.com/user-attachments/assets/be516cf7-326b-47d0-b8e5-2b9d0321a0bb

- **GCode**: A command-line tool that allows you to open directories from a remote server in your local IDE (VS Code or Cursor).
- **GSSH**: An enhanced SSH command that sets up a secure communication channel between your local machine and a remote server, enabling advanced features provided by GCode.

## Features

- **Seamless Remote Development**: Edit remote files in your local IDE without manual synchronization.
- **Secure Communication**: All data transfer occurs over SSH tunnels, ensuring security.
- **IDE Integration**: Automatically detects and launches the appropriate editor.
- **Session Management**: Handles multiple sessions with unique identifiers.

## Supported Platforms

GCode and GSSH are designed to work across multiple operating systems, providing flexibility for developers in various environments:

- Windows
- Linux
- macOS

## Usage

### Connecting to a Remote Server

Use GSSH to connect to your remote server:

```bash
gssh your-remote-server
```

GSSH accepts all standard SSH parameters, except `-R` and `-T`.

### Opening a Remote Directory

After connecting with GSSH, use GCode on the remote server to open directories in your local IDE:

```bash
gcode .       # Launches VS Code
gcursor .     # Launches Cursor
gwindsurf .   # Launches Windsurf
gtrae .       # Launches Trae
```

### Opening Remote Directories Locally

You can also use GCode locally to open remote directories directly in your IDE

```bash
gcode hostname remote-dir
```

These commands will open the current directory (`.`) from the remote server in your local IDE.

## How It Works

### GSSH

GSSH is a wrapper around the standard SSH command with added functionality:

1. **Session Management**: Generates a unique session ID and key for each connection.
2. **Secure Tunneling**: Creates an SSH tunnel for inter-process communication (IPC).
3. **Environment Setup**: Sets up necessary environment variables on the remote server.

When you connect to a remote server using `gssh`, it prepares both your local and remote environments for seamless interaction.

### GCode

Once connected via GSSH, GCode allows you to open directories on the remote server directly in your local IDE.

- **Communication**: Utilizes the secure channel established by GSSH to communicate between the remote server and your local machine.
- **IDE Integration**: Automatically launches the appropriate IDE (VS Code or Cursor) on your local machine.
- **Remote Directory Access**: Opens the specified remote directory in your local IDE, enabling you to edit files as if they were local.



## Installation

### unix/linux
1. **Install or Update GCode**:

   To install or update gcode, you should run the install script. To do that, you may either download and run the script manually, or use the following cURL or Wget command:

   ```shell
   curl -o- https://raw.githubusercontent.com/xingty/rcode-go/refs/heads/main/install.sh | bash
   ```

   ```shell
   wget -qO- https://raw.githubusercontent.com/xingty/rcode-go/refs/heads/main/install.sh | bash
   ```


2. **Update PATH**:

   Ensure `GCODE_HOME` is in your `$PATH` to enable the `gcode` command:

   ```bash
   export GCODE_HOME="$HOME/gcode"
   export PATH=$PATH:$GCODE_HOME/bin
   ```

   Add this line to your `~/.bashrc` or `~/.zshrc` to make it persistent.

### Windows
  You can intall gcode automatically via powershell command.
  ```powershell
  irm https://raw.githubusercontent.com/xingty/rcode-go/refs/heads/main/install.ps1 | iex

  ```
  Note: You may need to adjust your PowerShell execution policy first by running:
  ```powershell
  Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
  ```
  or Download [the powershell script](https://raw.githubusercontent.com/xingty/rcode-go/refs/heads/main/install.ps1) and execute it manually to install

## Manual Download

If you prefer to manually download the latest release package, you can do so from the GitHub Releases page:

1. Go to the [Releases](https://github.com/xingty/rcode-go/releases) page of the `rcode-go` repository.
2. Look for the latest release.
3. Download the appropriate package for your platform and architecture from the assets section.

After downloading, you may need to extract the package (if it is a `.zip` or `.tar.gz` file) and manually move the files to your desired location.

Remember to update your `$PATH` environment variable if necessary, to include the path to the `gcode` binary.


### Manual Build

For those who prefer to build the project manually or require more control over the build process, our `makefile` provides various targets to accommodate different build needs.

#### Prerequisites

Ensure you have `go` installed on your system. You can check your Go installation by running:

```bash
go version
```

#### Building the Project

To build the project for all supported platforms and architectures, simply run:

```bash
make all
```

This command will compile the project for Windows, Linux, and macOS (darwin) for the supported architectures (amd64, 386, and arm64).

#### Building for a Specific Platform and Architecture

If you need to build for a specific platform and architecture, you can use:

```bash
make build-one PLATFORM=platform ARCH=arch
```

Replace `platform` and `arch` with your desired platform (`windows`, `linux`, `darwin`) and architecture (`amd64`, `386`, `arm64`). For example, to build for Linux on amd64, use:

```bash
# export CGO_ENABLED=0 disable CGO if you want

make build-one PLATFORM=linux ARCH=amd64
```

#### Cleaning Build Artifacts

To clean up all build artifacts, run:

```bash
make clean
```

This command removes the `dist` directory, which contains the build outputs.


### Advanced Options


- **Custom IPC Host**:

  ```bash
  gssh --host <host> your-remote-server
  ```

- **Custom IPC Port**:

  ```bash
  gssh --port <port> your-remote-server
  ```

## Notes

- **SSH Configuration**:

  Ensure you have SSH keys set up for password-less login if required. Update your `~/.ssh/config` for easier access.

- **Environment Variables**:

  If commands are not recognized, ensure your `$PATH` includes directories where `gcode` and other scripts are installed.


## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).