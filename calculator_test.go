package main

import (
	"net"
	"testing"
)

func TestCIDRCalculator_ParseCIDR(t *testing.T) {
	calc := NewCIDRCalculator()

	tests := []struct {
		name    string
		cidr    string
		wantErr bool
		checks  func(*testing.T, *NetworkInfo)
	}{
		{
			name:    "valid /24 network",
			cidr:    "192.168.1.0/24",
			wantErr: false,
			checks: func(t *testing.T, info *NetworkInfo) {
				if info.NetworkID.String() != "192.168.1.0" {
					t.Errorf("Expected network ID 192.168.1.0, got %s", info.NetworkID.String())
				}
				if info.BroadcastAddr.String() != "192.168.1.255" {
					t.Errorf("Expected broadcast 192.168.1.255, got %s", info.BroadcastAddr.String())
				}
				if info.FirstUsableIP.String() != "192.168.1.1" {
					t.Errorf("Expected first usable 192.168.1.1, got %s", info.FirstUsableIP.String())
				}
				if info.LastUsableIP.String() != "192.168.1.254" {
					t.Errorf("Expected last usable 192.168.1.254, got %s", info.LastUsableIP.String())
				}
				if info.TotalHosts != 254 {
					t.Errorf("Expected 254 hosts, got %d", info.TotalHosts)
				}
				if info.PrefixLength != 24 {
					t.Errorf("Expected prefix length 24, got %d", info.PrefixLength)
				}
			},
		},
		{
			name:    "valid /26 network",
			cidr:    "172.21.4.0/26",
			wantErr: false,
			checks: func(t *testing.T, info *NetworkInfo) {
				if info.NetworkID.String() != "172.21.4.0" {
					t.Errorf("Expected network ID 172.21.4.0, got %s", info.NetworkID.String())
				}
				if info.BroadcastAddr.String() != "172.21.4.63" {
					t.Errorf("Expected broadcast 172.21.4.63, got %s", info.BroadcastAddr.String())
				}
				if info.FirstUsableIP.String() != "172.21.4.1" {
					t.Errorf("Expected first usable 172.21.4.1, got %s", info.FirstUsableIP.String())
				}
				if info.LastUsableIP.String() != "172.21.4.62" {
					t.Errorf("Expected last usable 172.21.4.62, got %s", info.LastUsableIP.String())
				}
				if info.TotalHosts != 62 {
					t.Errorf("Expected 62 hosts, got %d", info.TotalHosts)
				}
			},
		},
		{
			name:    "edge case /32 network (single host)",
			cidr:    "192.168.1.1/32",
			wantErr: false,
			checks: func(t *testing.T, info *NetworkInfo) {
				if info.NetworkID.String() != "192.168.1.1" {
					t.Errorf("Expected network ID 192.168.1.1, got %s", info.NetworkID.String())
				}
				if info.BroadcastAddr.String() != "192.168.1.1" {
					t.Errorf("Expected broadcast 192.168.1.1, got %s", info.BroadcastAddr.String())
				}
				if info.FirstUsableIP.String() != "192.168.1.1" {
					t.Errorf("Expected first usable 192.168.1.1, got %s", info.FirstUsableIP.String())
				}
				if info.LastUsableIP.String() != "192.168.1.1" {
					t.Errorf("Expected last usable 192.168.1.1, got %s", info.LastUsableIP.String())
				}
				if info.TotalHosts != 1 {
					t.Errorf("Expected 1 host, got %d", info.TotalHosts)
				}
			},
		},
		{
			name:    "edge case /31 network (point-to-point)",
			cidr:    "10.0.0.0/31",
			wantErr: false,
			checks: func(t *testing.T, info *NetworkInfo) {
				if info.NetworkID.String() != "10.0.0.0" {
					t.Errorf("Expected network ID 10.0.0.0, got %s", info.NetworkID.String())
				}
				if info.BroadcastAddr.String() != "10.0.0.1" {
					t.Errorf("Expected broadcast 10.0.0.1, got %s", info.BroadcastAddr.String())
				}
				if info.FirstUsableIP.String() != "10.0.0.0" {
					t.Errorf("Expected first usable 10.0.0.0, got %s", info.FirstUsableIP.String())
				}
				if info.LastUsableIP.String() != "10.0.0.1" {
					t.Errorf("Expected last usable 10.0.0.1, got %s", info.LastUsableIP.String())
				}
				if info.TotalHosts != 2 {
					t.Errorf("Expected 2 hosts, got %d", info.TotalHosts)
				}
			},
		},
		{
			name:    "edge case /0 network (entire internet)",
			cidr:    "0.0.0.0/0",
			wantErr: false,
			checks: func(t *testing.T, info *NetworkInfo) {
				if info.NetworkID.String() != "0.0.0.0" {
					t.Errorf("Expected network ID 0.0.0.0, got %s", info.NetworkID.String())
				}
				if info.BroadcastAddr.String() != "255.255.255.255" {
					t.Errorf("Expected broadcast 255.255.255.255, got %s", info.BroadcastAddr.String())
				}
				if info.PrefixLength != 0 {
					t.Errorf("Expected prefix length 0, got %d", info.PrefixLength)
				}
			},
		},
		{
			name:    "valid /30 network (4 addresses)",
			cidr:    "192.168.1.0/30",
			wantErr: false,
			checks: func(t *testing.T, info *NetworkInfo) {
				if info.TotalHosts != 2 {
					t.Errorf("Expected 2 hosts, got %d", info.TotalHosts)
				}
				if info.FirstUsableIP.String() != "192.168.1.1" {
					t.Errorf("Expected first usable 192.168.1.1, got %s", info.FirstUsableIP.String())
				}
				if info.LastUsableIP.String() != "192.168.1.2" {
					t.Errorf("Expected last usable 192.168.1.2, got %s", info.LastUsableIP.String())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.ParseCIDR(tt.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCIDR() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checks != nil {
				tt.checks(t, result)
			}
		})
	}
}

func TestCIDRCalculator_ParseCIDR_InvalidInputs(t *testing.T) {
	calc := NewCIDRCalculator()

	tests := []struct {
		name        string
		cidr        string
		expectedErr string
	}{
		{
			name:        "empty CIDR",
			cidr:        "",
			expectedErr: "CIDR notation cannot be empty",
		},
		{
			name:        "missing slash",
			cidr:        "192.168.1.0",
			expectedErr: "invalid CIDR notation. Expected format: x.x.x.x/y",
		},
		{
			name:        "invalid IP address",
			cidr:        "256.256.256.256/24",
			expectedErr: "invalid IP address format",
		},
		{
			name:        "invalid IP address - letters",
			cidr:        "abc.def.ghi.jkl/24",
			expectedErr: "invalid IP address format",
		},
		{
			name:        "invalid prefix - too large",
			cidr:        "192.168.1.0/33",
			expectedErr: "prefix length must be between 0 and 32",
		},
		{
			name:        "invalid prefix - negative",
			cidr:        "192.168.1.0/-1",
			expectedErr: "prefix length must be between 0 and 32",
		},
		{
			name:        "invalid prefix - not a number",
			cidr:        "192.168.1.0/abc",
			expectedErr: "invalid prefix length: abc",
		},
		{
			name:        "multiple slashes",
			cidr:        "192.168.1.0/24/25",
			expectedErr: "invalid CIDR notation. Expected format: x.x.x.x/y",
		},
		{
			name:        "IPv6 address",
			cidr:        "2001:db8::1/64",
			expectedErr: "IPv6 is not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := calc.ParseCIDR(tt.cidr)
			if err == nil {
				t.Errorf("Expected error for input %s, but got none", tt.cidr)
				return
			}
			if !contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestCIDRCalculator_calculateWildcardMask(t *testing.T) {
	calc := NewCIDRCalculator()

	tests := []struct {
		name           string
		subnetMask     net.IPMask
		expectedResult string
	}{
		{
			name:           "/24 wildcard mask",
			subnetMask:     net.CIDRMask(24, 32),
			expectedResult: "0.0.0.255",
		},
		{
			name:           "/26 wildcard mask",
			subnetMask:     net.CIDRMask(26, 32),
			expectedResult: "0.0.0.63",
		},
		{
			name:           "/30 wildcard mask",
			subnetMask:     net.CIDRMask(30, 32),
			expectedResult: "0.0.0.3",
		},
		{
			name:           "/32 wildcard mask",
			subnetMask:     net.CIDRMask(32, 32),
			expectedResult: "0.0.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.calculateWildcardMask(tt.subnetMask)
			resultIP := net.IP(result)
			if resultIP.String() != tt.expectedResult {
				t.Errorf("Expected wildcard mask %s, got %s", tt.expectedResult, resultIP.String())
			}
		})
	}
}

func TestCIDRCalculator_incrementIP(t *testing.T) {
	calc := NewCIDRCalculator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "increment normal IP",
			input:    "192.168.1.1",
			expected: "192.168.1.2",
		},
		{
			name:     "increment with carry",
			input:    "192.168.1.255",
			expected: "192.168.2.0",
		},
		{
			name:     "increment with multiple carries",
			input:    "192.168.255.255",
			expected: "192.169.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := net.ParseIP(tt.input)
			result := calc.incrementIP(input)
			if result.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestCIDRCalculator_decrementIP(t *testing.T) {
	calc := NewCIDRCalculator()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "decrement normal IP",
			input:    "192.168.1.2",
			expected: "192.168.1.1",
		},
		{
			name:     "decrement with borrow",
			input:    "192.168.2.0",
			expected: "192.168.1.255",
		},
		{
			name:     "decrement with multiple borrows",
			input:    "192.169.0.0",
			expected: "192.168.255.255",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := net.ParseIP(tt.input)
			result := calc.decrementIP(input)
			if result.String() != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result.String())
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 1; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}

func TestCIDRCalculator_CalculateSubnets(t *testing.T) {
	calc := NewCIDRCalculator()

	tests := []struct {
		name           string
		cidr           string
		expectedCount  int
		expectedFirst  string
		expectedSecond string
		expectedLast   string
	}{
		{
			name:           "/24 network subnetting to /25",
			cidr:           "192.168.1.0/24",
			expectedCount:  2,
			expectedFirst:  "192.168.1.0/25",
			expectedSecond: "192.168.1.128/25",
			expectedLast:   "192.168.1.128/25",
		},
		{
			name:           "/26 network subnetting to /27",
			cidr:           "172.21.4.0/26",
			expectedCount:  2,
			expectedFirst:  "172.21.4.0/27",
			expectedSecond: "172.21.4.32/27",
			expectedLast:   "172.21.4.32/27",
		},
		{
			name:           "/28 network subnetting to /29",
			cidr:           "10.0.0.0/28",
			expectedCount:  2,
			expectedFirst:  "10.0.0.0/29",
			expectedSecond: "10.0.0.8/29",
			expectedLast:   "10.0.0.8/29",
		},
		{
			name:           "/22 network subnetting to /23",
			cidr:           "192.168.0.0/22",
			expectedCount:  2,
			expectedFirst:  "192.168.0.0/23",
			expectedSecond: "192.168.2.0/23",
			expectedLast:   "192.168.2.0/23",
		},
		{
			name:           "/30 network subnetting to /31",
			cidr:           "192.168.1.0/30",
			expectedCount:  2,
			expectedFirst:  "192.168.1.0/31",
			expectedSecond: "192.168.1.2/31",
			expectedLast:   "192.168.1.2/31",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the network first
			networkInfo, err := calc.ParseCIDR(tt.cidr)
			if err != nil {
				t.Fatalf("Failed to parse CIDR %s: %v", tt.cidr, err)
			}

			// Calculate subnets
			subnets := calc.CalculateSubnets(networkInfo)

			// Check subnet count
			if len(subnets) != tt.expectedCount {
				t.Errorf("Expected %d subnets, got %d", tt.expectedCount, len(subnets))
			}

			if len(subnets) > 0 {
				// Check first subnet
				if subnets[0].CIDR != tt.expectedFirst {
					t.Errorf("Expected first subnet %s, got %s", tt.expectedFirst, subnets[0].CIDR)
				}

				// Check second subnet if it exists
				if len(subnets) > 1 && subnets[1].CIDR != tt.expectedSecond {
					t.Errorf("Expected second subnet %s, got %s", tt.expectedSecond, subnets[1].CIDR)
				}

				// Check last subnet
				lastSubnet := subnets[len(subnets)-1]
				if lastSubnet.CIDR != tt.expectedLast {
					t.Errorf("Expected last subnet %s, got %s", tt.expectedLast, lastSubnet.CIDR)
				}
			}
		})
	}
}

func TestCIDRCalculator_CalculateSubnets_EdgeCases(t *testing.T) {
	calc := NewCIDRCalculator()

	tests := []struct {
		name          string
		cidr          string
		expectedCount int
		description   string
	}{
		{
			name:          "/32 network (cannot subnet)",
			cidr:          "192.168.1.1/32",
			expectedCount: 0,
			description:   "Single host networks cannot be subnetted",
		},
		{
			name:          "/31 network subnetting to /32",
			cidr:          "10.0.0.0/31",
			expectedCount: 2,
			description:   "Point-to-point link can be divided into two /32 hosts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the network first
			networkInfo, err := calc.ParseCIDR(tt.cidr)
			if err != nil {
				t.Fatalf("Failed to parse CIDR %s: %v", tt.cidr, err)
			}

			// Calculate subnets
			subnets := calc.CalculateSubnets(networkInfo)

			// Check subnet count
			if len(subnets) != tt.expectedCount {
				t.Errorf("Expected %d subnets, got %d. %s", tt.expectedCount, len(subnets), tt.description)
			}
		})
	}
}

func TestCIDRCalculator_CalculateSubnets_LargeNetworks(t *testing.T) {
	calc := NewCIDRCalculator()

	tests := []struct {
		name        string
		cidr        string
		maxExpected int
		description string
	}{
		{
			name:        "/16 network (performance limit test)",
			cidr:        "10.0.0.0/16",
			maxExpected: 100,
			description: "Large networks should be limited to prevent memory issues",
		},
		{
			name:        "/8 network (very large network)",
			cidr:        "10.0.0.0/8",
			maxExpected: 100,
			description: "Very large networks should be limited to 100 subnets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the network first
			networkInfo, err := calc.ParseCIDR(tt.cidr)
			if err != nil {
				t.Fatalf("Failed to parse CIDR %s: %v", tt.cidr, err)
			}

			// Calculate subnets
			subnets := calc.CalculateSubnets(networkInfo)

			// Check that subnet count doesn't exceed maximum
			if len(subnets) > tt.maxExpected {
				t.Errorf("Expected at most %d subnets, got %d. %s", tt.maxExpected, len(subnets), tt.description)
			}

			// Verify that we got some subnets (not zero)
			if len(subnets) == 0 {
				t.Errorf("Expected some subnets, got 0")
			}
		})
	}
}

func TestCIDRCalculator_CalculateSubnets_BroadcastCalculation(t *testing.T) {
	calc := NewCIDRCalculator()

	// Test that broadcast addresses are calculated correctly for subnets
	networkInfo, err := calc.ParseCIDR("192.168.1.0/24")
	if err != nil {
		t.Fatalf("Failed to parse CIDR: %v", err)
	}

	subnets := calc.CalculateSubnets(networkInfo)

	if len(subnets) != 2 {
		t.Fatalf("Expected 2 subnets, got %d", len(subnets))
	}

	// First subnet: 192.168.1.0/25 should have broadcast 192.168.1.127
	if subnets[0].BroadcastAddr.String() != "192.168.1.127" {
		t.Errorf("Expected first subnet broadcast 192.168.1.127, got %s", subnets[0].BroadcastAddr.String())
	}

	// Second subnet: 192.168.1.128/25 should have broadcast 192.168.1.255
	if subnets[1].BroadcastAddr.String() != "192.168.1.255" {
		t.Errorf("Expected second subnet broadcast 192.168.1.255, got %s", subnets[1].BroadcastAddr.String())
	}
}

func TestCIDRCalculator_CalculateSubnets_VariousNetworkSizes(t *testing.T) {
	calc := NewCIDRCalculator()

	tests := []struct {
		name           string
		cidr           string
		expectedCount  int
		checkFirstCIDR string
	}{
		{
			name:           "/20 to /21 subnetting",
			cidr:           "172.16.0.0/20",
			expectedCount:  2,
			checkFirstCIDR: "172.16.0.0/21",
		},
		{
			name:           "/25 to /26 subnetting",
			cidr:           "192.168.1.0/25",
			expectedCount:  2,
			checkFirstCIDR: "192.168.1.0/26",
		},
		{
			name:           "/27 to /28 subnetting",
			cidr:           "10.1.1.0/27",
			expectedCount:  2,
			checkFirstCIDR: "10.1.1.0/28",
		},
		{
			name:           "/29 to /30 subnetting",
			cidr:           "172.20.1.0/29",
			expectedCount:  2,
			checkFirstCIDR: "172.20.1.0/30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the network first
			networkInfo, err := calc.ParseCIDR(tt.cidr)
			if err != nil {
				t.Fatalf("Failed to parse CIDR %s: %v", tt.cidr, err)
			}

			// Calculate subnets
			subnets := calc.CalculateSubnets(networkInfo)

			// Check subnet count
			if len(subnets) != tt.expectedCount {
				t.Errorf("Expected %d subnets, got %d", tt.expectedCount, len(subnets))
			}

			// Check first subnet CIDR
			if len(subnets) > 0 && subnets[0].CIDR != tt.checkFirstCIDR {
				t.Errorf("Expected first subnet %s, got %s", tt.checkFirstCIDR, subnets[0].CIDR)
			}

			// Validate all subnet info structures
			for i, subnet := range subnets {
				if err := subnet.Validate(); err != nil {
					t.Errorf("Subnet %d validation failed: %v", i, err)
				}
			}
		})
	}
}
