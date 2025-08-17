package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// OutputFormatter handles formatting of network information for console output
type OutputFormatter struct{}

// NewOutputFormatter creates a new output formatter instance
func NewOutputFormatter() *OutputFormatter {
	return &OutputFormatter{}
}

// FormatNetworkInfo formats comprehensive network information for console display
func (f *OutputFormatter) FormatNetworkInfo(info *NetworkInfo) string {
	var output strings.Builder

	// Network Information Section
	output.WriteString("Network Information:\n")
	output.WriteString(fmt.Sprintf("  %-15s %s\n", "CIDR:", fmt.Sprintf("%s/%d", info.NetworkID.String(), info.PrefixLength)))
	output.WriteString(fmt.Sprintf("  %-15s %s\n", "Network ID:", info.NetworkID.String()))
	output.WriteString(fmt.Sprintf("  %-15s %s\n", "Broadcast:", info.BroadcastAddr.String()))
	output.WriteString(fmt.Sprintf("  %-15s %s\n", "Subnet Mask:", f.formatIPMask(info.SubnetMask)))
	output.WriteString(fmt.Sprintf("  %-15s %s\n", "Wildcard Mask:", f.formatIPMask(info.WildcardMask)))
	output.WriteString("\n")

	// Host Information Section
	output.WriteString("Host Information:\n")

	// Handle edge cases for /31 and /32 networks
	switch info.PrefixLength {
	case 32:
		output.WriteString(fmt.Sprintf("  %-15s %s (single host)\n", "Host Address:", info.FirstUsableIP.String()))
		output.WriteString(fmt.Sprintf("  %-15s %d\n", "Total Hosts:", info.TotalHosts))
	case 31:
		output.WriteString(fmt.Sprintf("  %-15s %s (point-to-point)\n", "First Address:", info.FirstUsableIP.String()))
		output.WriteString(fmt.Sprintf("  %-15s %s (point-to-point)\n", "Second Address:", info.LastUsableIP.String()))
		output.WriteString(fmt.Sprintf("  %-15s %d\n", "Total Hosts:", info.TotalHosts))
	default:
		output.WriteString(fmt.Sprintf("  %-15s %s\n", "First Usable:", info.FirstUsableIP.String()))
		output.WriteString(fmt.Sprintf("  %-15s %s\n", "Last Usable:", info.LastUsableIP.String()))
		output.WriteString(fmt.Sprintf("  %-15s %d\n", "Total Hosts:", info.TotalHosts))
	}

	return output.String()
}

// FormatSubnets formats subnet information for console display
func (f *OutputFormatter) FormatSubnets(subnets []SubnetInfo, originalPrefix int) string {
	if len(subnets) == 0 {
		return "Subnet Information:\n  No subnets available (cannot subnet /32 networks)\n"
	}

	var output strings.Builder
	nextPrefix := originalPrefix + 1

	// Subnet Information Header
	output.WriteString("Subnet Information:\n")
	output.WriteString(fmt.Sprintf("  Possible /%d Subnets: %d\n", nextPrefix, len(subnets)))

	// Add note for limited display if applicable
	if originalPrefix <= 16 && len(subnets) == 100 {
		output.WriteString("  (Showing first 100 subnets for performance)\n")
	}

	output.WriteString("\n")
	output.WriteString("  Subnet List:\n")

	// Format each subnet with consistent alignment
	for _, subnet := range subnets {
		// Calculate the range for display
		rangeStr := f.formatSubnetRange(subnet)
		output.WriteString(fmt.Sprintf("    %-18s %s\n", subnet.CIDR, rangeStr))
	}

	return output.String()
}

// FormatComplete formats both network information and subnets together
func (f *OutputFormatter) FormatComplete(info *NetworkInfo, subnets []SubnetInfo) string {
	var output strings.Builder

	// Add network information
	output.WriteString(f.FormatNetworkInfo(info))
	output.WriteString("\n")

	// Add subnet information
	output.WriteString(f.FormatSubnets(subnets, info.PrefixLength))

	return output.String()
}

