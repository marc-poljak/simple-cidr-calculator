package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOutputFormatter_FormatNetworkInfo(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name     string
		network  *NetworkInfo
		expected []string // Expected strings that should be present in output
	}{
		{
			name: "Standard /24 network",
			network: &NetworkInfo{
				NetworkID:     net.ParseIP("192.168.1.0"),
				BroadcastAddr: net.ParseIP("192.168.1.255"),
				SubnetMask:    net.CIDRMask(24, 32),
				WildcardMask:  []byte{0, 0, 0, 255},
				FirstUsableIP: net.ParseIP("192.168.1.1"),
				LastUsableIP:  net.ParseIP("192.168.1.254"),
				TotalHosts:    254,
				PrefixLength:  24,
			},
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
			},
		},
		{
			name: "/32 single host network",
			network: &NetworkInfo{
				NetworkID:     net.ParseIP("10.0.0.1"),
				BroadcastAddr: net.ParseIP("10.0.0.1"),
				SubnetMask:    net.CIDRMask(32, 32),
				WildcardMask:  []byte{0, 0, 0, 0},
				FirstUsableIP: net.ParseIP("10.0.0.1"),
				LastUsableIP:  net.ParseIP("10.0.0.1"),
				TotalHosts:    1,
				PrefixLength:  32,
			},
			expected: []string{
				"Network Information:",
				"CIDR:           10.0.0.1/32",
				"Host Address:   10.0.0.1 (single host)",
				"Total Hosts:    1",
			},
		},
		{
			name: "/31 point-to-point network",
			network: &NetworkInfo{
				NetworkID:     net.ParseIP("172.16.0.0"),
				BroadcastAddr: net.ParseIP("172.16.0.1"),
				SubnetMask:    net.CIDRMask(31, 32),
				WildcardMask:  []byte{0, 0, 0, 1},
				FirstUsableIP: net.ParseIP("172.16.0.0"),
				LastUsableIP:  net.ParseIP("172.16.0.1"),
				TotalHosts:    2,
				PrefixLength:  31,
			},
			expected: []string{
				"Network Information:",
				"CIDR:           172.16.0.0/31",
				"First Address:  172.16.0.0 (point-to-point)",
				"Second Address: 172.16.0.1 (point-to-point)",
				"Total Hosts:    2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatter.FormatNetworkInfo(tt.network)

			// Check that all expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't.\nFull output:\n%s", expected, output)
				}
			}

			// Verify consistent formatting structure
			lines := strings.Split(output, "\n")

			// Should have Network Information section
			foundNetworkSection := false
			foundHostSection := false

			for _, line := range lines {
				if strings.Contains(line, "Network Information:") {
					foundNetworkSection = true
				}
				if strings.Contains(line, "Host Information:") {
					foundHostSection = true
				}
			}

			if !foundNetworkSection {
				t.Error("Output should contain 'Network Information:' section")
			}
			if !foundHostSection {
				t.Error("Output should contain 'Host Information:' section")
			}
		})
	}
}

func TestOutputFormatter_FormatSubnets(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name           string
		subnets        []SubnetInfo
		originalPrefix int
		expected       []string
	}{
		{
			name: "Standard subnet list",
			subnets: []SubnetInfo{
				{
					NetworkID:     net.ParseIP("192.168.1.0"),
					CIDR:          "192.168.1.0/25",
					BroadcastAddr: net.ParseIP("192.168.1.127"),
				},
				{
					NetworkID:     net.ParseIP("192.168.1.128"),
					CIDR:          "192.168.1.128/25",
					BroadcastAddr: net.ParseIP("192.168.1.255"),
				},
			},
			originalPrefix: 24,
			expected: []string{
				"Subnet Information:",
				"Possible /25 Subnets: 2",
				"Subnet List:",
				"192.168.1.0/25     (192.168.1.0 - 192.168.1.127)",
				"192.168.1.128/25   (192.168.1.128 - 192.168.1.255)",
			},
		},
		{
			name:           "Empty subnet list",
			subnets:        []SubnetInfo{},
			originalPrefix: 32,
			expected: []string{
				"Subnet Information:",
				"No subnets available (cannot subnet /32 networks)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatter.FormatSubnets(tt.subnets, tt.originalPrefix)

			// Check that all expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but it didn't.\nFull output:\n%s", expected, output)
				}
			}
		})
	}
}

