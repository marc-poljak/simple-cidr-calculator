package main

import (
	"net"
	"testing"
)

func TestValidateCIDR(t *testing.T) {
	tests := []struct {
		name    string
		cidr    string
		wantErr bool
	}{
		{
			name:    "valid CIDR",
			cidr:    "192.168.1.0/24",
			wantErr: false,
		},
		{
			name:    "valid CIDR with /32",
			cidr:    "192.168.1.1/32",
			wantErr: false,
		},
		{
			name:    "valid CIDR with /0",
			cidr:    "0.0.0.0/0",
			wantErr: false,
		},
		{
			name:    "empty CIDR",
			cidr:    "",
			wantErr: true,
		},
		{
			name:    "missing slash",
			cidr:    "192.168.1.0",
			wantErr: true,
		},
		{
			name:    "invalid IP",
			cidr:    "256.256.256.256/24",
			wantErr: true,
		},
		{
			name:    "invalid prefix",
			cidr:    "192.168.1.0/33",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCIDR(tt.cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCIDR() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNetworkInfo_Validate(t *testing.T) {
	validIP := net.ParseIP("192.168.1.0")
	validBroadcast := net.ParseIP("192.168.1.255")
	validMask := net.CIDRMask(24, 32)

	tests := []struct {
		name    string
		network NetworkInfo
		wantErr bool
	}{
		{
			name: "valid network info",
			network: NetworkInfo{
				NetworkID:     validIP,
				BroadcastAddr: validBroadcast,
				SubnetMask:    validMask,
				PrefixLength:  24,
			},
			wantErr: false,
		},
		{
			name: "nil network ID",
			network: NetworkInfo{
				NetworkID:     nil,
				BroadcastAddr: validBroadcast,
				SubnetMask:    validMask,
				PrefixLength:  24,
			},
			wantErr: true,
		},
		{
			name: "nil broadcast address",
			network: NetworkInfo{
				NetworkID:     validIP,
				BroadcastAddr: nil,
				SubnetMask:    validMask,
				PrefixLength:  24,
			},
			wantErr: true,
		},
		{
			name: "nil subnet mask",
			network: NetworkInfo{
				NetworkID:     validIP,
				BroadcastAddr: validBroadcast,
				SubnetMask:    nil,
				PrefixLength:  24,
			},
			wantErr: true,
		},
		{
			name: "invalid prefix length - negative",
			network: NetworkInfo{
				NetworkID:     validIP,
				BroadcastAddr: validBroadcast,
				SubnetMask:    validMask,
				PrefixLength:  -1,
			},
			wantErr: true,
		},
		{
			name: "invalid prefix length - too large",
			network: NetworkInfo{
				NetworkID:     validIP,
				BroadcastAddr: validBroadcast,
				SubnetMask:    validMask,
				PrefixLength:  33,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.network.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("NetworkInfo.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSubnetInfo_Validate(t *testing.T) {
	validIP := net.ParseIP("192.168.1.0")
	validBroadcast := net.ParseIP("192.168.1.127")

	tests := []struct {
		name    string
		subnet  SubnetInfo
		wantErr bool
	}{
		{
			name: "valid subnet info",
			subnet: SubnetInfo{
				NetworkID:     validIP,
				CIDR:          "192.168.1.0/25",
				BroadcastAddr: validBroadcast,
			},
			wantErr: false,
		},
		{
			name: "nil network ID",
			subnet: SubnetInfo{
				NetworkID:     nil,
				CIDR:          "192.168.1.0/25",
				BroadcastAddr: validBroadcast,
			},
			wantErr: true,
		},
		{
			name: "empty CIDR",
			subnet: SubnetInfo{
				NetworkID:     validIP,
				CIDR:          "",
				BroadcastAddr: validBroadcast,
			},
			wantErr: true,
		},
		{
			name: "invalid CIDR",
			subnet: SubnetInfo{
				NetworkID:     validIP,
				CIDR:          "invalid",
				BroadcastAddr: validBroadcast,
			},
			wantErr: true,
		},
		{
			name: "nil broadcast address",
			subnet: SubnetInfo{
				NetworkID:     validIP,
				CIDR:          "192.168.1.0/25",
				BroadcastAddr: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.subnet.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("SubnetInfo.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
