package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBuildArgs_AllFlags(t *testing.T) {
	args := []string{
		"--agent-id", "agent-1",
		"--working-dir", "/work",
		"--agent-workspace-dir", "/workspace",
		"--role-prompt", "You are a helper",
		"--memory-prompt", "Remember this",
		"--task", "do something",
		"--config", `{"model":"gemini-2.5-pro"}`,
		"--skills-dir", "/skills",
		"--roles-dir", "/roles1",
		"--roles-dir", "/roles2",
	}

	ba := parseBuildArgs(args)

	if ba.AgentID != "agent-1" {
		t.Errorf("AgentID = %q, want %q", ba.AgentID, "agent-1")
	}
	if ba.WorkingDir != "/work" {
		t.Errorf("WorkingDir = %q, want %q", ba.WorkingDir, "/work")
	}
	if ba.AgentWorkspaceDir != "/workspace" {
		t.Errorf("AgentWorkspaceDir = %q, want %q", ba.AgentWorkspaceDir, "/workspace")
	}
	if ba.RolePrompt != "You are a helper" {
		t.Errorf("RolePrompt = %q, want %q", ba.RolePrompt, "You are a helper")
	}
	if ba.MemoryPrompt != "Remember this" {
		t.Errorf("MemoryPrompt = %q, want %q", ba.MemoryPrompt, "Remember this")
	}
	if ba.Task != "do something" {
		t.Errorf("Task = %q, want %q", ba.Task, "do something")
	}
	if ba.Config != `{"model":"gemini-2.5-pro"}` {
		t.Errorf("Config = %q, want %q", ba.Config, `{"model":"gemini-2.5-pro"}`)
	}
	if len(ba.SkillsDirs) != 1 || ba.SkillsDirs[0] != "/skills" {
		t.Errorf("SkillsDirs = %v, want [/skills]", ba.SkillsDirs)
	}
	if len(ba.RolesDirs) != 2 || ba.RolesDirs[0] != "/roles1" || ba.RolesDirs[1] != "/roles2" {
		t.Errorf("RolesDirs = %v, want [/roles1 /roles2]", ba.RolesDirs)
	}
}

func TestParseBuildArgs_Empty(t *testing.T) {
	ba := parseBuildArgs([]string{})

	if ba.Config != "{}" {
		t.Errorf("Config default = %q, want %q", ba.Config, "{}")
	}
	if ba.AgentID != "" {
		t.Errorf("AgentID = %q, want empty", ba.AgentID)
	}
	if ba.RolesDirs != nil {
		t.Errorf("RolesDirs = %v, want nil", ba.RolesDirs)
	}
}

func TestParseBuildArgs_TrailingFlag(t *testing.T) {
	// A flag without a value at the end should not panic
	ba := parseBuildArgs([]string{"--agent-id"})

	if ba.AgentID != "" {
		t.Errorf("AgentID = %q, want empty (trailing flag ignored)", ba.AgentID)
	}
}

func TestCmdBuild_MinimalArgs(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "workspace")

	args := []string{
		"--agent-workspace-dir", wsDir,
		"--working-dir", "/project",
	}

	output := captureBuildOutput(t, args)

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	cmd, ok := result["cmd"].([]interface{})
	if !ok {
		t.Fatal("Missing 'cmd' in output")
	}

	// cmd is now ["sh", "-c", "<shell>"]
	if len(cmd) != 3 || cmd[0] != "sh" || cmd[1] != "-c" {
		t.Fatalf("cmd should be [sh, -c, <shell>], got %v", cmd)
	}

	shell := cmd[2].(string)
	assertShellContains(t, shell, "'gemini'")
	assertShellContains(t, shell, "--include-directories")
	assertShellContains(t, shell, wsDir)
	assertShellContains(t, shell, "2>/dev/null")

	if result["cwd"] != "/project" {
		t.Errorf("cwd = %v, want /project", result["cwd"])
	}

	// No -p when no task/prompts provided (minimal args)
	if strings.Contains(shell, "'-p'") {
		t.Error("shell cmd should NOT contain -p when no task or prompts provided")
	}
}

func TestCmdBuild_WithTaskAndModel(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "workspace")

	args := []string{
		"--agent-workspace-dir", wsDir,
		"--working-dir", "/project",
		"--task", "fix the bug",
		"--config", `{"model":"gemini-2.5-pro"}`,
	}

	output := captureBuildOutput(t, args)

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	shell := result["cmd"].([]interface{})[2].(string)
	assertShellContains(t, shell, "'-p' 'fix the bug'")
	assertShellContains(t, shell, "'-m' 'gemini-2.5-pro'")
}

func TestCmdBuild_PromptsInTask(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "workspace")

	args := []string{
		"--agent-workspace-dir", wsDir,
		"--role-prompt", "You are a coder",
		"--memory-prompt", "Use Go",
		"--task", "fix the bug",
	}

	output := captureBuildOutput(t, args)

	var result map[string]interface{}
	json.Unmarshal(output, &result)
	shell := result["cmd"].([]interface{})[2].(string)

	// Role + memory + task should all be combined in -p
	assertShellContains(t, shell, "You are a coder")
	assertShellContains(t, shell, "Use Go")
	assertShellContains(t, shell, "fix the bug")
	assertShellContains(t, shell, "'-p'")
}