// formatIPMask converts an IP mask to dotted decimal notation
func (f *OutputFormatter) formatIPMask(mask []byte) string {
	if len(mask) != 4 {
		return "Invalid mask"
	}
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
}

// formatSubnetRange creates a formatted range string for a subnet
func (f *OutputFormatter) formatSubnetRange(subnet SubnetInfo) string {
	return fmt.Sprintf("(%s - %s)", subnet.NetworkID.String(), subnet.BroadcastAddr.String())
}

// FormatError formats error messages with consistent styling
func (f *OutputFormatter) FormatError(err error) string {
	return fmt.Sprintf("Error: %s\n", err.Error())
}

// FormatUsage formats usage instructions
func (f *OutputFormatter) FormatUsage() string {
	var output strings.Builder

	output.WriteString("Usage: cidr-calc <CIDR>\n")
	output.WriteString("Example: cidr-calc 192.168.1.0/24\n")
	output.WriteString("\n")
	output.WriteString("Calculates and displays comprehensive subnet information for the given CIDR block.\n")

	return output.String()
}

// FormatAsHTML generates HTML formatted output with embedded CSS styling
func (f *OutputFormatter) FormatAsHTML(info *NetworkInfo, subnets []SubnetInfo) string {
	tmpl := template.Must(template.New("cidr-report").Parse(htmlTemplate))

	data := struct {
		NetworkInfo *NetworkInfo
		Subnets     []SubnetInfo
		HasSubnets  bool
		NextPrefix  int
		SubnetCount int
		ShowLimited bool
	}{
		NetworkInfo: info,
		Subnets:     subnets,
		HasSubnets:  len(subnets) > 0,
		NextPrefix:  info.PrefixLength + 1,
		SubnetCount: len(subnets),
		ShowLimited: info.PrefixLength <= 16 && len(subnets) == 100,
	}

	var output strings.Builder
	if err := tmpl.Execute(&output, data); err != nil {
		return fmt.Sprintf("Error generating HTML: %v", err)
	}

	return output.String()
}

// SaveToFile saves content to a specified file with comprehensive error handling and validation
func (f *OutputFormatter) SaveToFile(content string, filename string) error {
	// Validate input parameters
	if content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Validate and sanitize file path
	if err := f.validateFilePath(filename); err != nil {
		return fmt.Errorf("invalid file path: %v", err)
	}

	// Create directory if it doesn't exist
	if err := f.ensureDirectoryExists(filename); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create file with proper permissions
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filename, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log close error but don't override the main error
			fmt.Fprintf(os.Stderr, "Warning: failed to close file %s: %v\n", filename, closeErr)
		}
	}()

	// Write content to file
	bytesWritten, err := file.WriteString(content)
	if err != nil {
		return fmt.Errorf("failed to write to file %s: %v", filename, err)
	}

	// Verify all content was written
	if bytesWritten != len(content) {
		return fmt.Errorf("incomplete write to file %s: wrote %d bytes, expected %d", filename, bytesWritten, len(content))
	}

	// Sync to ensure data is written to disk
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file %s: %v", filename, err)
	}

	return nil
}

// SaveTextToFile saves text content to a file with .txt extension validation
func (f *OutputFormatter) SaveTextToFile(info *NetworkInfo, subnets []SubnetInfo, filename string) error {
	// Generate text content
	content := f.FormatComplete(info, subnets)

	// Validate file extension for text output
	if !f.hasValidTextExtension(filename) {
		return fmt.Errorf("text output requires .txt extension, got: %s", filename)
	}

	return f.SaveToFile(content, filename)
}

// SaveHTMLToFile saves HTML content to a file with .html extension validation
func (f *OutputFormatter) SaveHTMLToFile(info *NetworkInfo, subnets []SubnetInfo, filename string) error {
	// Generate HTML content
	content := f.FormatAsHTML(info, subnets)

	// Validate file extension for HTML output
	if !f.hasValidHTMLExtension(filename) {
		return fmt.Errorf("HTML output requires .html or .htm extension, got: %s", filename)
	}

	return f.SaveToFile(content, filename)
}

