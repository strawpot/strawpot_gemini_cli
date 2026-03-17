// Gemini CLI wrapper — translates StrawPot protocol to Gemini CLI.
//
// This wrapper is a pure translation layer: it maps StrawPot protocol args
// to "gemini" CLI flags.  It does NOT manage processes, sessions, or any
// infrastructure — that is handled by WrapperRuntime in StrawPot core.
//
// Subcommands: setup, build
package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: wrapper <setup|build> [args...]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "setup":
		cmdSetup()
	case "build":
		cmdBuild(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", os.Args[1])
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// setup
// ---------------------------------------------------------------------------

func cmdSetup() {
	geminiPath, err := exec.LookPath("gemini")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: gemini CLI not found on PATH.")
		fmt.Fprintln(os.Stderr, "Install it with: npm install -g @google/gemini-cli")
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Starting Gemini CLI login...")
	fmt.Fprintln(os.Stderr, "If a browser window does not open, copy the URL from the output below.")
	fmt.Fprintln(os.Stderr)

	cmd := exec.Command(geminiPath, "auth", "login")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Inherit full environment so DISPLAY/WAYLAND_DISPLAY are available
	// for browser opening on Linux.
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
}

// ---------------------------------------------------------------------------
// build
// ---------------------------------------------------------------------------

type buildArgs struct {
	AgentID           string
	WorkingDir        string
	AgentWorkspaceDir string
	RolePrompt        string
	MemoryPrompt      string
	Task              string
	Config            string
	SkillsDirs        []string
	RolesDirs         []string
	FilesDirs         []string
}

func parseBuildArgs(args []string) buildArgs {
	var ba buildArgs
	ba.Config = "{}"

	for i := 0; i < len(args); i++ {
		if i+1 >= len(args) {
			break
		}
		switch args[i] {
		case "--agent-id":
			i++
			ba.AgentID = args[i]
		case "--working-dir":
			i++
			ba.WorkingDir = args[i]
		case "--agent-workspace-dir":
			i++
			ba.AgentWorkspaceDir = args[i]
		case "--role-prompt":
			i++
			ba.RolePrompt = args[i]
		case "--memory-prompt":
			i++
			ba.MemoryPrompt = args[i]
		case "--task":
			i++
			ba.Task = args[i]
		case "--config":
			i++
			ba.Config = args[i]
		case "--skills-dir":
			i++
			ba.SkillsDirs = append(ba.SkillsDirs, args[i])
		case "--roles-dir":
			i++
			ba.RolesDirs = append(ba.RolesDirs, args[i])
		case "--files-dir":
			i++
			ba.FilesDirs = append(ba.FilesDirs, args[i])
		}
	}
	return ba
}

// symlink creates a symlink from dst pointing to src.
func symlink(src, dst string) error {
	return os.Symlink(src, dst)
}

func cmdBuild(args []string) {
	ba := parseBuildArgs(args)

	// Parse config JSON
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(ba.Config), &config); err != nil {
		config = map[string]interface{}{}
	}

	// Validate required args
	if ba.AgentWorkspaceDir == "" {
		fmt.Fprintln(os.Stderr, "Error: --agent-workspace-dir is required")
		os.Exit(1)
	}

	// Use agent workspace dir directly as the --include-directories dir for Gemini.
	if err := os.MkdirAll(ba.AgentWorkspaceDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create workspace dir: %v\n", err)
		os.Exit(1)
	}

	// Write prompt file (GEMINI.md) into workspace.
	// Gemini CLI automatically discovers GEMINI.md as context/system instructions.
	promptFile := filepath.Join(ba.AgentWorkspaceDir, "GEMINI.md")
	var parts []string
	if ba.RolePrompt != "" {
		parts = append(parts, ba.RolePrompt)
	}
	if ba.MemoryPrompt != "" {
		parts = append(parts, ba.MemoryPrompt)
	}
	if err := os.WriteFile(promptFile, []byte(strings.Join(parts, "\n\n")), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write prompt file: %v\n", err)
		os.Exit(1)
	}

	// Symlink each subdirectory from each skills-dir into skills/<name>/
	for _, skillsDir := range ba.SkillsDirs {
		if skillsDir == "" {
			continue
		}
		entries, err := os.ReadDir(skillsDir)
		if err == nil && len(entries) > 0 {
			skillsTarget := filepath.Join(ba.AgentWorkspaceDir, "skills")
			if err := os.MkdirAll(skillsTarget, 0o755); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to create skills dir: %v\n", err)
				os.Exit(1)
			}
			for _, entry := range entries {
				if !entry.IsDir() && entry.Type()&fs.ModeSymlink == 0 {
					continue
				}
				src := filepath.Join(skillsDir, entry.Name())
				link := filepath.Join(skillsTarget, entry.Name())
				if _, err := os.Lstat(link); err == nil {
					continue
				}
				if err := symlink(src, link); err != nil {
					fmt.Fprintf(os.Stderr, "Failed to link skill %s: %v\n", entry.Name(), err)
					os.Exit(1)
				}
			}
		}
	}

	// Symlink each subdirectory from each roles-dir into roles/<name>/
	for _, rolesDir := range ba.RolesDirs {
		if rolesDir == "" {
			continue
		}
		entries, err := os.ReadDir(rolesDir)
		if err != nil || len(entries) == 0 {
			continue
		}
		rolesTarget := filepath.Join(ba.AgentWorkspaceDir, "roles")
		if err := os.MkdirAll(rolesTarget, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create roles dir: %v\n", err)
			os.Exit(1)
		}
		for _, entry := range entries {
			if !entry.IsDir() && entry.Type()&fs.ModeSymlink == 0 {
				continue
			}
			src := filepath.Join(rolesDir, entry.Name())
			link := filepath.Join(rolesTarget, entry.Name())
			// Skip if already exists (e.g. pre-placed by another roles-dir)
			if _, err := os.Lstat(link); err == nil {
				continue
			}
			if err := symlink(src, link); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to link role %s: %v\n", entry.Name(), err)
				os.Exit(1)
			}
		}
	}

	// Build gemini command
	cmd := []string{"gemini"}

	if ba.Task != "" {
		cmd = append(cmd, "-p", ba.Task)
	}

	if model, ok := config["model"].(string); ok && model != "" {
		cmd = append(cmd, "-m", model)
	}

	if sm := os.Getenv("SANDBOX_MODE"); sm != "" {
		cmd = append(cmd, "--sandbox")
	}

	// Default: enable --yolo (auto-approve all tool calls) unless explicitly disabled.
	if skip, ok := config["dangerously_skip_permissions"].(bool); !ok || skip {
		cmd = append(cmd, "--yolo")
	}

	// Include agent workspace as additional directory
	cmd = append(cmd, "--include-directories", ba.AgentWorkspaceDir)

	// Add project files directories if provided
	for _, filesDir := range ba.FilesDirs {
		if filesDir != "" {
			cmd = append(cmd, "--include-directories", filesDir)
		}
	}

	// Output JSON
	result := map[string]interface{}{
		"cmd": cmd,
		"cwd": ba.WorkingDir,
	}

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", err)
		os.Exit(1)
	}
}