func TestOutputFormatter_FormatComplete(t *testing.T) {
	formatter := NewOutputFormatter()

	network := &NetworkInfo{
		NetworkID:     net.ParseIP("10.0.0.0"),
		BroadcastAddr: net.ParseIP("10.0.0.255"),
		SubnetMask:    net.CIDRMask(24, 32),
		WildcardMask:  []byte{0, 0, 0, 255},
		FirstUsableIP: net.ParseIP("10.0.0.1"),
		LastUsableIP:  net.ParseIP("10.0.0.254"),
		TotalHosts:    254,
		PrefixLength:  24,
	}

	subnets := []SubnetInfo{
		{
			NetworkID:     net.ParseIP("10.0.0.0"),
			CIDR:          "10.0.0.0/25",
			BroadcastAddr: net.ParseIP("10.0.0.127"),
		},
	}

	output := formatter.FormatComplete(network, subnets)

	// Should contain both network and subnet information
	expectedSections := []string{
		"Network Information:",
		"Host Information:",
		"Subnet Information:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected complete output to contain '%s' section", section)
		}
	}
}

func TestOutputFormatter_FormatIPMask(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name     string
		mask     []byte
		expected string
	}{
		{
			name:     "Standard /24 mask",
			mask:     []byte{255, 255, 255, 0},
			expected: "255.255.255.0",
		},
		{
			name:     "Standard /16 mask",
			mask:     []byte{255, 255, 0, 0},
			expected: "255.255.0.0",
		},
		{
			name:     "Invalid mask length",
			mask:     []byte{255, 255},
			expected: "Invalid mask",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatIPMask(tt.mask)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestOutputFormatter_FormatSubnetRange(t *testing.T) {
	formatter := NewOutputFormatter()

	subnet := SubnetInfo{
		NetworkID:     net.ParseIP("192.168.1.0"),
		CIDR:          "192.168.1.0/25",
		BroadcastAddr: net.ParseIP("192.168.1.127"),
	}

	result := formatter.formatSubnetRange(subnet)
	expected := "(192.168.1.0 - 192.168.1.127)"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestOutputFormatter_FormatError(t *testing.T) {
	formatter := NewOutputFormatter()

	err := fmt.Errorf("test error message")
	result := formatter.FormatError(err)
	expected := "Error: test error message\n"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestOutputFormatter_FormatUsage(t *testing.T) {
	formatter := NewOutputFormatter()

	result := formatter.FormatUsage()

	expectedStrings := []string{
		"Usage: cidr-calc <CIDR>",
		"Example: cidr-calc 192.168.1.0/24",
		"Calculates and displays comprehensive subnet information",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected usage to contain '%s', but it didn't.\nFull output:\n%s", expected, result)
		}
	}
}

func TestOutputFormatter_ConsistentAlignment(t *testing.T) {
	formatter := NewOutputFormatter()

	// Test with different IP lengths to ensure consistent alignment
	networks := []*NetworkInfo{
		{
			NetworkID:     net.ParseIP("1.1.1.0"),
			BroadcastAddr: net.ParseIP("1.1.1.255"),
			SubnetMask:    net.CIDRMask(24, 32),
			WildcardMask:  []byte{0, 0, 0, 255},
			FirstUsableIP: net.ParseIP("1.1.1.1"),
			LastUsableIP:  net.ParseIP("1.1.1.254"),
			TotalHosts:    254,
			PrefixLength:  24,
		},
		{
			NetworkID:     net.ParseIP("192.168.100.0"),
			BroadcastAddr: net.ParseIP("192.168.100.255"),
			SubnetMask:    net.CIDRMask(24, 32),
			WildcardMask:  []byte{0, 0, 0, 255},
			FirstUsableIP: net.ParseIP("192.168.100.1"),
			LastUsableIP:  net.ParseIP("192.168.100.254"),
			TotalHosts:    254,
			PrefixLength:  24,
		},
	}

	for i, network := range networks {
		output := formatter.FormatNetworkInfo(network)
		lines := strings.Split(output, "\n")

		// Check that labels are consistently aligned
		for _, line := range lines {
			if strings.Contains(line, ":") && strings.HasPrefix(line, "  ") {
				// Find the position of the colon
				colonPos := strings.Index(line, ":")
				if colonPos > 0 {
					label := line[2:colonPos] // Skip the "  " prefix
					// Labels should be left-padded to 15 characters for consistent alignment
					if len(label) > 15 {
						t.Errorf("Network %d: Label '%s' is longer than expected alignment width", i, label)
					}
				}
			}
		}
	}
}

func TestOutputFormatter_FormatAsHTML(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name     string
		network  *NetworkInfo
		subnets  []SubnetInfo
		expected []string // Expected strings that should be present in HTML output
	}{
		{
			name: "Standard /24 network with subnets",
			network: &NetworkInfo{
				NetworkID:     net.ParseIP("192.168.1.0"),
				BroadcastAddr: net.ParseIP("192.168.1.255"),
				SubnetMask:    net.CIDRMask(24, 32),
				WildcardMask:  []byte{0, 0, 0, 255},
				FirstUsableIP: net.ParseIP("192.168.1.1"),
				LastUsableIP:  net.ParseIP("192.168.1.254"),
				TotalHosts:    254,
				PrefixLength:  24,
			},
			subnets: []SubnetInfo{
				{
					NetworkID:     net.ParseIP("192.168.1.0"),
					CIDR:          "192.168.1.0/25",
					BroadcastAddr: net.ParseIP("192.168.1.127"),
				},
				{
					NetworkID:     net.ParseIP("192.168.1.128"),
					CIDR:          "192.168.1.128/25",
					BroadcastAddr: net.ParseIP("192.168.1.255"),
				},
			},
			expected: []string{
				"<!DOCTYPE html>",
				"<title>CIDR Calculator Report - 192.168.1.0/24</title>",
				"<div class=\"cidr\">192.168.1.0/24</div>",
				"<h2>Network Information</h2>",
				"<td>192.168.1.0/24</td>",
				"<td>192.168.1.0</td>",
				"<td>192.168.1.255</td>",
				"<td>255.255.255.0</td>",
				"<td>0.0.0.255</td>",
				"<h2>Host Information</h2>",
				"<td>192.168.1.1</td>",
				"<td>192.168.1.254</td>",
				"<td>254</td>",
				"<h2>Subnet Information</h2>",
				"<td>2</td>",
				"<span class=\"subnet-cidr\">192.168.1.0/25</span>",
				"<span class=\"subnet-range\">(192.168.1.0 - 192.168.1.127)</span>",
				"<span class=\"subnet-cidr\">192.168.1.128/25</span>",
				"<span class=\"subnet-range\">(192.168.1.128 - 192.168.1.255)</span>",
				"function toggleSubnets()",
			},
		},
		{
			name: "/32 single host network",
			network: &NetworkInfo{
				NetworkID:     net.ParseIP("10.0.0.1"),
				BroadcastAddr: net.ParseIP("10.0.0.1"),
				SubnetMask:    net.CIDRMask(32, 32),
				WildcardMask:  []byte{0, 0, 0, 0},
				FirstUsableIP: net.ParseIP("10.0.0.1"),
				LastUsableIP:  net.ParseIP("10.0.0.1"),
				TotalHosts:    1,
				PrefixLength:  32,
			},
			subnets: []SubnetInfo{},
			expected: []string{
				"<title>CIDR Calculator Report - 10.0.0.1/32</title>",
				"<div class=\"cidr\">10.0.0.1/32</div>",
				"<td>10.0.0.1 <span style=\"color: #666;\">(single host)</span></td>",
				"<td>1</td>",
				"This is a /32 network representing a single host address",
				"<div class=\"no-subnets\">",
				"No subnets available (cannot subnet /32 networks)",
			},
		},
		{
			name: "/31 point-to-point network",
			network: &NetworkInfo{
				NetworkID:     net.ParseIP("172.16.0.0"),
				BroadcastAddr: net.ParseIP("172.16.0.1"),
				SubnetMask:    net.CIDRMask(31, 32),
				WildcardMask:  []byte{0, 0, 0, 1},
				FirstUsableIP: net.ParseIP("172.16.0.0"),
				LastUsableIP:  net.ParseIP("172.16.0.1"),
				TotalHosts:    2,
				PrefixLength:  31,
			},
			subnets: []SubnetInfo{},
			expected: []string{
				"<title>CIDR Calculator Report - 172.16.0.0/31</title>",
				"<td>172.16.0.0 <span style=\"color: #666;\">(point-to-point)</span></td>",
				"<td>172.16.0.1 <span style=\"color: #666;\">(point-to-point)</span></td>",
				"<td>2</td>",
				"This is a /31 network typically used for point-to-point links",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := formatter.FormatAsHTML(tt.network, tt.subnets)

			// Check that all expected strings are present
			for _, expected := range tt.expected {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected HTML output to contain '%s', but it didn't.\nFull output length: %d", expected, len(output))
				}
			}

			// Verify HTML structure
			if !strings.HasPrefix(output, "<!DOCTYPE html>") {
				t.Error("HTML output should start with DOCTYPE declaration")
			}

			if !strings.Contains(output, "<html lang=\"en\">") {
				t.Error("HTML output should contain proper html tag with lang attribute")
			}

			if !strings.Contains(output, "</html>") {
				t.Error("HTML output should end with closing html tag")
			}

			// Check for essential CSS classes
			essentialClasses := []string{
				"container", "header", "content", "section", "info-table",
			}

			for _, class := range essentialClasses {
				if !strings.Contains(output, fmt.Sprintf("class=\"%s\"", class)) {
					t.Errorf("HTML output should contain CSS class '%s'", class)
				}
			}

			// Verify responsive design meta tag
			if !strings.Contains(output, "viewport") {
				t.Error("HTML output should contain viewport meta tag for responsive design")
			}
		})
	}
}

func TestOutputFormatter_FormatAsHTML_LargeSubnetList(t *testing.T) {
	formatter := NewOutputFormatter()

	// Create a network that would generate many subnets
	network := &NetworkInfo{
		NetworkID:     net.ParseIP("10.0.0.0"),
		BroadcastAddr: net.ParseIP("10.255.255.255"),
		SubnetMask:    net.CIDRMask(8, 32),
		WildcardMask:  []byte{0, 255, 255, 255},
		FirstUsableIP: net.ParseIP("10.0.0.1"),
		LastUsableIP:  net.ParseIP("10.255.255.254"),
		TotalHosts:    16777214,
		PrefixLength:  8,
	}

	// Create 100 subnets to simulate the performance limit
	subnets := make([]SubnetInfo, 100)
	for i := 0; i < 100; i++ {
		subnets[i] = SubnetInfo{
			NetworkID:     net.ParseIP(fmt.Sprintf("10.%d.0.0", i)),
			CIDR:          fmt.Sprintf("10.%d.0.0/9", i),
			BroadcastAddr: net.ParseIP(fmt.Sprintf("10.%d.255.255", i)),
		}
	}

	output := formatter.FormatAsHTML(network, subnets)

	// Should contain performance warning
	if !strings.Contains(output, "Performance Note") {
		t.Error("HTML output should contain performance warning for large subnet lists")
	}

	if !strings.Contains(output, "Showing first 100 subnets") {
		t.Error("HTML output should mention showing first 100 subnets")
	}

	// Should contain toggle functionality
	if !strings.Contains(output, "toggleSubnets()") {
		t.Error("HTML output should contain toggle functionality for large subnet lists")
	}
}

func TestOutputFormatter_SaveToFile(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name        string
		content     string
		filename    string
		expectError bool
		errorMsg    string
		cleanup     func()
	}{
		{
			name:        "Successful file save",
			content:     "<html><body>Test content</body></html>",
			filename:    "test_output.html",
			expectError: false,
			cleanup: func() {
				_ = os.Remove("test_output.html")
			},
		},
		{
			name:        "Empty content",
			content:     "",
			filename:    "test_empty.html",
			expectError: true,
			errorMsg:    "content cannot be empty",
		},
		{
			name:        "Empty filename",
			content:     "test content",
			filename:    "",
			expectError: true,
			errorMsg:    "filename cannot be empty",
		},
		{
			name:        "Path traversal attempt",
			content:     "test content",
			filename:    "../../../etc/passwd",
			expectError: true,
			errorMsg:    "path traversal not allowed",
		},
		{
			name:        "System directory write attempt",
			content:     "test content",
			filename:    "/etc/test.txt",
			expectError: true,
			errorMsg:    "writing to system directories not allowed",
		},
		{
			name:        "Invalid character in filename",
			content:     "test content",
			filename:    "test<file>.html",
			expectError: true,
			errorMsg:    "filename contains invalid character",
		},
		{
			name:        "File in subdirectory",
			content:     "test content",
			filename:    "subdir/test_file.html",
			expectError: false,
			cleanup: func() {
				_ = os.Remove("subdir/test_file.html")
				_ = os.Remove("subdir")
			},
		},
		{
			name:        "Very long filename",
			content:     "test content",
			filename:    strings.Repeat("a", 300) + ".html",
			expectError: true,
			errorMsg:    "filename too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			err := formatter.SaveToFile(tt.content, tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for test '%s', got nil", tt.name)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for test '%s', got: %v", tt.name, err)
				} else {
					// Verify file was created and contains correct content
					if _, err := os.Stat(tt.filename); os.IsNotExist(err) {
						t.Errorf("Expected file to be created for test '%s', but it doesn't exist", tt.name)
					}

					// Read file content and verify
					content, err := os.ReadFile(tt.filename)
					if err != nil {
						t.Errorf("Error reading test file for '%s': %v", tt.name, err)
					}

					if string(content) != tt.content {
						t.Errorf("Expected file content '%s', got '%s' for test '%s'", tt.content, string(content), tt.name)
					}
				}
			}
		})
	}
}

