# gcode and gssh

## Introduction

Gcode is a Go implementation of rcode, inspired by yihong's rcode project. Thanks to yihong.

**gcode** and **gssh** are tools designed to enhance remote development workflows by integrating SSH connections with local development environments. They allow developers to seamlessly work on remote projects as if they were local, improving productivity and convenience.

## Overview

- **GCode**: A command-line tool that allows you to open directories from a remote server in your local IDE (VS Code or Cursor).
- **GSSH**: An enhanced SSH command that sets up a secure communication channel between your local machine and a remote server, enabling advanced features provided by GCode.

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

1. **Install GCode**:

   Download `gcode` from the latest release

2. **Update PATH**:

   Ensure `GCODE_HOME` is in your `$PATH` to enable the `gcode` command:

   ```bash
   export GCODE_HOME="change to your gcode path"
   export PATH=$PATH:$GCODE_HOME/bin
   ```

   Add this line to your `~/.bashrc` or `~/.zshrc` to make it persistent.

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
```

These commands will open the current directory (`.`) from the remote server in your local IDE.

### Advanced Options

- **Custom IPC Host**:

  ```bash
  gssh --host <host> your-remote-server
  ```

- **Custom IPC Port**:

  ```bash
  gssh --port <port> your-remote-server
  ```

## Features

- **Seamless Remote Development**: Edit remote files in your local IDE without manual synchronization.
- **Secure Communication**: All data transfer occurs over SSH tunnels, ensuring security.
- **IDE Integration**: Automatically detects and launches the appropriate editor.
- **Session Management**: Handles multiple sessions with unique identifiers.


## Notes

- **SSH Configuration**:

  Ensure you have SSH keys set up for password-less login if required. Update your `~/.ssh/config` for easier access.

- **Environment Variables**:

  If commands are not recognized, ensure your `$PATH` includes directories where `gcode` and other scripts are installed.


## Contributing

Contributions are welcome! Please fork the repository and submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).