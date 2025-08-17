package main

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// CIDRCalculator handles CIDR parsing and network calculations
type CIDRCalculator struct{}

// NewCIDRCalculator creates a new CIDR calculator instance
func NewCIDRCalculator() *CIDRCalculator {
	return &CIDRCalculator{}
}

// ParseCIDR parses CIDR notation and returns comprehensive network information
func (c *CIDRCalculator) ParseCIDR(cidr string) (*NetworkInfo, error) {
	// Validate input format
	if err := c.validateCIDRFormat(cidr); err != nil {
		return nil, err
	}

	// Parse CIDR using Go's net package
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR notation: %v", err)
	}

	// Ensure we're working with IPv4
	if ip.To4() == nil {
		return nil, fmt.Errorf("IPv6 is not supported, please provide an IPv4 CIDR")
	}

	// Get prefix length
	prefixLength, _ := ipNet.Mask.Size()

	// Calculate network information
	networkInfo := &NetworkInfo{
		Network:      *ipNet,
		NetworkID:    ipNet.IP,
		PrefixLength: prefixLength,
		SubnetMask:   ipNet.Mask,
	}

	// Calculate wildcard mask
	networkInfo.WildcardMask = c.calculateWildcardMask(ipNet.Mask)

	// Calculate broadcast address
	networkInfo.BroadcastAddr = c.calculateBroadcastAddress(ipNet.IP, networkInfo.WildcardMask)

	// Calculate usable IP range and host count (handle edge cases)
	c.calculateUsableRange(networkInfo)

	return networkInfo, nil
}

// validateCIDRFormat performs comprehensive CIDR format validation
func (c *CIDRCalculator) validateCIDRFormat(cidr string) error {
	if cidr == "" {
		return fmt.Errorf("CIDR notation cannot be empty")
	}

	// Check if CIDR contains slash
	if !strings.Contains(cidr, "/") {
		return fmt.Errorf("invalid CIDR notation. Expected format: x.x.x.x/y (e.g., 192.168.1.0/24)")
	}

	// Split IP and prefix
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid CIDR notation. Expected format: x.x.x.x/y (e.g., 192.168.1.0/24)")
	}

	ipStr := parts[0]
	prefixStr := parts[1]

	// Validate IP address format
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return fmt.Errorf("invalid IP address format: %s", ipStr)
	}

	// Ensure IPv4
	if ip.To4() == nil {
		return fmt.Errorf("IPv6 is not supported, please provide an IPv4 address")
	}

	// Validate prefix length
	prefix, err := strconv.Atoi(prefixStr)
	if err != nil {
		return fmt.Errorf("invalid prefix length: %s (must be a number between 0 and 32)", prefixStr)
	}

	if prefix < 0 || prefix > 32 {
		return fmt.Errorf("prefix length must be between 0 and 32, got: %d", prefix)
	}

	return nil
}

// calculateWildcardMask calculates the wildcard mask from subnet mask
func (c *CIDRCalculator) calculateWildcardMask(subnetMask net.IPMask) net.IPMask {
	wildcardMask := make(net.IPMask, len(subnetMask))
	for i, b := range subnetMask {
		wildcardMask[i] = ^b
	}
	return wildcardMask
}

// calculateBroadcastAddress calculates broadcast address using network ID and wildcard mask
func (c *CIDRCalculator) calculateBroadcastAddress(networkID net.IP, wildcardMask net.IPMask) net.IP {
	broadcast := make(net.IP, len(networkID))
	for i := range networkID {
		broadcast[i] = networkID[i] | wildcardMask[i]
	}
	return broadcast
}

// calculateUsableRange calculates first/last usable IPs and total host count
// Handles edge cases for /31 and /32 networks
func (c *CIDRCalculator) calculateUsableRange(info *NetworkInfo) {
	switch info.PrefixLength {
	case 32:
		// /32 is a single host - no usable range for other hosts
		info.FirstUsableIP = info.NetworkID
		info.LastUsableIP = info.NetworkID
		info.TotalHosts = 1
	case 31:
		// /31 is point-to-point link - both IPs are usable
		info.FirstUsableIP = info.NetworkID
		info.LastUsableIP = info.BroadcastAddr
		info.TotalHosts = 2
	default:
		// Standard networks - exclude network and broadcast addresses
		info.FirstUsableIP = c.incrementIP(info.NetworkID)
		info.LastUsableIP = c.decrementIP(info.BroadcastAddr)

		// Calculate total hosts: 2^(32-prefix) - 2 (network and broadcast)
		hostBits := 32 - info.PrefixLength
		if hostBits >= 30 {
			// Handle large networks to avoid overflow
			info.TotalHosts = (1 << uint(hostBits)) - 2
		} else {
			info.TotalHosts = (1 << uint(hostBits)) - 2
		}
	}
}