func TestOutputFormatter_SaveTextToFile(t *testing.T) {
	formatter := NewOutputFormatter()

	// Create test network info
	network := &NetworkInfo{
		NetworkID:     net.ParseIP("192.168.1.0"),
		BroadcastAddr: net.ParseIP("192.168.1.255"),
		SubnetMask:    net.CIDRMask(24, 32),
		WildcardMask:  []byte{0, 0, 0, 255},
		FirstUsableIP: net.ParseIP("192.168.1.1"),
		LastUsableIP:  net.ParseIP("192.168.1.254"),
		TotalHosts:    254,
		PrefixLength:  24,
	}

	subnets := []SubnetInfo{
		{
			NetworkID:     net.ParseIP("192.168.1.0"),
			CIDR:          "192.168.1.0/25",
			BroadcastAddr: net.ParseIP("192.168.1.127"),
		},
	}

	tests := []struct {
		name        string
		filename    string
		expectError bool
		errorMsg    string
		cleanup     func()
	}{
		{
			name:        "Valid text file",
			filename:    "test_output.txt",
			expectError: false,
			cleanup: func() {
				_ = os.Remove("test_output.txt")
			},
		},
		{
			name:        "Valid text file with .text extension",
			filename:    "test_output.text",
			expectError: false,
			cleanup: func() {
				_ = os.Remove("test_output.text")
			},
		},
		{
			name:        "Invalid extension for text",
			filename:    "test_output.html",
			expectError: true,
			errorMsg:    "text output requires .txt extension",
		},
		{
			name:        "No extension",
			filename:    "test_output",
			expectError: true,
			errorMsg:    "text output requires .txt extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			err := formatter.SaveTextToFile(network, subnets, tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for test '%s', got nil", tt.name)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for test '%s', got: %v", tt.name, err)
				} else {
					// Verify file was created and contains text content
					content, err := os.ReadFile(tt.filename)
					if err != nil {
						t.Errorf("Error reading test file for '%s': %v", tt.name, err)
					}

					// Verify it contains expected text format elements
					contentStr := string(content)
					expectedElements := []string{
						"Network Information:",
						"Host Information:",
						"Subnet Information:",
						"192.168.1.0/24",
					}

					for _, element := range expectedElements {
						if !strings.Contains(contentStr, element) {
							t.Errorf("Expected text content to contain '%s' for test '%s'", element, tt.name)
						}
					}

					// Verify it doesn't contain HTML tags
					if strings.Contains(contentStr, "<html>") || strings.Contains(contentStr, "<body>") {
						t.Errorf("Text file should not contain HTML tags for test '%s'", tt.name)
					}
				}
			}
		})
	}
}

