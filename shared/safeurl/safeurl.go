package safeurl

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

func Blocked(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified() ||
		ip.Equal(net.IPv4(169, 254, 169, 254))
}

func Dialer(allowLocal bool) func(ctx context.Context, network, addr string) (net.Conn, error) {
	d := &net.Dialer{Timeout: 10 * time.Second}
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if allowLocal {
			return d.DialContext(ctx, network, addr)
		}
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}
		ips, err := net.DefaultResolver.LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, s := range ips {
			if ip := net.ParseIP(s); ip != nil && Blocked(ip) {
				return nil, fmt.Errorf("connection to private address %s blocked", s)
			}
		}
		return d.DialContext(ctx, network, net.JoinHostPort(ips[0], port))
	}
}

func Client(timeout time.Duration) *http.Client {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.DialContext = Dialer(false)
	return &http.Client{Timeout: timeout, Transport: t}
}
