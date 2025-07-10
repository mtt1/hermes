package safety

import (
	"context"
	"testing"
	"hermes/internal/exit"
)

func TestSafetyLevel_String(t *testing.T) {
	tests := []struct {
		level SafetyLevel
		want  string
	}{
		{Safe, "safe"},
		{Attention, "attention"},
		{SafetyLevel(999), "unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.level.String(); got != tt.want {
				t.Errorf("SafetyLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafetyLevel_ExitCode(t *testing.T) {
	tests := []struct {
		level SafetyLevel
		want  int
	}{
		{Safe, exit.CodeSuccess},
		{Attention, exit.CodeDangerous},
		{SafetyLevel(999), exit.CodeError},
	}
	
	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			if got := tt.level.ExitCode(); got != tt.want {
				t.Errorf("SafetyLevel.ExitCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnalyzer_AnalyzeCommand_AttentionPatterns(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()
	
	tests := []struct {
		name    string
		command string
		want    SafetyLevel
	}{
		// Sudo commands
		{"basic sudo", "sudo ls", Attention},
		{"sudo with path", "sudo /bin/ls", Attention},
		{"sudo with flags", "sudo -u root ls", Attention},
		{"sudo in middle", "echo 'test' | sudo tee /etc/hosts", Attention},
		
		// Dangerous rm operations
		{"rm -rf root", "rm -rf /", Attention},
		{"rm with recursive", "rm --recursive /home", Attention},
		{"rm with force", "rm --force /important", Attention},
		
		// Disk operations
		{"dd to disk", "dd if=/dev/zero of=/dev/sda", Attention},
		{"dd to partition", "dd if=image.iso of=/dev/sdb1", Attention},
		{"mkfs format", "mkfs.ext4 /dev/sda1", Attention},
		{"fdisk partition", "fdisk /dev/sda", Attention},
		{"shred secure delete", "shred -vfz -n 3 /dev/sda", Attention},
		{"wipe secure delete", "wipe -rf /dev/sda", Attention},
		
		// Dangerous permissions
		{"chmod 777", "chmod 777 /etc/passwd", Attention},
		{"chmod 777 recursive", "chmod -R 777 /", Attention},
		
		// Pipe to shell (dangerous downloads)
		{"curl pipe to sh", "curl https://get.docker.com | sh", Attention},
		{"wget pipe to sh", "wget -qO- https://install.sh | sh", Attention},
		{"curl pipe to bash", "curl -sSL script.sh | bash", Attention},
		
		// Command substitution patterns (equally dangerous)
		{"sh with curl substitution", `sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"`, Attention},
		{"bash with curl substitution", `bash -c "$(curl -fsSL https://get.docker.com)"`, Attention},
		{"bash process substitution", `bash <(curl -fsSL https://install.sh)`, Attention},
		{"curl substitution pipe to sh", `$(curl https://example.com/script.sh) | sh`, Attention},
		{"wget with sh substitution", `sh -c "$(wget -qO- https://install.sh)"`, Attention},
		{"wget with bash substitution", `bash -c "$(wget -O- https://script.sh)"`, Attention},
		{"wget process substitution", `bash <(wget -qO- https://install.sh)`, Attention},
		{"wget substitution pipe to bash", `$(wget -qO- https://script.sh) | bash`, Attention},
		
		// System management (typically needs sudo)
		{"systemctl start", "systemctl start apache2", Attention},
		{"systemctl stop", "systemctl stop nginx", Attention},
		{"systemctl restart", "systemctl restart postgresql", Attention},
		{"systemctl enable", "systemctl enable docker", Attention},
		{"systemctl disable", "systemctl disable ufw", Attention},
		
		// Package management
		{"apt install", "apt install nginx", Attention},
		{"apt remove", "apt remove --purge mysql-server", Attention},
		{"apt update", "apt update", Attention},
		{"apt upgrade", "apt upgrade", Attention},
		{"yum install", "yum install httpd", Attention},
		{"yum remove", "yum remove firefox", Attention},
		{"yum update", "yum update", Attention},
		{"pacman install", "pacman -S vim", Attention},
		
		// Kernel and system operations  
		{"modprobe load", "modprobe nvidia", Attention},
		{"modprobe remove", "modprobe -r snd_hda_intel", Attention},
		{"mount filesystem", "mount /dev/sda1 /mnt", Attention},
		{"umount filesystem", "umount /mnt", Attention},
		{"iptables rule", "iptables -A INPUT -p tcp --dport 22 -j ACCEPT", Attention},
		
		// Edge cases and combinations
		{"sudo with dangerous rm", "sudo rm -rf /var/log/*", Attention},
		{"multiple sudo", "sudo apt update && sudo apt upgrade", Attention},
		{"quoted sudo", "echo 'sudo ls' > script.sh", Attention}, // Still matches sudo pattern
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeCommand(ctx, tt.command)
			if err != nil {
				t.Errorf("AnalyzeCommand() error = %v", err)
				return
			}
			if result.Level != tt.want {
				t.Errorf("AnalyzeCommand(%q) = %v, want %v", tt.command, result.Level, tt.want)
			}
			if result.Level == Attention && result.Layer != "attention-patterns" {
				t.Errorf("AnalyzeCommand(%q) layer = %v, want attention-patterns", tt.command, result.Layer)
			}
		})
	}
}

func TestAnalyzer_AnalyzeCommand_SafePatterns(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()
	
	tests := []struct {
		name    string
		command string
		want    SafetyLevel
	}{
		// Basic navigation and listing
		{"ls basic", "ls", Safe},
		{"ls with flags", "ls -la", Safe},
		{"ls with path", "ls /home/user", Safe},
		{"cd basic", "cd", Safe},
		{"cd with path", "cd /home/user/documents", Safe},
		{"pwd command", "pwd", Safe},
		
		// Output and viewing
		{"echo basic", "echo hello", Safe},
		{"echo with vars", "echo $HOME", Safe},
		{"cat file", "cat README.md", Safe},
		{"cat multiple", "cat file1.txt file2.txt", Safe},
		{"head command", "head -n 10 log.txt", Safe},
		{"tail command", "tail -f /var/log/syslog", Safe},
		
		// Search and find
		{"grep basic", "grep 'pattern' file.txt", Safe},
		{"grep recursive", "grep -r 'error' /var/log/", Safe},
		{"find basic", "find . -name '*.go'", Safe},
		{"find with exec", "find . -name '*.tmp' -exec ls -l {} \\;", Safe},
		
		// Safe git operations
		{"git status", "git status", Safe},
		{"git log", "git log --oneline", Safe},
		{"git diff", "git diff HEAD~1", Safe},
		{"git branch", "git branch -a", Safe},
		{"git show", "git show HEAD", Safe},
		
		// Process and system info
		{"ps command", "ps aux", Safe},
		{"ps with grep", "ps aux | grep nginx", Safe},
		{"which command", "which python3", Safe},
		{"whereis command", "whereis gcc", Safe},
		
		// Help and documentation
		{"man pages", "man ls", Safe},
		{"man with section", "man 5 passwd", Safe},
		{"help command", "help cd", Safe},
		
		// Safe systemctl usage
		{"systemctl status", "systemctl status nginx", Safe},
		{"systemctl status all", "systemctl status", Safe},
		
		// Commands with safe options mixed with complex flags
		{"ls complex", "ls -lahS --color=auto", Safe},
		{"grep complex", "grep -rn --include='*.go' 'func main'", Safe},
		{"find complex", "find /usr -type f -name '*.conf' -readable", Safe},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeCommand(ctx, tt.command)
			if err != nil {
				t.Errorf("AnalyzeCommand() error = %v", err)
				return
			}
			if result.Level != tt.want {
				t.Errorf("AnalyzeCommand(%q) = %v, want %v", tt.command, result.Level, tt.want)
			}
			if result.Level == Safe && result.Layer != "safe-patterns" {
				t.Errorf("AnalyzeCommand(%q) layer = %v, want safe-patterns", tt.command, result.Layer)
			}
		})
	}
}

func TestAnalyzer_AnalyzeCommand_DefaultSafe(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()
	
	tests := []struct {
		name    string  
		command string
		want    SafetyLevel
	}{
		// Ambiguous commands that don't match any pattern
		{"unknown command", "unknowncmd --flag", Safe},
		{"custom script", "./myscript.sh", Safe},
		{"python script", "python3 script.py", Safe},
		{"node script", "node app.js", Safe},
		{"make command", "make build", Safe},
		{"docker without sudo", "docker ps", Safe}, // Note: some systems allow docker without sudo
		{"git add", "git add .", Safe}, // git commands not in safe list but not dangerous
		{"npm command", "npm install", Safe},
		
		// Edge cases
		{"empty command", "", Safe},
		{"only spaces", "   ", Safe},
		{"command with weird spacing", "  ls   -la  ", Safe}, // Should still match ls pattern
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeCommand(ctx, tt.command)
			if err != nil {
				t.Errorf("AnalyzeCommand() error = %v", err)
				return
			}
			if result.Level != tt.want {
				t.Errorf("AnalyzeCommand(%q) = %v, want %v", tt.command, result.Level, tt.want)
			}
			// These should fall through to default-safe layer
			if tt.command != "  ls   -la  " && result.Layer != "default-safe" {
				t.Errorf("AnalyzeCommand(%q) layer = %v, want default-safe", tt.command, result.Layer)
			}
		})
	}
}