func TestOutputFormatter_SaveHTMLToFile(t *testing.T) {
	formatter := NewOutputFormatter()

	// Create test network info
	network := &NetworkInfo{
		NetworkID:     net.ParseIP("192.168.1.0"),
		BroadcastAddr: net.ParseIP("192.168.1.255"),
		SubnetMask:    net.CIDRMask(24, 32),
		WildcardMask:  []byte{0, 0, 0, 255},
		FirstUsableIP: net.ParseIP("192.168.1.1"),
		LastUsableIP:  net.ParseIP("192.168.1.254"),
		TotalHosts:    254,
		PrefixLength:  24,
	}

	subnets := []SubnetInfo{
		{
			NetworkID:     net.ParseIP("192.168.1.0"),
			CIDR:          "192.168.1.0/25",
			BroadcastAddr: net.ParseIP("192.168.1.127"),
		},
	}

	tests := []struct {
		name        string
		filename    string
		expectError bool
		errorMsg    string
		cleanup     func()
	}{
		{
			name:        "Valid HTML file",
			filename:    "test_output.html",
			expectError: false,
			cleanup: func() {
				_ = os.Remove("test_output.html")
			},
		},
		{
			name:        "Valid HTML file with .htm extension",
			filename:    "test_output.htm",
			expectError: false,
			cleanup: func() {
				_ = os.Remove("test_output.htm")
			},
		},
		{
			name:        "Invalid extension for HTML",
			filename:    "test_output.txt",
			expectError: true,
			errorMsg:    "HTML output requires .html or .htm extension",
		},
		{
			name:        "No extension",
			filename:    "test_output",
			expectError: true,
			errorMsg:    "HTML output requires .html or .htm extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			err := formatter.SaveHTMLToFile(network, subnets, tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for test '%s', got nil", tt.name)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for test '%s', got: %v", tt.name, err)
				} else {
					// Verify file was created and contains HTML content
					content, err := os.ReadFile(tt.filename)
					if err != nil {
						t.Errorf("Error reading test file for '%s': %v", tt.name, err)
					}

					// Verify it contains expected HTML format elements
					contentStr := string(content)
					expectedElements := []string{
						"<!DOCTYPE html>",
						"<html lang=\"en\">",
						"<head>",
						"<body>",
						"<title>CIDR Calculator Report - 192.168.1.0/24</title>",
						"<h2>Network Information</h2>",
						"<h2>Host Information</h2>",
						"<h2>Subnet Information</h2>",
					}

					for _, element := range expectedElements {
						if !strings.Contains(contentStr, element) {
							t.Errorf("Expected HTML content to contain '%s' for test '%s'", element, tt.name)
						}
					}
				}
			}
		})
	}
}

