package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TmuxWindow defines a tmux window structure.
type TmuxWindow struct {
	Name     string   `json:"name"`
	Commands []string `json:"commands"`
}

// TmuxSession defines the structure of a tmux session.
type TmuxSession struct {
	SessionName string       `json:"session_name"`
	Windows     []TmuxWindow `json:"windows"`
}

// ReadConfig reads the .tp.conf file from the home directory.
func ReadConfig() (map[string]string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not find home directory: %v", err)
	}

	configFile := filepath.Join(homeDir, ".tp.conf")
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	config := make(map[string]string)
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue // skip empty lines and comments
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			config[parts[0]] = strings.TrimSpace(parts[1])
		}
	}
	return config, nil
}

// ListScripts searches for .tmux files in the specified directory.
func ListScripts(root string) ([]string, error) {
	var scripts []string

	err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		depth := len(strings.Split(relativePath, string(os.PathSeparator)))

		if depth > 3 {
			if info.IsDir() {
				return filepath.SkipDir // skip current dir in walk
			}
			return nil
		}

		if info.Name() == ".tmux" && !info.IsDir() {
			scripts = append(scripts, path)
		}
		return nil
	})

	return scripts, err
}

// SelectScriptWithFzf allows the user to select a script using fzf.
func SelectScriptWithFzf(scripts []string) (string, error) {
	cmd := exec.Command("fzf")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	go func() {
		defer stdin.Close()
		for _, script := range scripts {
			fmt.Fprintln(stdin, script)
		}
	}()

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// ReadSessionConfig reads the selected tmux session configuration from the .tmux file.
func ReadSessionConfig(path string) (*TmuxSession, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read tmux config file: %v", err)
	}

	var session TmuxSession
	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, fmt.Errorf("could not parse tmux config: %v", err)
	}

	return &session, nil
}

// sessionExists checks if a tmux session with the given name already exists.
func sessionExists(name string) bool {
	cmd := exec.Command("tmux", "has-session", "-t", name)
	err := cmd.Run()
	return err == nil
}

// createSession creates a new tmux session based on the configuration.
func createSession(session *TmuxSession, basePath string) error {
	// Create a new tmux session
	err := exec.Command("tmux", "new-session", "-d", "-s", session.SessionName).Run()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}

	for i, window := range session.Windows {
		targetWindow := fmt.Sprintf("%s:%d", session.SessionName, i)

		if i == 0 {
			// Rename the first window in the session
			err = exec.Command("tmux", "rename-window", "-t", targetWindow, window.Name).Run()
		} else {
			// Create new windows for the session
			err = exec.Command("tmux", "new-window", "-t", session.SessionName, "-n", window.Name).Run()
		}

		if err != nil {
			return fmt.Errorf("failed to create window: %v", err)
		}

		// Change directory in each tmux window to the base path of the selected .tmux file
		err = exec.Command("tmux", "send-keys", "-t", targetWindow, fmt.Sprintf("cd %s", basePath), "C-m").Run()
		if err != nil {
			return fmt.Errorf("failed to change directory in window %d: %v", i, err)
		}

		// Send commands to each window
		for _, cmdStr := range window.Commands {
			err = exec.Command("tmux", "send-keys", "-t", targetWindow, cmdStr, "C-m").Run()
			if err != nil {
				return fmt.Errorf("failed to send command to window %d: %v", i, err)
			}
		}
	}

	return nil
}

// attachSession attaches to an existing tmux session.
func attachSession(name string) error {
	fmt.Printf("Attempting to attach to session: %s\n", name)

	cmd := exec.Command("tmux", "attach-session", "-t", name)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to attach to session: %v", err)
	}

	return nil
}

// initTmuxConfig creates a minimal .tmux file in the current directory.
func initTmuxConfig() error {
	minimalConfig := TmuxSession{
		SessionName: "my-session",
		Windows: []TmuxWindow{
			{
				Name:     "window1",
				Commands: []string{"echo 'Hello, tmux!'"},
			},
		},
	}

	data, err := json.MarshalIndent(minimalConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal minimal config: %v", err)
	}

	err = ioutil.WriteFile(".tmux", data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write .tmux file: %v", err)
	}

	fmt.Println("Created minimal .tmux file in the current directory.")
	return nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		err := initTmuxConfig()
		if err != nil {
			fmt.Println("Error initializing tmux config:", err)
			os.Exit(1)
		}
		return
	}

	// Read the .tp.conf configuration
	config, err := ReadConfig()
	if err != nil {
		fmt.Println("Error reading configuration:", err)
		os.Exit(1)
	}

	root, ok := config["TP_DIRECTORY"]
	if !ok {
		fmt.Println("TP_DIRECTORY is not set in the configuration")
		os.Exit(1)
	}

	stat, err := os.Stat(root)
	if err != nil || !stat.IsDir() {
		fmt.Printf("%s is not a valid directory (%s)\n", "TP_DIRECTORY", root)
		os.Exit(1)
	}

	// List available .tmux files
	scripts, err := ListScripts(root)
	if err != nil {
		fmt.Println("Error listing scripts:", err)
		os.Exit(1)
	}

	if len(scripts) == 0 {
		fmt.Println("No .tmux scripts found.")
		os.Exit(1)
	}

	// Use fzf to select a script
	selectedScript, err := SelectScriptWithFzf(scripts)
	if err != nil {
		fmt.Println("Error selecting script:", err)
		os.Exit(1)
	}

	// Read the tmux session configuration
	sessionConfig, err := ReadSessionConfig(selectedScript)
	if err != nil {
		fmt.Println("Error reading tmux session configuration:", err)
		os.Exit(1)
	}

	// Get the base path of the selected script to use as the directory in tmux windows
	basePath := filepath.Dir(selectedScript)

	// Check if the session already exists
	if sessionExists(sessionConfig.SessionName) {
		fmt.Printf("Session %s already exists. Attaching to it.\n", sessionConfig.SessionName)
		err = attachSession(sessionConfig.SessionName)
		if err != nil {
			fmt.Printf("Error attaching to session: %v\n", err)
		}
	} else {
		// Create the session and attach
		err = createSession(sessionConfig, basePath)
		if err != nil {
			fmt.Printf("Error creating session: %v\n", err)
			os.Exit(1)
		}

		err = attachSession(sessionConfig.SessionName)
		if err != nil {
			fmt.Printf("Error attaching to session: %v\n", err)
			os.Exit(1)
		}
	}
}
