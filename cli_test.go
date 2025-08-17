package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIHandler_parseFlags(t *testing.T) {
	handler := NewCLIHandler()

	tests := []struct {
		name        string
		args        []string
		expectCIDR  string
		expectFile  string
		expectHTML  bool
		expectHelp  bool
		expectError bool
	}{
		{
			name:       "basic CIDR input",
			args:       []string{"cidr-calc", "192.168.1.0/24"},
			expectCIDR: "192.168.1.0/24",
		},
		{
			name:       "output file flag short",
			args:       []string{"cidr-calc", "-o", "output.txt", "10.0.0.0/8"},
			expectCIDR: "10.0.0.0/8",
			expectFile: "output.txt",
		},
		{
			name:       "output file flag long",
			args:       []string{"cidr-calc", "--output", "report.txt", "172.16.0.0/16"},
			expectCIDR: "172.16.0.0/16",
			expectFile: "report.txt",
		},
		{
			name:       "HTML flag short",
			args:       []string{"cidr-calc", "-h", "192.168.0.0/16"},
			expectCIDR: "192.168.0.0/16",
			expectHTML: true,
		},
		{
			name:       "HTML flag long",
			args:       []string{"cidr-calc", "--html", "10.10.0.0/24"},
			expectCIDR: "10.10.0.0/24",
			expectHTML: true,
		},
		{
			name:       "combined flags",
			args:       []string{"cidr-calc", "-h", "-o", "network.html", "172.21.4.0/26"},
			expectCIDR: "172.21.4.0/26",
			expectFile: "network.html",
			expectHTML: true,
		},
		{
			name:       "help flag",
			args:       []string{"cidr-calc", "--help"},
			expectHelp: true,
		},
		{
			name:        "HTML output with non-HTML file extension",
			args:        []string{"cidr-calc", "--html", "-o", "output.txt", "192.168.1.0/24"},
			expectError: true,
		},
		{
			name:        "HTML file extension without HTML flag",
			args:        []string{"cidr-calc", "-o", "output.html", "192.168.1.0/24"},
			expectError: true,
		},
		{
			name:        "invalid flag",
			args:        []string{"cidr-calc", "--invalid", "192.168.1.0/24"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := handler.parseFlags(tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if config.CIDR != tt.expectCIDR {
				t.Errorf("expected CIDR %q, got %q", tt.expectCIDR, config.CIDR)
			}

			if config.OutputFile != tt.expectFile {
				t.Errorf("expected output file %q, got %q", tt.expectFile, config.OutputFile)
			}

			if config.HTMLOutput != tt.expectHTML {
				t.Errorf("expected HTML output %v, got %v", tt.expectHTML, config.HTMLOutput)
			}

			if config.ShowHelp != tt.expectHelp {
				t.Errorf("expected show help %v, got %v", tt.expectHelp, config.ShowHelp)
			}
		})
	}
}

func TestCLIHandler_validateConfig(t *testing.T) {
	handler := NewCLIHandler()

	tests := []struct {
		name        string
		config      *Config
		expectError bool
	}{
		{
			name: "valid text output",
			config: &Config{
				CIDR:       "192.168.1.0/24",
				OutputFile: "output.txt",
				HTMLOutput: false,
			},
			expectError: false,
		},
		{
			name: "valid HTML output",
			config: &Config{
				CIDR:       "192.168.1.0/24",
				OutputFile: "output.html",
				HTMLOutput: true,
			},
			expectError: false,
		},
		{
			name: "HTML output with .htm extension",
			config: &Config{
				CIDR:       "192.168.1.0/24",
				OutputFile: "output.htm",
				HTMLOutput: true,
			},
			expectError: false,
		},
		{
			name: "HTML flag with non-HTML extension",
			config: &Config{
				CIDR:       "192.168.1.0/24",
				OutputFile: "output.txt",
				HTMLOutput: true,
			},
			expectError: true,
		},
		{
			name: "HTML extension without HTML flag",
			config: &Config{
				CIDR:       "192.168.1.0/24",
				OutputFile: "output.html",
				HTMLOutput: false,
			},
			expectError: true,
		},
		{
			name: "no output file",
			config: &Config{
				CIDR:       "192.168.1.0/24",
				OutputFile: "",
				HTMLOutput: false,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateConfig(tt.config)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCLIHandler_Run_Integration(t *testing.T) {
	handler := NewCLIHandler()

	tests := []struct {
		name        string
		args        []string
		expectError bool
		checkOutput bool
	}{
		{
			name:        "valid CIDR console output",
			args:        []string{"cidr-calc", "192.168.1.0/24"},
			expectError: false,
			checkOutput: true,
		},
		{
			name:        "help command",
			args:        []string{"cidr-calc", "--help"},
			expectError: false,
		},
		{
			name:        "invalid CIDR",
			args:        []string{"cidr-calc", "invalid-cidr"},
			expectError: true,
		},
		{
			name:        "no arguments",
			args:        []string{"cidr-calc"},
			expectError: true,
		},
		{
			name:        "edge case /32 network",
			args:        []string{"cidr-calc", "192.168.1.1/32"},
			expectError: false,
			checkOutput: true,
		},
		{
			name:        "edge case /31 network",
			args:        []string{"cidr-calc", "192.168.1.0/31"},
			expectError: false,
			checkOutput: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Run(tt.args)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestCLIHandler_FileOutput_Integration(t *testing.T) {
	handler := NewCLIHandler()

	// Create temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		args        []string
		expectError bool
		checkFile   bool
		isHTML      bool
	}{
		{
			name:        "text file output",
			args:        []string{"cidr-calc", "-o", filepath.Join(tempDir, "output.txt"), "192.168.1.0/24"},
			expectError: false,
			checkFile:   true,
			isHTML:      false,
		},
		{
			name:        "HTML file output",
			args:        []string{"cidr-calc", "--html", "-o", filepath.Join(tempDir, "output.html"), "10.0.0.0/8"},
			expectError: false,
			checkFile:   true,
			isHTML:      true,
		},
		{
			name:        "invalid file path",
			args:        []string{"cidr-calc", "-o", "/invalid/path/output.txt", "192.168.1.0/24"},
			expectError: true,
			checkFile:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Run(tt.args)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.checkFile && !tt.expectError {
				// Extract filename from args
				var filename string
				for i, arg := range tt.args {
					if arg == "-o" || arg == "--output" {
						if i+1 < len(tt.args) {
							filename = tt.args[i+1]
							break
						}
					}
				}

				if filename == "" {
					t.Errorf("could not extract filename from args")
					return
				}

				// Check if file exists
				if _, err := os.Stat(filename); os.IsNotExist(err) {
					t.Errorf("output file was not created: %s", filename)
					return
				}

				// Read file content
				content, err := os.ReadFile(filename)
				if err != nil {
					t.Errorf("failed to read output file: %v", err)
					return
				}

				contentStr := string(content)

				// Basic content validation
				if len(contentStr) == 0 {
					t.Errorf("output file is empty")
				}

				if tt.isHTML {
					// Check for HTML structure
					if !strings.Contains(contentStr, "<html lang=\"en\">") {
						t.Errorf("HTML output missing <html> tag")
					}
					if !strings.Contains(contentStr, "Network Information") {
						t.Errorf("HTML output missing expected content")
					}
				} else {
					// Check for text structure
					if !strings.Contains(contentStr, "Network Information:") {
						t.Errorf("text output missing expected content")
					}
				}
			}
		})
	}
}

func TestCLIHandler_showUsage(t *testing.T) {
	handler := NewCLIHandler()

	// Capture usage output (this is mainly to ensure it doesn't panic)
	// In a real scenario, we might want to capture stdout to test the content
	handler.showUsage()

	// Test passes if no panic occurs
}

func TestCLIHandler_ErrorHandling(t *testing.T) {
	handler := NewCLIHandler()

	tests := []struct {
		name        string
		args        []string
		expectError string
	}{
		{
			name:        "empty CIDR",
			args:        []string{"cidr-calc", ""},
			expectError: "CIDR notation is required",
		},
		{
			name:        "malformed CIDR",
			args:        []string{"cidr-calc", "192.168.1"},
			expectError: "failed to parse CIDR",
		},
		{
			name:        "invalid IP in CIDR",
			args:        []string{"cidr-calc", "999.999.999.999/24"},
			expectError: "failed to parse CIDR",
		},
		{
			name:        "invalid prefix length",
			args:        []string{"cidr-calc", "192.168.1.0/99"},
			expectError: "failed to parse CIDR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Run(tt.args)

			if err == nil {
				t.Errorf("expected error but got none")
				return
			}

			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("expected error containing %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

// Benchmark tests for CLI performance
func BenchmarkCLIHandler_Run(b *testing.B) {
	handler := NewCLIHandler()
	args := []string{"cidr-calc", "192.168.1.0/24"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.Run(args)
	}
}

func BenchmarkCLIHandler_parseFlags(b *testing.B) {
	handler := NewCLIHandler()
	args := []string{"cidr-calc", "-h", "-o", "output.html", "192.168.1.0/24"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.parseFlags(args)
	}
}
