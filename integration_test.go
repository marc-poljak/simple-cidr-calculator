package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestFormatterIntegration tests the formatter with real calculator output
func TestFormatterIntegration(t *testing.T) {
	calculator := NewCIDRCalculator()
	formatter := NewOutputFormatter()

	tests := []struct {
		name     string
		cidr     string
		expected []string
	}{
		{
			name: "Standard /24 network integration",
			cidr: "192.168.1.0/24",
			expected: []string{
				"Network Information:",
				"CIDR:           192.168.1.0/24",
				"Network ID:     192.168.1.0",
				"Broadcast:      192.168.1.255",
				"Subnet Mask:    255.255.255.0",
				"Wildcard Mask:  0.0.0.255",
				"Host Information:",
				"First Usable:   192.168.1.1",
				"Last Usable:    192.168.1.254",
				"Total Hosts:    254",
				"Subnet Information:",
				"Possible /25 Subnets: 2",
				"192.168.1.0/25",
				"192.168.1.128/25",
			},
		},
		{
			name: "/26 network integration",
			cidr: "172.21.4.0/26",
			expected: []string{
				"CIDR:           172.21.4.0/26",
				"Network ID:     172.21.4.0",
				"Broadcast:      172.21.4.63",
				"Subnet Mask:    255.255.255.192",
				"Wildcard Mask:  0.0.0.63",
				"First Usable:   172.21.4.1",
				"Last Usable:    172.21.4.62",
				"Total Hosts:    62",
				"Possible /27 Subnets: 2",
			},
		},
		{
			name: "/32 single host integration",
			cidr: "10.0.0.1/32",
			expected: []string{
				"CIDR:           10.0.0.1/32",
				"Host Address:   10.0.0.1 (single host)",
				"Total Hosts:    1",
				"No subnets available (cannot subnet /32 networks)",
			},
		},
		{
			name: "/31 point-to-point integration",
			cidr: "172.16.0.0/31",
			expected: []string{
				"CIDR:           172.16.0.0/31",
				"First Address:  172.16.0.0 (point-to-point)",
				"Second Address: 172.16.0.1 (point-to-point)",
				"Total Hosts:    2",
				"Possible /32 Subnets: 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse CIDR using calculator
			networkInfo, err := calculator.ParseCIDR(tt.cidr)
			if err != nil {
				t.Fatalf("Failed to parse CIDR %s: %v", tt.cidr, err)
			}

			// Calculate subnets
			subnets := calculator.CalculateSubnets(networkInfo)

			// Format complete output
			output := formatter.FormatComplete(networkInfo, subnets)

			// Verify all expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't.\nFull output:\n%s", expected, output)
				}
			}

			// Verify output structure
			if !strings.Contains(output, "Network Information:") {
				t.Error("Output should contain Network Information section")
			}
			if !strings.Contains(output, "Host Information:") {
				t.Error("Output should contain Host Information section")
			}
			if !strings.Contains(output, "Subnet Information:") {
				t.Error("Output should contain Subnet Information section")
			}
		})
	}
}

// TestFormatterErrorHandling tests error formatting
func TestFormatterErrorHandling(t *testing.T) {
	calculator := NewCIDRCalculator()
	formatter := NewOutputFormatter()

	invalidCIDRs := []string{
		"invalid",
		"192.168.1.0/33",
		"256.256.256.256/24",
		"192.168.1.0",
		"",
	}

	for _, cidr := range invalidCIDRs {
		t.Run("Invalid CIDR: "+cidr, func(t *testing.T) {
			_, err := calculator.ParseCIDR(cidr)
			if err == nil {
				t.Errorf("Expected error for invalid CIDR: %s", cidr)
				return
			}

			// Format the error
			errorOutput := formatter.FormatError(err)

			// Should start with "Error: "
			if !strings.HasPrefix(errorOutput, "Error: ") {
				t.Errorf("Error output should start with 'Error: ', got: %s", errorOutput)
			}

			// Should end with newline
			if !strings.HasSuffix(errorOutput, "\n") {
				t.Errorf("Error output should end with newline, got: %s", errorOutput)
			}
		})
	}
}