func TestCmdBuild_PromptsOnly(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "workspace")

	args := []string{
		"--agent-workspace-dir", wsDir,
		"--role-prompt", "You are a coder",
		"--memory-prompt", "Use Go",
	}

	output := captureBuildOutput(t, args)

	var result map[string]interface{}
	json.Unmarshal(output, &result)
	shell := result["cmd"].([]interface{})[2].(string)

	// Role + memory should be in -p even without task
	assertShellContains(t, shell, "You are a coder")
	assertShellContains(t, shell, "Use Go")
	assertShellContains(t, shell, "'-p'")
}

func TestCmdBuild_SkillsSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "workspace")
	skillsDir := filepath.Join(tmpDir, "skills")

	// Create skill directories and a file (file should be skipped)
	os.MkdirAll(filepath.Join(skillsDir, "skill-a"), 0o755)
	os.MkdirAll(filepath.Join(skillsDir, "skill-b"), 0o755)
	os.WriteFile(filepath.Join(skillsDir, "not-a-dir.txt"), []byte("skip"), 0o644)

	args := []string{
		"--agent-workspace-dir", wsDir,
		"--skills-dir", skillsDir,
	}

	captureBuildOutput(t, args)

	// Verify symlinks created for directories only
	for _, name := range []string{"skill-a", "skill-b"} {
		link := filepath.Join(wsDir, "skills", name)
		target, err := os.Readlink(link)
		if err != nil {
			t.Errorf("Symlink %s not created: %v", name, err)
			continue
		}
		expected := filepath.Join(skillsDir, name)
		if target != expected {
			t.Errorf("Symlink %s -> %q, want %q", name, target, expected)
		}
	}

	// File should not be symlinked
	notADir := filepath.Join(wsDir, "skills", "not-a-dir.txt")
	if _, err := os.Lstat(notADir); err == nil {
		t.Error("non-directory 'not-a-dir.txt' should not be symlinked")
	}
}

func TestCmdBuild_RolesSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "workspace")
	rolesDir1 := filepath.Join(tmpDir, "roles1")
	rolesDir2 := filepath.Join(tmpDir, "roles2")

	os.MkdirAll(filepath.Join(rolesDir1, "role-a"), 0o755)
	os.MkdirAll(filepath.Join(rolesDir2, "role-a"), 0o755) // duplicate — should be skipped
	os.MkdirAll(filepath.Join(rolesDir2, "role-b"), 0o755)

	args := []string{
		"--agent-workspace-dir", wsDir,
		"--roles-dir", rolesDir1,
		"--roles-dir", rolesDir2,
	}

	captureBuildOutput(t, args)

	// role-a should point to rolesDir1 (first wins)
	target, err := os.Readlink(filepath.Join(wsDir, "roles", "role-a"))
	if err != nil {
		t.Fatalf("role-a symlink not created: %v", err)
	}
	if target != filepath.Join(rolesDir1, "role-a") {
		t.Errorf("role-a -> %q, want first roles-dir", target)
	}

	// role-b should point to rolesDir2
	target, err = os.Readlink(filepath.Join(wsDir, "roles", "role-b"))
	if err != nil {
		t.Fatalf("role-b symlink not created: %v", err)
	}
	if target != filepath.Join(rolesDir2, "role-b") {
		t.Errorf("role-b -> %q, want %q", target, filepath.Join(rolesDir2, "role-b"))
	}
}

func TestCmdBuild_SandboxMode(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "workspace")

	t.Setenv("SANDBOX_MODE", "docker")

	args := []string{
		"--agent-workspace-dir", wsDir,
	}

	output := captureBuildOutput(t, args)

	var result map[string]interface{}
	json.Unmarshal(output, &result)
	shell := result["cmd"].([]interface{})[2].(string)
	assertShellContains(t, shell, "'--sandbox'")
}

func TestCmdBuild_Yolo_Default(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "workspace")

	args := []string{
		"--agent-workspace-dir", wsDir,
	}

	output := captureBuildOutput(t, args)

	var result map[string]interface{}
	json.Unmarshal(output, &result)
	shell := result["cmd"].([]interface{})[2].(string)
	assertShellContains(t, shell, "'--yolo'")
}

func TestCmdBuild_Yolo_ExplicitTrue(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "workspace")

	args := []string{
		"--agent-workspace-dir", wsDir,
		"--config", `{"dangerously_skip_permissions":true}`,
	}

	output := captureBuildOutput(t, args)

	var result map[string]interface{}
	json.Unmarshal(output, &result)
	shell := result["cmd"].([]interface{})[2].(string)
	assertShellContains(t, shell, "'--yolo'")
}

func TestCmdBuild_Yolo_Disabled(t *testing.T) {
	tmpDir := t.TempDir()
	wsDir := filepath.Join(tmpDir, "workspace")

	args := []string{
		"--agent-workspace-dir", wsDir,
		"--config", `{"dangerously_skip_permissions":false}`,
	}

	output := captureBuildOutput(t, args)

	var result map[string]interface{}
	json.Unmarshal(output, &result)
	shell := result["cmd"].([]interface{})[2].(string)
	if strings.Contains(shell, "--yolo") {
		t.Error("shell cmd should NOT contain --yolo when disabled")
	}
}

// --- helpers ---

// captureBuildOutput runs cmdBuild and captures its stdout JSON output.
func captureBuildOutput(t *testing.T, args []string) []byte {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	cmdBuild(args)

	w.Close()
	os.Stdout = oldStdout

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	r.Close()

	return buf[:n]
}

func assertShellContains(t *testing.T, shell, substr string) {
	t.Helper()
	if !strings.Contains(shell, substr) {
		t.Errorf("shell cmd %q does not contain %q", shell, substr)
	}
}
