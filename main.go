package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Config holds command-line configuration options
type Config struct {
	CIDR       string
	OutputFile string
	HTMLOutput bool
	ShowHelp   bool
}

// CLIHandler manages command-line interface operations
type CLIHandler struct {
	calculator *CIDRCalculator
	formatter  *OutputFormatter
}

// NewCLIHandler creates a new CLI handler instance
func NewCLIHandler() *CLIHandler {
	return &CLIHandler{
		calculator: NewCIDRCalculator(),
		formatter:  NewOutputFormatter(),
	}
}

// Run executes the CLI application with provided arguments
func (c *CLIHandler) Run(args []string) error {
	// Parse command-line flags
	config, err := c.parseFlags(args)
	if err != nil {
		return err
	}

	// Show help if requested
	if config.ShowHelp {
		c.showUsage()
		return nil
	}

	// Validate CIDR input
	if config.CIDR == "" {
		c.showUsage()
		return fmt.Errorf("CIDR notation is required")
	}

	// Parse and calculate network information
	networkInfo, err := c.calculator.ParseCIDR(config.CIDR)
	if err != nil {
		return fmt.Errorf("failed to parse CIDR: %v", err)
	}

	// Calculate subnets
	subnets := c.calculator.CalculateSubnets(networkInfo)

	// Handle output based on configuration
	return c.handleOutput(networkInfo, subnets, config)
}

// parseFlags parses command-line arguments and returns configuration
func (c *CLIHandler) parseFlags(args []string) (*Config, error) {
	config := &Config{}

	// Create custom flag set to avoid conflicts with testing
	flagSet := flag.NewFlagSet("cidr-calc", flag.ContinueOnError)

	// Capture help output to avoid printing during parsing
	var helpOutput strings.Builder
	flagSet.SetOutput(&helpOutput)

	// Define flags
	flagSet.StringVar(&config.OutputFile, "o", "", "Save output to file")
	flagSet.StringVar(&config.OutputFile, "output", "", "Save output to file")
	flagSet.BoolVar(&config.HTMLOutput, "h", false, "Generate HTML formatted output")
	flagSet.BoolVar(&config.HTMLOutput, "html", false, "Generate HTML formatted output")
	flagSet.BoolVar(&config.ShowHelp, "help", false, "Show help message")

	// Parse flags
	err := flagSet.Parse(args[1:]) // Skip program name
	if err != nil {
		if err == flag.ErrHelp {
			config.ShowHelp = true
			return config, nil
		}
		return nil, fmt.Errorf("flag parsing error: %v", err)
	}

	// Get remaining arguments (should be CIDR)
	remaining := flagSet.Args()
	if len(remaining) > 0 {
		config.CIDR = remaining[0]
	}

	// Validate flag combinations
	if err := c.validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

// validateConfig validates the configuration for consistency
func (c *CLIHandler) validateConfig(config *Config) error {
	// If HTML output is requested, ensure output file has proper extension
	if config.HTMLOutput && config.OutputFile != "" {
		if !strings.HasSuffix(strings.ToLower(config.OutputFile), ".html") &&
			!strings.HasSuffix(strings.ToLower(config.OutputFile), ".htm") {
			return fmt.Errorf("HTML output requires .html or .htm file extension")
		}
	}

	// If output file is specified without HTML flag, ensure it's not HTML extension
	if !config.HTMLOutput && config.OutputFile != "" {
		ext := strings.ToLower(config.OutputFile)
		if strings.HasSuffix(ext, ".html") || strings.HasSuffix(ext, ".htm") {
			return fmt.Errorf("HTML file extension requires --html flag")
		}
	}

	return nil
}

// handleOutput processes and outputs the results based on configuration
func (c *CLIHandler) handleOutput(networkInfo *NetworkInfo, subnets []SubnetInfo, config *Config) error {
	if config.OutputFile != "" {
		// Save to file
		if config.HTMLOutput {
			return c.formatter.SaveHTMLToFile(networkInfo, subnets, config.OutputFile)
		} else {
			return c.formatter.SaveTextToFile(networkInfo, subnets, config.OutputFile)
		}
	} else {
		// Output to console
		if config.HTMLOutput {
			// HTML output to console
			htmlContent := c.formatter.FormatAsHTML(networkInfo, subnets)
			fmt.Print(htmlContent)
		} else {
			// Text output to console
			textContent := c.formatter.FormatComplete(networkInfo, subnets)
			fmt.Print(textContent)
		}
	}

	return nil
}

// showUsage displays usage instructions and examples
func (c *CLIHandler) showUsage() {
	fmt.Print(`CIDR Calculator - Network Subnet Information Tool

Usage:
  cidr-calc [OPTIONS] <CIDR>

Arguments:
  CIDR                 Network in CIDR notation (e.g., 192.168.1.0/24)

Options:
  -o, --output FILE    Save output to specified file
  -h, --html          Generate HTML formatted output
  --help              Show this help message

Examples:
  cidr-calc 192.168.1.0/24
  cidr-calc -o report.txt 172.16.0.0/16
  cidr-calc --html -o network.html 10.0.0.0/8
  cidr-calc --help

Description:
  Calculates and displays comprehensive subnet information for the given CIDR block,
  including network ID, broadcast address, usable IP range, and subnet listings.

`)
}

func main() {
	handler := NewCLIHandler()

	if err := handler.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
