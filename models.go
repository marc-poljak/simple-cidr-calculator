package main

import (
	"fmt"
	"net"
	"strings"
)

// NetworkInfo represents comprehensive information about a network
type NetworkInfo struct {
	Network       net.IPNet
	NetworkID     net.IP
	BroadcastAddr net.IP
	SubnetMask    net.IPMask
	WildcardMask  net.IPMask
	FirstUsableIP net.IP
	LastUsableIP  net.IP
	TotalHosts    uint32
	PrefixLength  int
}

// SubnetInfo represents information about a subnet
type SubnetInfo struct {
	NetworkID     net.IP
	CIDR          string
	BroadcastAddr net.IP
}

// ValidateCIDR validates CIDR notation format
func ValidateCIDR(cidr string) error {
	if cidr == "" {
		return fmt.Errorf("CIDR notation cannot be empty")
	}

	// Check if CIDR contains slash
	if !strings.Contains(cidr, "/") {
		return fmt.Errorf("invalid CIDR notation. Expected format: x.x.x.x/y")
	}

	// Parse CIDR using Go's net package
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR notation: %v", err)
	}

	return nil
}

// ValidateNetworkInfo validates NetworkInfo struct fields
func (n *NetworkInfo) Validate() error {
	if n.NetworkID == nil {
		return fmt.Errorf("network ID cannot be nil")
	}

	if n.BroadcastAddr == nil {
		return fmt.Errorf("broadcast address cannot be nil")
	}

	if n.SubnetMask == nil {
		return fmt.Errorf("subnet mask cannot be nil")
	}

	if n.PrefixLength < 0 || n.PrefixLength > 32 {
		return fmt.Errorf("prefix length must be between 0 and 32")
	}

	return nil
}

// ValidateSubnetInfo validates SubnetInfo struct fields
func (s *SubnetInfo) Validate() error {
	if s.NetworkID == nil {
		return fmt.Errorf("subnet network ID cannot be nil")
	}

	if s.CIDR == "" {
		return fmt.Errorf("subnet CIDR cannot be empty")
	}

	// Validate CIDR format
	if err := ValidateCIDR(s.CIDR); err != nil {
		return fmt.Errorf("invalid subnet CIDR: %v", err)
	}

	if s.BroadcastAddr == nil {
		return fmt.Errorf("subnet broadcast address cannot be nil")
	}

	return nil
}
