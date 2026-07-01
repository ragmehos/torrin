package safeurl

import (
	"net"
	"testing"
)

func TestBlocked(t *testing.T) {
	block := []string{
		"127.0.0.1", "10.0.0.5", "192.168.1.1", "172.16.0.1",
		"169.254.169.254", "::1", "0.0.0.0", "fe80::1", "fc00::1",
	}
	allow := []string{"8.8.8.8", "1.1.1.1", "93.184.216.34", "2606:2800:220:1:248:1893:25c8:1946"}
	for _, s := range block {
		if !Blocked(net.ParseIP(s)) {
			t.Errorf("%s should be blocked", s)
		}
	}
	for _, s := range allow {
		if Blocked(net.ParseIP(s)) {
			t.Errorf("%s should be allowed (public)", s)
		}
	}
}