// formatIPMaskHTML formats IP mask for HTML display
func (f *OutputFormatter) formatIPMaskHTML(mask []byte) string {
	if len(mask) != 4 {
		return "Invalid mask"
	}
	return fmt.Sprintf("%d.%d.%d.%d", mask[0], mask[1], mask[2], mask[3])
}

// validateFilePath validates the file path for security and correctness
func (f *OutputFormatter) validateFilePath(filename string) error {
	// Check for empty filename
	if strings.TrimSpace(filename) == "" {
		return fmt.Errorf("filename cannot be empty or whitespace only")
	}

	// Clean the path to resolve any relative path components
	cleanPath := filepath.Clean(filename)

	// Check for path traversal attempts
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("path traversal not allowed: %s", filename)
	}

	// Check for absolute paths that might be dangerous
	if filepath.IsAbs(cleanPath) {
		// Allow absolute paths but warn about potential issues
		if strings.HasPrefix(cleanPath, "/etc/") || strings.HasPrefix(cleanPath, "/sys/") ||
			strings.HasPrefix(cleanPath, "/proc/") || strings.HasPrefix(cleanPath, "/dev/") {
			return fmt.Errorf("writing to system directories not allowed: %s", cleanPath)
		}
	}

	// Check filename length (reasonable limit)
	if len(filepath.Base(cleanPath)) > 255 {
		return fmt.Errorf("filename too long (max 255 characters): %s", filepath.Base(cleanPath))
	}

	// Check for invalid characters in filename
	invalidChars := []string{"\x00", "<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(cleanPath, char) {
			return fmt.Errorf("filename contains invalid character '%s': %s", char, cleanPath)
		}
	}

	return nil
}

// ensureDirectoryExists creates the directory structure if it doesn't exist
func (f *OutputFormatter) ensureDirectoryExists(filename string) error {
	dir := filepath.Dir(filename)

	// If directory is current directory, no need to create
	if dir == "." {
		return nil
	}

	// Check if directory already exists
	if info, err := os.Stat(dir); err == nil {
		if !info.IsDir() {
			return fmt.Errorf("path exists but is not a directory: %s", dir)
		}
		return nil
	}

	// Create directory with appropriate permissions
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	return nil
}

// hasValidTextExtension checks if filename has a valid text extension
func (f *OutputFormatter) hasValidTextExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := []string{".txt", ".text"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// hasValidHTMLExtension checks if filename has a valid HTML extension
func (f *OutputFormatter) hasValidHTMLExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := []string{".html", ".htm"}

	for _, validExt := range validExtensions {
		if ext == validExt {
			return true
		}
	}
	return false
}

