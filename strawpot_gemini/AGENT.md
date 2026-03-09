---
name: strawpot-gemini
description: Gemini CLI agent
metadata:
  strawpot:
    bin:
      macos: strawpot_gemini
      linux: strawpot_gemini
    install:
      macos: curl -fsSL https://raw.githubusercontent.com/strawpot/strawpot_gemini_cli/main/strawpot_gemini/install.sh | sh
      linux: curl -fsSL https://raw.githubusercontent.com/strawpot/strawpot_gemini_cli/main/strawpot_gemini/install.sh | sh
    tools:
      gemini:
        description: Gemini CLI
        install:
          macos: npm install -g @google/gemini-cli
          linux: npm install -g @google/gemini-cli
    params:
      model:
        type: string
        description: Model override (omit to use gemini CLI default)
      dangerously_skip_permissions:
        type: boolean
        default: true
        description: Skip permission prompts (enabled by default, set to false to require approval)
    env:
      GEMINI_API_KEY:
        required: false
        description: Gemini API key (optional if logged in with Google account)
---

# Gemini CLI Agent

Runs Gemini CLI as a subprocess. Supports interactive and non-interactive
modes, custom model selection, and skill-based prompt augmentation.
