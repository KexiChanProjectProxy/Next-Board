package middleware

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// TrustedProxyMiddleware handles X-Forwarded-For headers from trusted proxies
func TrustedProxyMiddleware(trustedProxies []string) gin.HandlerFunc {
	trustedMap := make(map[string]bool)
	for _, proxy := range trustedProxies {
		trustedMap[proxy] = true
	}

	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check if the request is from a trusted proxy
		if isTrustedProxy(clientIP, trustedMap) {
			// Get X-Forwarded-For header
			xff := c.GetHeader("X-Forwarded-For")
			if xff != "" {
				// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
				// Take the first one (leftmost) as the real client IP
				ips := strings.Split(xff, ",")
				if len(ips) > 0 {
					realIP := strings.TrimSpace(ips[0])
					if realIP != "" {
						// Set the real client IP in context
						c.Set("client_ip", realIP)
					}
				}
			}

			// Also check X-Real-IP header as fallback
			if c.GetString("client_ip") == "" {
				xRealIP := c.GetHeader("X-Real-IP")
				if xRealIP != "" {
					c.Set("client_ip", xRealIP)
				}
			}
		}

		// If no forwarded IP found, use the direct client IP
		if c.GetString("client_ip") == "" {
			c.Set("client_ip", clientIP)
		}

		c.Next()
	}
}

// isTrustedProxy checks if the given IP is in the trusted proxy list
func isTrustedProxy(ip string, trustedMap map[string]bool) bool {
	// Handle IPv6 loopback
	if ip == "::1" || ip == "[::1]" {
		return trustedMap["::1"] || trustedMap["127.0.0.1"]
	}

	// Handle IPv4 loopback
	if ip == "127.0.0.1" || strings.HasPrefix(ip, "127.0.0.") {
		return trustedMap["127.0.0.1"]
	}

	// Extract IP from [IP]:port format
	host, _, err := net.SplitHostPort(ip)
	if err == nil {
		ip = host
	}

	// Check exact match
	if trustedMap[ip] {
		return true
	}

	// Check if IP is in trusted CIDR ranges (if needed in the future)
	// For now, just exact match

	return false
}

// GetClientIP is a helper function to get the real client IP from context
func GetClientIP(c *gin.Context) string {
	if ip, exists := c.Get("client_ip"); exists {
		if ipStr, ok := ip.(string); ok {
			return ipStr
		}
	}
	return c.ClientIP()
}