// HTML template with embedded CSS for professional styling
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CIDR Calculator Report - {{.NetworkInfo.NetworkID}}/{{.NetworkInfo.PrefixLength}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f5f5f5;
            padding: 20px;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
            overflow: hidden;
        }
        
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            text-align: center;
        }
        
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            font-weight: 300;
        }
        
        .header .cidr {
            font-size: 1.5em;
            font-family: 'Courier New', monospace;
            background: rgba(255,255,255,0.2);
            padding: 10px 20px;
            border-radius: 25px;
            display: inline-block;
        }
        
        .content {
            padding: 30px;
        }
        
        .section {
            margin-bottom: 40px;
        }
        
        .section h2 {
            color: #667eea;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
            margin-bottom: 20px;
            font-size: 1.5em;
        }
        
        .info-table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 20px;
            background: white;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        
        .info-table th,
        .info-table td {
            padding: 15px;
            text-align: left;
            border-bottom: 1px solid #eee;
        }
        
        .info-table th {
            background: #f8f9fa;
            font-weight: 600;
            color: #555;
            width: 200px;
        }
        
        .info-table td {
            font-family: 'Courier New', monospace;
            font-size: 1.1em;
            color: #333;
        }
        
        .info-table tr:last-child td,
        .info-table tr:last-child th {
            border-bottom: none;
        }
        
        .info-table tr:hover {
            background: #f8f9fa;
        }
        
        .subnet-controls {
            margin-bottom: 20px;
        }
        
        .toggle-btn {
            background: #667eea;
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 6px;
            cursor: pointer;
            font-size: 1em;
            transition: background 0.3s ease;
        }
        
        .toggle-btn:hover {
            background: #5a6fd8;
        }
        
        .subnet-list {
            max-height: 400px;
            overflow-y: auto;
            border: 1px solid #ddd;
            border-radius: 6px;
            background: white;
        }
        
        .subnet-item {
            padding: 12px 20px;
            border-bottom: 1px solid #eee;
            display: flex;
            justify-content: space-between;
            align-items: center;
            transition: background 0.2s ease;
        }
        
        .subnet-item:last-child {
            border-bottom: none;
        }
        
        .subnet-item:hover {
            background: #f8f9fa;
        }
        
        .subnet-cidr {
            font-family: 'Courier New', monospace;
            font-weight: bold;
            color: #667eea;
            min-width: 150px;
        }
        
        .subnet-range {
            font-family: 'Courier New', monospace;
            color: #666;
        }
        
        .warning {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            color: #856404;
            padding: 15px;
            border-radius: 6px;
            margin-bottom: 20px;
        }
        
        .no-subnets {
            text-align: center;
            color: #666;
            font-style: italic;
            padding: 40px;
            background: #f8f9fa;
            border-radius: 6px;
        }
        
        .special-case {
            background: #e3f2fd;
            border-left: 4px solid #2196f3;
            padding: 15px;
            margin: 15px 0;
            border-radius: 0 6px 6px 0;
        }
        
        .special-case .label {
            font-weight: bold;
            color: #1976d2;
        }
        
        @media print {
            body {
                background: white;
                padding: 0;
            }
            
            .container {
                box-shadow: none;
                border-radius: 0;
            }
            
            .header {
                background: #667eea !important;
                -webkit-print-color-adjust: exact;
            }
            
            .toggle-btn {
                display: none;
            }
            
            .subnet-list {
                max-height: none;
                overflow: visible;
            }
        }
        
        @media (max-width: 768px) {
            body {
                padding: 10px;
            }
            
            .header {
                padding: 20px;
            }
            
            .header h1 {
                font-size: 2em;
            }
            
            .header .cidr {
                font-size: 1.2em;
                padding: 8px 16px;
            }
            
            .content {
                padding: 20px;
            }
            
            .info-table th,
            .info-table td {
                padding: 10px;
            }
            
            .info-table th {
                width: 120px;
                font-size: 0.9em;
            }
            
            .info-table td {
                font-size: 0.9em;
            }
            
            .subnet-item {
                flex-direction: column;
                align-items: flex-start;
                gap: 5px;
            }
            
            .subnet-cidr {
                min-width: auto;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>CIDR Calculator Report</h1>
            <div class="cidr">{{.NetworkInfo.NetworkID}}/{{.NetworkInfo.PrefixLength}}</div>
        </div>
        
        <div class="content">
            <div class="section">
                <h2>Network Information</h2>
                <table class="info-table">
                    <tr>
                        <th>CIDR</th>
                        <td>{{.NetworkInfo.NetworkID}}/{{.NetworkInfo.PrefixLength}}</td>
                    </tr>
                    <tr>
                        <th>Network ID</th>
                        <td>{{.NetworkInfo.NetworkID}}</td>
                    </tr>
                    <tr>
                        <th>Broadcast Address</th>
                        <td>{{.NetworkInfo.BroadcastAddr}}</td>
                    </tr>
                    <tr>
                        <th>Subnet Mask</th>
                        <td>{{printf "%d.%d.%d.%d" (index .NetworkInfo.SubnetMask 0) (index .NetworkInfo.SubnetMask 1) (index .NetworkInfo.SubnetMask 2) (index .NetworkInfo.SubnetMask 3)}}</td>
                    </tr>
                    <tr>
                        <th>Wildcard Mask</th>
                        <td>{{printf "%d.%d.%d.%d" (index .NetworkInfo.WildcardMask 0) (index .NetworkInfo.WildcardMask 1) (index .NetworkInfo.WildcardMask 2) (index .NetworkInfo.WildcardMask 3)}}</td>
                    </tr>
                </table>
            </div>
            
            <div class="section">
                <h2>Host Information</h2>
                <table class="info-table">
                    {{if eq .NetworkInfo.PrefixLength 32}}
                        <tr>
                            <th>Host Address</th>
                            <td>{{.NetworkInfo.FirstUsableIP}} <span style="color: #666;">(single host)</span></td>
                        </tr>
                    {{else if eq .NetworkInfo.PrefixLength 31}}
                        <tr>
                            <th>First Address</th>
                            <td>{{.NetworkInfo.FirstUsableIP}} <span style="color: #666;">(point-to-point)</span></td>
                        </tr>
                        <tr>
                            <th>Second Address</th>
                            <td>{{.NetworkInfo.LastUsableIP}} <span style="color: #666;">(point-to-point)</span></td>
                        </tr>
                    {{else}}
                        <tr>
                            <th>First Usable IP</th>
                            <td>{{.NetworkInfo.FirstUsableIP}}</td>
                        </tr>
                        <tr>
                            <th>Last Usable IP</th>
                            <td>{{.NetworkInfo.LastUsableIP}}</td>
                        </tr>
                    {{end}}
                    <tr>
                        <th>Total Hosts</th>
                        <td>{{.NetworkInfo.TotalHosts}}</td>
                    </tr>
                </table>
                
                {{if eq .NetworkInfo.PrefixLength 32}}
                    <div class="special-case">
                        <span class="label">Note:</span> This is a /32 network representing a single host address.
                    </div>
                {{else if eq .NetworkInfo.PrefixLength 31}}
                    <div class="special-case">
                        <span class="label">Note:</span> This is a /31 network typically used for point-to-point links with no broadcast address.
                    </div>
                {{end}}
            </div>
            
            <div class="section">
                <h2>Subnet Information</h2>
                {{if .HasSubnets}}
                    <table class="info-table">
                        <tr>
                            <th>Possible /{{.NextPrefix}} Subnets</th>
                            <td>{{.SubnetCount}}</td>
                        </tr>
                    </table>
                    
                    {{if .ShowLimited}}
                        <div class="warning">
                            <strong>Performance Note:</strong> Showing first 100 subnets for performance. The network can be divided into {{.SubnetCount}} total subnets.
                        </div>
                    {{end}}
                    
                    <div class="subnet-controls">
                        <button class="toggle-btn" onclick="toggleSubnets()">Toggle Subnet List</button>
                    </div>
                    
                    <div class="subnet-list" id="subnetList">
                        {{range .Subnets}}
                            <div class="subnet-item">
                                <span class="subnet-cidr">{{.CIDR}}</span>
                                <span class="subnet-range">({{.NetworkID}} - {{.BroadcastAddr}})</span>
                            </div>
                        {{end}}
                    </div>
                {{else}}
                    <div class="no-subnets">
                        No subnets available (cannot subnet /32 networks)
                    </div>
                {{end}}
            </div>
        </div>
    </div>
    
    <script>
        function toggleSubnets() {
            const subnetList = document.getElementById('subnetList');
            const btn = document.querySelector('.toggle-btn');
            
            if (subnetList.style.display === 'none') {
                subnetList.style.display = 'block';
                btn.textContent = 'Hide Subnet List';
            } else {
                subnetList.style.display = 'none';
                btn.textContent = 'Show Subnet List';
            }
        }
        
        // Initially hide subnet list if there are many subnets
        document.addEventListener('DOMContentLoaded', function() {
            const subnetList = document.getElementById('subnetList');
            const subnetCount = {{.SubnetCount}};
            
            if (subnetCount > 20) {
                subnetList.style.display = 'none';
                document.querySelector('.toggle-btn').textContent = 'Show Subnet List';
            }
        });
    </script>
</body>
</html>`