func TestOutputFormatter_ValidateFilePath(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name        string
		filename    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid filename",
			filename:    "test.html",
			expectError: false,
		},
		{
			name:        "Valid path with subdirectory",
			filename:    "output/test.html",
			expectError: false,
		},
		{
			name:        "Empty filename",
			filename:    "",
			expectError: true,
			errorMsg:    "filename cannot be empty",
		},
		{
			name:        "Whitespace only filename",
			filename:    "   ",
			expectError: true,
			errorMsg:    "filename cannot be empty or whitespace only",
		},
		{
			name:        "Path traversal attempt",
			filename:    "../test.html",
			expectError: true,
			errorMsg:    "path traversal not allowed",
		},
		{
			name:        "System directory",
			filename:    "/etc/test.html",
			expectError: true,
			errorMsg:    "writing to system directories not allowed",
		},
		{
			name:        "Invalid character <",
			filename:    "test<file.html",
			expectError: true,
			errorMsg:    "filename contains invalid character",
		},
		{
			name:        "Invalid character >",
			filename:    "test>file.html",
			expectError: true,
			errorMsg:    "filename contains invalid character",
		},
		{
			name:        "Invalid character |",
			filename:    "test|file.html",
			expectError: true,
			errorMsg:    "filename contains invalid character",
		},
		{
			name:        "Very long filename",
			filename:    strings.Repeat("a", 300) + ".html",
			expectError: true,
			errorMsg:    "filename too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := formatter.validateFilePath(tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for test '%s', got nil", tt.name)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for test '%s', got: %v", tt.name, err)
				}
			}
		})
	}
}