// incrementIP returns the next IP address
func (c *CIDRCalculator) incrementIP(ip net.IP) net.IP {
	result := make(net.IP, len(ip))
	copy(result, ip)

	for i := len(result) - 1; i >= 0; i-- {
		result[i]++
		if result[i] != 0 {
			break
		}
	}
	return result
}

// decrementIP returns the previous IP address
func (c *CIDRCalculator) decrementIP(ip net.IP) net.IP {
	result := make(net.IP, len(ip))
	copy(result, ip)

	for i := len(result) - 1; i >= 0; i-- {
		if result[i] != 0 {
			result[i]--
			break
		}
		result[i] = 255
	}
	return result
}

// CalculateSubnets generates all possible subnets for the next prefix level
// Implements performance optimization by limiting display for large networks
func (c *CIDRCalculator) CalculateSubnets(network *NetworkInfo) []SubnetInfo {
	// Cannot subnet /32 networks
	if network.PrefixLength >= 32 {
		return []SubnetInfo{}
	}

	nextPrefixLength := network.PrefixLength + 1
	subnetSize := uint32(1) << uint(32-nextPrefixLength)

	// Calculate number of possible subnets
	numSubnets := uint32(1) << uint(nextPrefixLength-network.PrefixLength)

	// Performance optimization: limit display for very large networks
	// For networks larger than /16, limit to first 100 subnets to prevent memory issues
	maxSubnetsToDisplay := uint32(100)
	if network.PrefixLength <= 16 && numSubnets > maxSubnetsToDisplay {
		numSubnets = maxSubnetsToDisplay
	}

	subnets := make([]SubnetInfo, 0, numSubnets)

	// Start with the network ID
	currentNetworkID := make(net.IP, len(network.NetworkID))
	copy(currentNetworkID, network.NetworkID)

	for i := uint32(0); i < numSubnets; i++ {
		// Calculate broadcast address for this subnet
		broadcastAddr := c.calculateSubnetBroadcast(currentNetworkID, nextPrefixLength)

		// Create subnet info
		subnet := SubnetInfo{
			NetworkID:     make(net.IP, len(currentNetworkID)),
			CIDR:          fmt.Sprintf("%s/%d", currentNetworkID.String(), nextPrefixLength),
			BroadcastAddr: broadcastAddr,
		}
		copy(subnet.NetworkID, currentNetworkID)

		subnets = append(subnets, subnet)

		// Move to next subnet by adding subnet size to current network ID
		currentNetworkID = c.addToIP(currentNetworkID, subnetSize)
	}

	return subnets
}

// calculateSubnetBroadcast calculates the broadcast address for a subnet
func (c *CIDRCalculator) calculateSubnetBroadcast(networkID net.IP, prefixLength int) net.IP {
	// Create subnet mask for the given prefix length
	subnetMask := net.CIDRMask(prefixLength, 32)
	wildcardMask := c.calculateWildcardMask(subnetMask)

	// Calculate broadcast: Network ID OR Wildcard Mask
	broadcast := make(net.IP, len(networkID))
	for i := range networkID {
		broadcast[i] = networkID[i] | wildcardMask[i]
	}

	return broadcast
}

// addToIP adds a value to an IP address (used for subnet iteration)
func (c *CIDRCalculator) addToIP(ip net.IP, value uint32) net.IP {
	result := make(net.IP, len(ip))
	copy(result, ip)

	// Convert IP to uint32, add value, convert back
	ipUint32 := uint32(result[0])<<24 + uint32(result[1])<<16 + uint32(result[2])<<8 + uint32(result[3])
	ipUint32 += value

	result[0] = byte(ipUint32 >> 24)
	result[1] = byte(ipUint32 >> 16)
	result[2] = byte(ipUint32 >> 8)
	result[3] = byte(ipUint32)

	return result
}
