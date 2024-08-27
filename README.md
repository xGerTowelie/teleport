# Tmux Session Manager

This Go program is designed to streamline the process of initializing and managing `tmux` sessions. It allows you to define your tmux session structure using JSON configuration files, and it provides a mechanism to search for and select these configurations using `fzf`. Additionally, the program supports general configuration via a `.tp.conf` file located in your home directory.

## Todos
- [ ] fix bug: open in terminal spawns new subprocess

## Features

- Automatically initialize `tmux` sessions based on predefined configurations.
- Easily switch between different tmux session configurations using `fzf`.
- Supports custom configurations for specifying script directories and other settings.
- Automatically attaches to an existing session if it is already running.

## Dependencies

To use this program, you need to have the following tools installed:

- **Go**: The program is written in Go, so you need Go installed to compile and run the program.
- **tmux**: A terminal multiplexer that allows you to manage multiple terminal sessions from a single window.
- **fzf**: A command-line fuzzy finder that allows you to quickly search and select from a list of items.

### Installation

To install the dependencies:

#### On Ubuntu/Debian-based systems:

```bash
sudo apt-get install tmux fzf
```

#### On macOS (using Homebrew):

```bash
brew install tmux fzf
```

#### On Arch Linux:

```bash
sudo pacman -S tmux fzf
```

## Configuration

The program can be customized using two types of configuration files:

1. **General Configuration (`.tp.conf`)**: A file located in your home directory that provides general settings for the program.
2. **Tmux Session Configuration (`.tmux`)**: JSON files located in project directories that define how specific tmux sessions should be built.

### 1. General Configuration (`.tp.conf`)

This file should be placed in your home directory (`~/.tp.conf`). It uses a simple key-value format. Below is an example configuration:

```ini
# .tp.conf
TP_DIRECTORY=/path/to/your/scripts
```

- **`TP_DIRECTORY`**: The directory where the program will search for `.tmux` configuration files. This is mandatory.

### 2. Tmux Session Configuration (`.tmux`)

Each `.tmux` file should be a JSON file located in a project directory. This file defines the tmux session, including the session name, windows, and the commands to run in each window. Here’s an example of what a `.tmux` file might look like:

```json
{
    "session_name": "my_project",
    "windows": [
        {
            "name": "Editor",
            "commands": [
                "cd ~/projects/my_project",
                "vim"
            ]
        },
        {
            "name": "Server",
            "commands": [
                "cd ~/projects/my_project",
                "npm start"
            ]
        },
        {
            "name": "Database",
            "commands": [
                "cd ~/projects/my_project",
                "docker-compose up"
            ]
        }
    ]
}
```

- **`session_name`**: The name of the tmux session.
- **`windows`**: A list of windows to be created within the tmux session.
  - **`name`**: The name of the window.
  - **`commands`**: A list of commands to execute in the window.

## Usage

1. **Set up the `.tp.conf` file** in your home directory to point to the directory where your `.tmux` configuration files are located.

2. **Create your tmux session configurations** as JSON files named `.tmux` in your project directories.

3. **Run the Go program**:

   ```bash
   go run tmux_initializer.go
   ```

4. The program will:
   - List all `.tmux` configuration files found in the directory specified by `TP_DIRECTORY`.
   - Allow you to select one using `fzf`.
   - Initialize the tmux session based on the selected configuration.

5. If the session already exists, the program will simply attach to it.