// TestCompleteApplicationWorkflow tests the complete end-to-end application workflow
func TestCompleteApplicationWorkflow(t *testing.T) {
	handler := NewCLIHandler()
	tempDir := t.TempDir()

	tests := []struct {
		name           string
		args           []string
		expectError    bool
		checkConsole   bool
		checkFile      bool
		expectedFile   string
		isHTML         bool
		expectedOutput []string
	}{
		{
			name:         "Console output workflow",
			args:         []string{"cidr-calc", "192.168.100.0/26"},
			expectError:  false,
			checkConsole: true,
			expectedOutput: []string{
				"Network Information:",
				"CIDR:           192.168.100.0/26",
				"Network ID:     192.168.100.0",
				"Broadcast:      192.168.100.63",
				"First Usable:   192.168.100.1",
				"Last Usable:    192.168.100.62",
				"Total Hosts:    62",
				"Possible /27 Subnets: 2",
			},
		},
		{
			name:         "Text file output workflow",
			args:         []string{"cidr-calc", "-o", tempDir + "/network.txt", "10.0.0.0/30"},
			expectError:  false,
			checkFile:    true,
			expectedFile: tempDir + "/network.txt",
			isHTML:       false,
			expectedOutput: []string{
				"Network Information:",
				"CIDR:           10.0.0.0/30",
				"Total Hosts:    2",
				"Possible /31 Subnets: 2",
			},
		},
		{
			name:         "HTML file output workflow",
			args:         []string{"cidr-calc", "--html", "-o", tempDir + "/report.html", "172.16.0.0/28"},
			expectError:  false,
			checkFile:    true,
			expectedFile: tempDir + "/report.html",
			isHTML:       true,
			expectedOutput: []string{
				"<!DOCTYPE html>",
				"<title>CIDR Calculator Report - 172.16.0.0/28</title>",
				"172.16.0.0/28",
				"Network Information",
				"Host Information",
				"Subnet Information",
			},
		},
		{
			name:        "Error handling workflow - invalid CIDR",
			args:        []string{"cidr-calc", "invalid-input"},
			expectError: true,
		},
		{
			name:        "Error handling workflow - missing CIDR",
			args:        []string{"cidr-calc"},
			expectError: true,
		},
		{
			name:         "Edge case workflow - /32 network",
			args:         []string{"cidr-calc", "192.168.1.100/32"},
			expectError:  false,
			checkConsole: true,
			expectedOutput: []string{
				"Host Address:   192.168.1.100 (single host)",
				"Total Hosts:    1",
				"No subnets available (cannot subnet /32 networks)",
			},
		},
		{
			name:         "Edge case workflow - /31 network",
			args:         []string{"cidr-calc", "10.1.1.0/31"},
			expectError:  false,
			checkConsole: true,
			expectedOutput: []string{
				"First Address:  10.1.1.0 (point-to-point)",
				"Second Address: 10.1.1.1 (point-to-point)",
				"Total Hosts:    2",
				"Possible /32 Subnets: 2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute the CLI handler
			err := handler.Run(tt.args)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
				return
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// If we expect an error, we're done
			if tt.expectError {
				return
			}

			// Check file output if specified
			if tt.checkFile {
				// Verify file exists
				if _, err := os.Stat(tt.expectedFile); os.IsNotExist(err) {
					t.Errorf("Expected output file was not created: %s", tt.expectedFile)
					return
				}

				// Read and verify file content
				content, err := os.ReadFile(tt.expectedFile)
				if err != nil {
					t.Errorf("Failed to read output file: %v", err)
					return
				}

				contentStr := string(content)
				if len(contentStr) == 0 {
					t.Errorf("Output file is empty")
					return
				}

				// Verify expected content
				for _, expected := range tt.expectedOutput {
					if !strings.Contains(contentStr, expected) {
						t.Errorf("File output missing expected content: %s\nFull content:\n%s", expected, contentStr)
					}
				}

				// Additional HTML-specific checks
				if tt.isHTML {
					if !strings.Contains(contentStr, "<html lang=\"en\">") {
						t.Errorf("HTML output missing proper HTML structure")
					}
					if !strings.Contains(contentStr, "</html>") {
						t.Errorf("HTML output missing closing HTML tag")
					}
				}
			}
		})
	}
}

// TestApplicationErrorHandlingWorkflow tests comprehensive error handling scenarios
func TestApplicationErrorHandlingWorkflow(t *testing.T) {
	handler := NewCLIHandler()
	tempDir := t.TempDir()

	errorTests := []struct {
		name        string
		args        []string
		expectError string
	}{
		{
			name:        "Invalid CIDR format",
			args:        []string{"cidr-calc", "192.168.1"},
			expectError: "failed to parse CIDR",
		},
		{
			name:        "Invalid IP address",
			args:        []string{"cidr-calc", "999.999.999.999/24"},
			expectError: "failed to parse CIDR",
		},
		{
			name:        "Invalid prefix length",
			args:        []string{"cidr-calc", "192.168.1.0/99"},
			expectError: "failed to parse CIDR",
		},
		{
			name:        "HTML flag with wrong extension",
			args:        []string{"cidr-calc", "--html", "-o", tempDir + "/output.txt", "192.168.1.0/24"},
			expectError: "HTML output requires .html or .htm file extension",
		},
		{
			name:        "HTML extension without HTML flag",
			args:        []string{"cidr-calc", "-o", tempDir + "/output.html", "192.168.1.0/24"},
			expectError: "HTML file extension requires --html flag",
		},
	}

	for _, tt := range errorTests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Run(tt.args)

			if err == nil {
				t.Errorf("Expected error but got none")
				return
			}

			if !strings.Contains(err.Error(), tt.expectError) {
				t.Errorf("Expected error containing %q, got %q", tt.expectError, err.Error())
			}
		})
	}
}

// TestApplicationPerformanceWorkflow tests performance with various network sizes
func TestApplicationPerformanceWorkflow(t *testing.T) {
	handler := NewCLIHandler()

	performanceTests := []struct {
		name string
		cidr string
	}{
		{"Small network /28", "192.168.1.0/28"},
		{"Medium network /24", "192.168.0.0/24"},
		{"Large network /20", "172.16.0.0/20"},
		{"Very large network /16", "10.0.0.0/16"},
	}

	for _, tt := range performanceTests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"cidr-calc", tt.cidr}

			// Measure execution time
			start := time.Now()
			err := handler.Run(args)
			duration := time.Since(start)

			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.cidr, err)
				return
			}

			// Performance should be reasonable (under 1 second for all test cases)
			if duration > time.Second {
				t.Errorf("Performance issue: %s took %v to process", tt.cidr, duration)
			}

			t.Logf("%s processed in %v", tt.cidr, duration)
		})
	}
}