func TestOutputFormatter_EnsureDirectoryExists(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name        string
		filename    string
		expectError bool
		errorMsg    string
		cleanup     func()
	}{
		{
			name:        "Current directory file",
			filename:    "test.html",
			expectError: false,
		},
		{
			name:        "New subdirectory",
			filename:    "testdir/test.html",
			expectError: false,
			cleanup: func() {
				_ = os.Remove("testdir/test.html")
				_ = os.Remove("testdir")
			},
		},
		{
			name:        "Nested subdirectories",
			filename:    "testdir/subdir/test.html",
			expectError: false,
			cleanup: func() {
				_ = os.Remove("testdir/subdir/test.html")
				_ = os.Remove("testdir/subdir")
				_ = os.Remove("testdir")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			err := formatter.ensureDirectoryExists(tt.filename)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for test '%s', got nil", tt.name)
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for test '%s', got: %v", tt.name, err)
				}

				// Verify directory was created if needed
				dir := filepath.Dir(tt.filename)
				if dir != "." {
					if _, err := os.Stat(dir); os.IsNotExist(err) {
						t.Errorf("Expected directory '%s' to be created for test '%s'", dir, tt.name)
					}
				}
			}
		})
	}
}

func TestOutputFormatter_HasValidTextExtension(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"Valid .txt extension", "test.txt", true},
		{"Valid .text extension", "test.text", true},
		{"Valid .TXT extension (case insensitive)", "test.TXT", true},
		{"Invalid .html extension", "test.html", false},
		{"Invalid .htm extension", "test.htm", false},
		{"No extension", "test", false},
		{"Multiple dots with valid extension", "test.backup.txt", true},
		{"Multiple dots with invalid extension", "test.backup.html", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.hasValidTextExtension(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected %v for filename '%s', got %v", tt.expected, tt.filename, result)
			}
		})
	}
}