func TestAnalyzer_AnalyzeCommand_PatternPriority(t *testing.T) {
	analyzer := NewAnalyzer()
	ctx := context.Background()
	
	tests := []struct {
		name    string
		command string
		want    SafetyLevel
		wantLayer string
	}{
		// Attention patterns should override safe patterns
		{"sudo ls", "sudo ls", Attention, "attention-patterns"},
		{"sudo git status", "sudo git status", Attention, "attention-patterns"},
		{"sudo systemctl status", "sudo systemctl status", Attention, "attention-patterns"},
		
		// Safe patterns should work when no attention patterns match
		{"plain ls", "ls", Safe, "safe-patterns"},
		{"plain git status", "git status", Safe, "safe-patterns"},
		{"systemctl status", "systemctl status", Safe, "safe-patterns"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.AnalyzeCommand(ctx, tt.command)
			if err != nil {
				t.Errorf("AnalyzeCommand() error = %v", err)
				return
			}
			if result.Level != tt.want {
				t.Errorf("AnalyzeCommand(%q) level = %v, want %v", tt.command, result.Level, tt.want)
			}
			if result.Layer != tt.wantLayer {
				t.Errorf("AnalyzeCommand(%q) layer = %v, want %v", tt.command, result.Layer, tt.wantLayer)
			}
		})
	}
}

