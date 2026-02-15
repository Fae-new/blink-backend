package validator

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateExecutionURL performs SSRF protection checks before executing a request
func ValidateExecutionURL(urlStr string, allowLocalhost, allowPrivateIPs bool) error {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Only allow http and https
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %s", parsed.Scheme)
	}

	// Extract hostname
	hostname := parsed.Hostname()
	if hostname == "" {
		return fmt.Errorf("URL must contain a hostname")
	}

	// Block localhost and loopback addresses
	if isLocalhost(hostname) {
		if !allowLocalhost {
			return fmt.Errorf("requests to localhost are not allowed")
		}
	}

	// Resolve hostname to IP addresses
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("failed to resolve hostname: %w", err)
	}

	// Check if any resolved IP is private
	for _, ip := range ips {
		if isPrivateIP(ip) {
			if !allowPrivateIPs {
				return fmt.Errorf("requests to private IP ranges are not allowed: %s", ip.String())
			}
		}
	}

	return nil
}

func isLocalhost(hostname string) bool {
	hostname = strings.ToLower(hostname)
	return hostname == "localhost" ||
		hostname == "127.0.0.1" ||
		hostname == "::1" ||
		strings.HasPrefix(hostname, "127.") ||
		hostname == "0.0.0.0" ||
		hostname == "[::]"
}

func isPrivateIP(ip net.IP) bool {
	// Check for loopback
	if ip.IsLoopback() {
		return true
	}

	// Check for link-local
	if ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	// Define private IP ranges
	privateRanges := []string{
		"10.0.0.0/8",     // Private network
		"172.16.0.0/12",  // Private network
		"192.168.0.0/16", // Private network
		"169.254.0.0/16", // Link-local
		"fc00::/7",       // Unique local address (IPv6)
		"fe80::/10",      // Link-local (IPv6)
		"::1/128",        // Loopback (IPv6)
		"169.254.0.0/16", // AWS metadata
		"100.64.0.0/10",  // Shared address space (CGNAT)
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	// Block cloud metadata IPs
	metadataIPs := []string{
		"169.254.169.254", // AWS, Azure, GCP metadata
		"fd00:ec2::254",   // AWS IMDSv2 IPv6
	}

	ipStr := ip.String()
	for _, metaIP := range metadataIPs {
		if ipStr == metaIP {
			return true
		}
	}

	return false
}
