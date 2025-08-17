# 🌐 Simple CIDR Calculator

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/marc-poljak/simple-cidr-calculator)

A clean, fast command-line tool written in Go that calculates and displays comprehensive subnet information for CIDR notation inputs. Perfect for network engineers, system administrators, and anyone working with IP subnetting.

## ✨ Features

- 🔍 **CIDR Parsing**: Parse and validate CIDR notation (e.g., 192.168.1.0/24)
- 📊 **Network Information**: Display network ID, broadcast address, subnet mask, and wildcard mask
- 🏠 **Host Information**: Show first/last usable IP addresses and total host count
- 🔀 **Subnet Analysis**: Calculate and list all possible subnets for the next prefix level
- 📄 **Multiple Output Formats**: Support for console text output and HTML file generation
- ⚡ **Edge Case Handling**: Proper handling of /31 (point-to-point) and /32 (host route) networks
- 🌍 **Cross-Platform**: Works on Linux, macOS, and Windows

## 🚀 Installation

### Prerequisites

- Go 1.19 or later

### Build from Source

1. Clone the repository:
```bash
git clone https://github.com/marc-poljak/simple-cidr-calculator.git
cd simple-cidr-calculator
```

2. Build the binary:
```bash
go build -o cidr-calc
```

3. (Optional) Install globally:
```bash
go install
```

### Using Make

If you have Make installed, you can use the provided Makefile:

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean

# Run tests
make test
```

## 📖 Usage

### Basic Usage

```bash
cidr-calc <CIDR>
```

### Command Line Options

```
Usage:
  cidr-calc [OPTIONS] <CIDR>

Arguments:
  CIDR                 Network in CIDR notation (e.g., 192.168.1.0/24)

Options:
  -o, --output FILE    Save output to specified file
  -h, --html          Generate HTML formatted output
  --help              Show help message
```

### Examples

#### Basic Network Analysis
```bash
cidr-calc 192.168.1.0/24
```

Output:
```
Network Information:
  CIDR:           192.168.1.0/24
  Network ID:     192.168.1.0
  Broadcast:      192.168.1.255
  Subnet Mask:    255.255.255.0
  Wildcard Mask:  0.0.0.255

Host Information:
  First Usable:   192.168.1.1
  Last Usable:    192.168.1.254
  Total Hosts:    254

Subnet Information:
  Possible /25 Subnets: 2

  Subnet List:
    192.168.1.0/25     (192.168.1.0 - 192.168.1.127)
    192.168.1.128/25   (192.168.1.128 - 192.168.1.255)
```

#### Save to Text File
```bash
cidr-calc -o network-report.txt 172.16.0.0/16
```

#### Generate HTML Report
```bash
cidr-calc --html -o network-report.html 10.0.0.0/8
```

#### Edge Cases

**Point-to-Point Link (/31)**:
```bash
cidr-calc 192.168.1.0/31
```

**Host Route (/32)**:
```bash
cidr-calc 192.168.1.1/32
```

## 📋 Output Formats

### Text Output

The default text output provides a clean, aligned format suitable for terminal viewing and text file storage.

### HTML Output

HTML output generates a professional-looking report with:
- Responsive design for mobile and desktop viewing
- CSS styling with gradient headers and clean tables
- Collapsible sections for large subnet lists
- Print-friendly formatting
- Self-contained file with embedded CSS

## 🧮 Subnet Calculation Logic

The tool calculates subnets by adding exactly one bit to the network prefix, creating two equal-sized subnets that together comprise the original network:

- **Input**: 192.168.1.0/24 → **Output**: Two /25 subnets (each with 128 addresses)
- **Input**: 10.0.0.0/16 → **Output**: Two /17 subnets (each with 32,768 addresses)
- **Input**: 172.16.0.0/20 → **Output**: Two /21 subnets (each with 2,048 addresses)

This approach shows the most common subnetting scenario - dividing a network into two equal halves. Each resulting subnet has exactly half the address space of the original network.

## 🔧 Supported Network Types

- **Standard Networks**: /8, /16, /24, /28, etc.
- **Point-to-Point Links**: /31 networks (RFC 3021)
- **Host Routes**: /32 networks (single host)
- **Large Networks**: Efficiently handles /8 networks with performance optimizations

## ⚠️ Error Handling

The tool provides clear error messages for common issues:

- Invalid CIDR format
- Invalid IP addresses
- Invalid prefix lengths
- File writing permissions
- Flag combination errors

## ⚡ Performance

- Optimized for large networks (e.g., /8 networks)
- Efficient memory usage
- Fast startup time
- Subnet listing limited for very large networks to maintain performance

## 🛠️ Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...
```

### Building for Multiple Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o cidr-calc-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o cidr-calc-macos

# Windows
GOOS=windows GOARCH=amd64 go build -o cidr-calc-windows.exe
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🐛 Support

For issues and questions, please [create an issue](https://github.com/marc-poljak/simple-cidr-calculator/issues) in the repository.

## 🙏 Credits

Enhanced and improved with assistance from [Kiro AI](https://kiro.ai) - an AI-powered development assistant that helped optimize the codebase, improve documentation, and add comprehensive testing.