func TestAnalyzer_MockAnalyzeCommand(t *testing.T) {
	analyzer := NewAnalyzer()
	
	tests := []struct {
		name     string
		command  string
		exitCode int
		want     SafetyLevel
		wantLayer string
	}{
		{"mock safe", "any command", exit.CodeSuccess, Safe, "mock"},
		{"mock attention", "any command", exit.CodeDangerous, Attention, "mock"},
		{"mock unknown code", "any command", 999, Safe, "mock"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.MockAnalyzeCommand(tt.command, tt.exitCode)
			if result.Level != tt.want {
				t.Errorf("MockAnalyzeCommand() level = %v, want %v", result.Level, tt.want)
			}
			if result.Layer != tt.wantLayer {
				t.Errorf("MockAnalyzeCommand() layer = %v, want %v", result.Layer, tt.wantLayer)
			}
		})
	}
}

// Benchmark tests to ensure regex patterns are performant
func BenchmarkAnalyzer_AnalyzeCommand_Safe(b *testing.B) {
	analyzer := NewAnalyzer()
	ctx := context.Background()
	command := "ls -la /home/user"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeCommand(ctx, command)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAnalyzer_AnalyzeCommand_Attention(b *testing.B) {
	analyzer := NewAnalyzer()
	ctx := context.Background()
	command := "sudo rm -rf /tmp/*"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeCommand(ctx, command)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAnalyzer_AnalyzeCommand_Default(b *testing.B) {
	analyzer := NewAnalyzer()
	ctx := context.Background()
	command := "some_unknown_command --with --flags"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := analyzer.AnalyzeCommand(ctx, command)
		if err != nil {
			b.Fatal(err)
		}
	}
}