func TestOutputFormatter_HasValidHTMLExtension(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{"Valid .html extension", "test.html", true},
		{"Valid .htm extension", "test.htm", true},
		{"Valid .HTML extension (case insensitive)", "test.HTML", true},
		{"Invalid .txt extension", "test.txt", false},
		{"Invalid .text extension", "test.text", false},
		{"No extension", "test", false},
		{"Multiple dots with valid extension", "test.backup.html", true},
		{"Multiple dots with invalid extension", "test.backup.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.hasValidHTMLExtension(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected %v for filename '%s', got %v", tt.expected, tt.filename, result)
			}
		})
	}
}

func TestOutputFormatter_FormatIPMaskHTML(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name     string
		mask     []byte
		expected string
	}{
		{
			name:     "Standard /24 mask",
			mask:     []byte{255, 255, 255, 0},
			expected: "255.255.255.0",
		},
		{
			name:     "Standard /16 mask",
			mask:     []byte{255, 255, 0, 0},
			expected: "255.255.0.0",
		},
		{
			name:     "Invalid mask length",
			mask:     []byte{255, 255},
			expected: "Invalid mask",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatIPMaskHTML(tt.mask)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestOutputFormatter_HTMLTemplate_Validation(t *testing.T) {
	formatter := NewOutputFormatter()

	// Test with minimal valid network info
	network := &NetworkInfo{
		NetworkID:     net.ParseIP("192.168.1.0"),
		BroadcastAddr: net.ParseIP("192.168.1.255"),
		SubnetMask:    net.CIDRMask(24, 32),
		WildcardMask:  []byte{0, 0, 0, 255},
		FirstUsableIP: net.ParseIP("192.168.1.1"),
		LastUsableIP:  net.ParseIP("192.168.1.254"),
		TotalHosts:    254,
		PrefixLength:  24,
	}

	output := formatter.FormatAsHTML(network, []SubnetInfo{})

	// Validate HTML structure
	htmlStructureChecks := []string{
		"<!DOCTYPE html>",
		"<html lang=\"en\">",
		"<head>",
		"<meta charset=\"UTF-8\">",
		"<meta name=\"viewport\"",
		"<title>",
		"<style>",
		"</style>",
		"</head>",
		"<body>",
		"</body>",
		"</html>",
	}

	for _, check := range htmlStructureChecks {
		if !strings.Contains(output, check) {
			t.Errorf("HTML output should contain '%s'", check)
		}
	}

	// Validate CSS includes responsive design
	responsiveChecks := []string{
		"@media print",
		"@media (max-width: 768px)",
		"max-width: 1200px",
		"box-shadow:",
		"border-radius:",
	}

	for _, check := range responsiveChecks {
		if !strings.Contains(output, check) {
			t.Errorf("HTML output should contain responsive CSS '%s'", check)
		}
	}

	// Validate JavaScript functionality
	jsChecks := []string{
		"function toggleSubnets()",
		"document.getElementById('subnetList')",
		"addEventListener('DOMContentLoaded'",
	}

	for _, check := range jsChecks {
		if !strings.Contains(output, check) {
			t.Errorf("HTML output should contain JavaScript functionality '%s'", check)
		}
	}
}
