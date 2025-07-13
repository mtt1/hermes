# Hermes

Hermes translates natural language into shell commands. Say what you want, get the command ready to run.

## Setup

1. Get a Gemini API key from Google AI Studio
2. Build: `go build -o hermes cmd/hermes/main.go`
3. Add to your shell:
   
   **Zsh:** `echo 'eval "$(./hermes init zsh)"' >> ~/.zshrc && source ~/.zshrc`
   
   **Bash:** `echo 'eval "$(./hermes init bash)"' >> ~/.bashrc && source ~/.bashrc`
   
   **Fish:** `mkdir -p ~/.config/fish/functions && ./hermes init fish > ~/.config/fish/functions/hermes.fish`

4. Set your API key:
   - Environment variable: `export GEMINI_API_KEY=your_key_here`
   - CLI flag: `--gemini-api-key your_key_here`
   - Config file: `~/.config/hermes/config.toml`

## Usage

```bash
hermes gen list all files
hermes gen find python files
hermes generate "delete old logs"
# All generate the appropriate commands

hermes gen --verbose find python files
hermes gen -v delete old logs
# Shows explanation + generates command

hermes exp ls
hermes exp -- ls -la
hermes explain "grep -r pattern ."
# Explains what commands do
```

The generated command appears in your shell buffer. Review it before pressing enter.

Dangerous commands show warnings. You always have final control.

## Commands

- `hermes [gen|generate] <description>` - Generate a command
- `hermes [gen|generate] --verbose/-v <description>` - Generate command with detailed explanation
- `hermes [exp|explain] <command>` - Explain what a command does (quotes or `--` for complex descriptions)
- `hermes init [zsh|bash|fish]` - Print shell integration code
- `hermes --help` - Show help
- `hermes --version` - Show version
