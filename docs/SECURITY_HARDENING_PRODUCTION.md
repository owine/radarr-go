# Security Hardening for Production

This guide provides comprehensive security hardening recommendations for Radarr Go production deployments, covering network security, authentication, authorization, container security, and compliance best practices.

## Table of Contents

1. [Overview](#overview)
2. [Network Security](#network-security)
3. [Authentication and Authorization](#authentication-and-authorization)
4. [API Security](#api-security)
5. [Container Security](#container-security)
6. [Database Security](#database-security)
7. [File System Security](#file-system-security)
8. [SSL/TLS Configuration](#ssltls-configuration)
9. [Monitoring and Auditing](#monitoring-and-auditing)
10. [Automated Security Scripts](#automated-security-scripts)

## Overview

Security hardening for Radarr Go involves multiple layers of protection:

### Security Principles

- **Defense in Depth**: Multiple security layers
- **Least Privilege**: Minimal necessary permissions
- **Zero Trust**: Verify all connections and requests
- **Fail Secure**: Secure defaults and failure modes
- **Regular Updates**: Keep all components current

### Security Components

- **Network Security**: Firewall rules and network segmentation
- **Application Security**: Authentication, authorization, input validation
- **Container Security**: Secure images and runtime configuration
- **Data Security**: Encryption at rest and in transit
- **Monitoring**: Security event logging and alerting

## Network Security

### Firewall Configuration

#### iptables Rules

```bash
#!/bin/bash
# scripts/configure-firewall.sh
# Production firewall configuration

set -euo pipefail

log() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"; }

configure_iptables() {
    log "Configuring iptables firewall rules..."

    # Flush existing rules
    iptables -F
    iptables -X
    iptables -t nat -F
    iptables -t nat -X
    iptables -t mangle -F
    iptables -t mangle -X

    # Default policies
    iptables -P INPUT DROP
    iptables -P FORWARD DROP
    iptables -P OUTPUT ACCEPT

    # Allow loopback traffic
    iptables -A INPUT -i lo -j ACCEPT
    iptables -A OUTPUT -o lo -j ACCEPT

    # Allow established and related connections
    iptables -A INPUT -m conntrack --ctstate ESTABLISHED,RELATED -j ACCEPT

    # Allow SSH (change port as needed)
    iptables -A INPUT -p tcp --dport 22 -m conntrack --ctstate NEW,ESTABLISHED -j ACCEPT

    # Allow HTTP/HTTPS for reverse proxy
    iptables -A INPUT -p tcp --dport 80 -m conntrack --ctstate NEW,ESTABLISHED -j ACCEPT
    iptables -A INPUT -p tcp --dport 443 -m conntrack --ctstate NEW,ESTABLISHED -j ACCEPT

    # Allow Radarr application (only from trusted sources)
    # Restrict to specific IP ranges or use reverse proxy
    iptables -A INPUT -p tcp --dport 7878 -s 10.0.0.0/8 -m conntrack --ctstate NEW,ESTABLISHED -j ACCEPT
    iptables -A INPUT -p tcp --dport 7878 -s 172.16.0.0/12 -m conntrack --ctstate NEW,ESTABLISHED -j ACCEPT
    iptables -A INPUT -p tcp --dport 7878 -s 192.168.0.0/16 -m conntrack --ctstate NEW,ESTABLISHED -j ACCEPT

    # Allow database access (only from application servers)
    iptables -A INPUT -p tcp --dport 5432 -s 10.0.0.0/8 -m conntrack --ctstate NEW,ESTABLISHED -j ACCEPT

    # Allow monitoring (restrict to monitoring network)
    iptables -A INPUT -p tcp --dport 9090 -s 10.0.1.0/24 -m conntrack --ctstate NEW,ESTABLISHED -j ACCEPT  # Prometheus
    iptables -A INPUT -p tcp --dport 3000 -s 10.0.1.0/24 -m conntrack --ctstate NEW,ESTABLISHED -j ACCEPT  # Grafana

    # Rate limiting for API access
    iptables -A INPUT -p tcp --dport 7878 -m hashlimit \
        --hashlimit-name api_rate_limit \
        --hashlimit-above 60/min \
        --hashlimit-burst 20 \
        --hashlimit-mode srcip \
        --hashlimit-htable-expire 300000 \
        -j DROP

    # Protection against common attacks

    # Drop invalid packets
    iptables -A INPUT -m conntrack --ctstate INVALID -j DROP

    # Protection against port scanning
    iptables -A INPUT -m recent --name portscan --rcheck --seconds 86400 -j DROP
    iptables -A INPUT -m recent --name portscan --remove
    iptables -A INPUT -p tcp --tcp-flags ALL ACK,RST,SYN,FIN -m recent --name portscan --set -j DROP

    # Protection against SYN flood
    iptables -A INPUT -p tcp --syn -m limit --limit 25/sec --limit-burst 100 -j ACCEPT
    iptables -A INPUT -p tcp --syn -j DROP

    # Protection against ping flood
    iptables -A INPUT -p icmp --icmp-type echo-request -m limit --limit 10/sec -j ACCEPT
    iptables -A INPUT -p icmp --icmp-type echo-request -j DROP

    # Log dropped packets
    iptables -A INPUT -m limit --limit 5/min -j LOG --log-prefix "iptables denied: " --log-level 7

    # Save rules (distribution-specific)
    if command -v iptables-save >/dev/null 2>&1; then
        iptables-save > /etc/iptables/rules.v4 2>/dev/null || \
        iptables-save > /etc/sysconfig/iptables 2>/dev/null || true
    fi

    log "Firewall configuration completed"
}

# Configure UFW (Ubuntu/Debian alternative)
configure_ufw() {
    log "Configuring UFW firewall..."

    # Reset UFW
    ufw --force reset

    # Default policies
    ufw default deny incoming
    ufw default allow outgoing

    # Allow SSH
    ufw allow ssh

    # Allow HTTP/HTTPS
    ufw allow http
    ufw allow https

    # Allow Radarr from private networks only
    ufw allow from 10.0.0.0/8 to any port 7878
    ufw allow from 172.16.0.0/12 to any port 7878
    ufw allow from 192.168.0.0/16 to any port 7878

    # Allow database access from application subnet
    ufw allow from 10.0.0.0/24 to any port 5432

    # Enable rate limiting
    ufw limit ssh
    ufw limit 7878/tcp

    # Enable UFW
    ufw --force enable

    log "UFW configuration completed"
}

# Main execution
case "${1:-iptables}" in
    "iptables") configure_iptables ;;
    "ufw") configure_ufw ;;
    *)
        echo "Usage: $0 {iptables|ufw}"
        exit 1
        ;;
esac
```

#### Network Segmentation

```yaml
# docker-compose with network segmentation
version: '3.8'

services:
  radarr-go:
    image: ghcr.io/radarr/radarr-go:latest
    networks:
      - frontend      # Web traffic
      - backend       # Database access
      - monitoring    # Metrics collection
    # Restrict container capabilities
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - DAC_OVERRIDE
      - SETGID
      - SETUID

  postgres:
    image: postgres:17-alpine
    networks:
      - backend       # Database network only
      - monitoring    # Metrics only
    # No external access

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    networks:
      - frontend      # Web traffic only
    # Web-facing service

networks:
  frontend:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.1.0/24
  backend:
    driver: bridge
    internal: true    # No external access
    ipam:
      config:
        - subnet: 172.20.2.0/24
  monitoring:
    driver: bridge
    internal: true    # No external access
    ipam:
      config:
        - subnet: 172.20.3.0/24
```

### VPN and Secure Tunneling

```bash
#!/bin/bash
# scripts/configure-wireguard.sh
# WireGuard VPN configuration for secure access

set -euo pipefail

INTERFACE="wg0"
SERVER_PRIVATE_KEY=""
SERVER_PUBLIC_KEY=""
ALLOWED_IPS="10.8.0.0/24"

configure_wireguard_server() {
    log "Configuring WireGuard server..."

    # Generate server keys if not provided
    if [ -z "$SERVER_PRIVATE_KEY" ]; then
        SERVER_PRIVATE_KEY=$(wg genkey)
        SERVER_PUBLIC_KEY=$(echo "$SERVER_PRIVATE_KEY" | wg pubkey)
    fi

    # Create WireGuard configuration
    cat > "/etc/wireguard/${INTERFACE}.conf" << EOF
[Interface]
PrivateKey = $SERVER_PRIVATE_KEY
Address = 10.8.0.1/24
ListenPort = 51820
SaveConfig = true

# Enable packet forwarding
PreUp = echo 1 > /proc/sys/net/ipv4/ip_forward
PostUp = iptables -A FORWARD -i $INTERFACE -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE
PostDown = iptables -D FORWARD -i $INTERFACE -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE

# Client configurations will be added here
EOF

    # Set secure permissions
    chmod 600 "/etc/wireguard/${INTERFACE}.conf"

    # Enable and start WireGuard
    systemctl enable wg-quick@${INTERFACE}
    systemctl start wg-quick@${INTERFACE}

    log "WireGuard server configured. Public key: $SERVER_PUBLIC_KEY"
}

add_wireguard_client() {
    local client_name="$1"
    local client_ip="$2"

    # Generate client keys
    local client_private_key=$(wg genkey)
    local client_public_key=$(echo "$client_private_key" | wg pubkey)

    # Add client to server configuration
    cat >> "/etc/wireguard/${INTERFACE}.conf" << EOF

[Peer]
# $client_name
PublicKey = $client_public_key
AllowedIPs = $client_ip/32
EOF

    # Create client configuration
    cat > "${client_name}.conf" << EOF
[Interface]
PrivateKey = $client_private_key
Address = $client_ip/24
DNS = 1.1.1.1, 8.8.8.8

[Peer]
PublicKey = $SERVER_PUBLIC_KEY
Endpoint = YOUR_SERVER_IP:51820
AllowedIPs = 10.8.0.0/24
PersistentKeepalive = 25
EOF

    # Restart WireGuard to apply changes
    systemctl restart wg-quick@${INTERFACE}

    log "Client $client_name added with IP $client_ip"
    log "Client configuration saved to ${client_name}.conf"
}

# Usage example
configure_wireguard_server
add_wireguard_client "admin-laptop" "10.8.0.10"
add_wireguard_client "ops-workstation" "10.8.0.11"
```

## Authentication and Authorization

### Enhanced API Key Management

```go
// Example: Enhanced API key management
package auth

import (
    "crypto/rand"
    "crypto/subtle"
    "encoding/hex"
    "fmt"
    "strings"
    "time"
)

type APIKey struct {
    ID          string    `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null"`
    KeyHash     string    `json:"-" gorm:"not null;unique"` // Never expose the hash
    Permissions []string  `json:"permissions" gorm:"type:json"`
    ExpiresAt   *time.Time `json:"expires_at"`
    LastUsedAt  *time.Time `json:"last_used_at"`
    CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    UserAgent   string    `json:"user_agent"`
    IPWhitelist []string  `json:"ip_whitelist" gorm:"type:json"`
}

type APIKeyService struct {
    repo APIKeyRepository
}

func (s *APIKeyService) GenerateAPIKey(name string, permissions []string, expiresAt *time.Time, ipWhitelist []string) (string, *APIKey, error) {
    // Generate cryptographically secure API key
    keyBytes := make([]byte, 32) // 256 bits
    if _, err := rand.Read(keyBytes); err != nil {
        return "", nil, fmt.Errorf("failed to generate API key: %w", err)
    }

    keyString := hex.EncodeToString(keyBytes)

    // Hash the key for storage (using bcrypt or similar)
    keyHash, err := bcrypt.GenerateFromPassword([]byte(keyString), bcrypt.DefaultCost)
    if err != nil {
        return "", nil, fmt.Errorf("failed to hash API key: %w", err)
    }

    apiKey := &APIKey{
        ID:          generateUUID(),
        Name:        name,
        KeyHash:     string(keyHash),
        Permissions: permissions,
        ExpiresAt:   expiresAt,
        IsActive:    true,
        IPWhitelist: ipWhitelist,
    }

    if err := s.repo.Create(apiKey); err != nil {
        return "", nil, fmt.Errorf("failed to store API key: %w", err)
    }

    return keyString, apiKey, nil
}

func (s *APIKeyService) ValidateAPIKey(keyString, clientIP, userAgent string) (*APIKey, error) {
    // Extract API key from various formats
    keyString = s.extractKeyFromString(keyString)

    if len(keyString) != 64 { // 32 bytes = 64 hex chars
        return nil, fmt.Errorf("invalid API key format")
    }

    // Get all active keys and check against each (prevent timing attacks)
    keys, err := s.repo.GetActiveKeys()
    if err != nil {
        return nil, fmt.Errorf("failed to retrieve API keys: %w", err)
    }

    var validKey *APIKey
    for _, key := range keys {
        if err := bcrypt.CompareHashAndPassword([]byte(key.KeyHash), []byte(keyString)); err == nil {
            validKey = &key
            break
        }
    }

    if validKey == nil {
        return nil, fmt.Errorf("invalid API key")
    }

    // Check expiration
    if validKey.ExpiresAt != nil && time.Now().After(*validKey.ExpiresAt) {
        return nil, fmt.Errorf("API key expired")
    }

    // Check IP whitelist
    if len(validKey.IPWhitelist) > 0 && !s.isIPAllowed(clientIP, validKey.IPWhitelist) {
        return nil, fmt.Errorf("IP not allowed for this API key")
    }

    // Update last used information
    now := time.Now()
    validKey.LastUsedAt = &now
    validKey.UserAgent = userAgent
    s.repo.UpdateLastUsed(validKey.ID, now, userAgent)

    return validKey, nil
}

func (s *APIKeyService) extractKeyFromString(input string) string {
    // Handle various formats: "Bearer token", "token", etc.
    if strings.HasPrefix(input, "Bearer ") {
        return strings.TrimPrefix(input, "Bearer ")
    }
    return input
}

func (s *APIKeyService) isIPAllowed(clientIP string, whitelist []string) bool {
    for _, allowedIP := range whitelist {
        if clientIP == allowedIP {
            return true
        }
        // TODO: Add CIDR support
    }
    return false
}

// Permission checking
func (s *APIKeyService) HasPermission(apiKey *APIKey, permission string) bool {
    for _, p := range apiKey.Permissions {
        if p == "*" || p == permission {
            return true
        }
        // Support wildcard permissions like "movies.*"
        if strings.HasSuffix(p, ".*") {
            prefix := strings.TrimSuffix(p, ".*")
            if strings.HasPrefix(permission, prefix+".") {
                return true
            }
        }
    }
    return false
}
```

### Multi-Factor Authentication (MFA)

```go
// Example: TOTP-based MFA implementation
package auth

import (
    "crypto/rand"
    "encoding/base32"
    "fmt"
    "image/png"
    "strings"
    "time"

    "github.com/pquerna/otp"
    "github.com/pquerna/otp/totp"
)

type MFAService struct {
    issuer string
}

func NewMFAService(issuer string) *MFAService {
    return &MFAService{issuer: issuer}
}

func (s *MFAService) GenerateSecret(accountName string) (*otp.Key, error) {
    key, err := totp.Generate(totp.GenerateOpts{
        Issuer:      s.issuer,
        AccountName: accountName,
        SecretSize:  32,
    })

    if err != nil {
        return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
    }

    return key, nil
}

func (s *MFAService) GenerateQRCode(key *otp.Key) ([]byte, error) {
    img, err := key.Image(256, 256)
    if err != nil {
        return nil, fmt.Errorf("failed to generate QR code: %w", err)
    }

    var buf bytes.Buffer
    if err := png.Encode(&buf, img); err != nil {
        return nil, fmt.Errorf("failed to encode QR code: %w", err)
    }

    return buf.Bytes(), nil
}

func (s *MFAService) ValidateToken(secret, token string) bool {
    return totp.Validate(token, secret)
}

func (s *MFAService) ValidateTokenWithSkew(secret, token string, skew uint) bool {
    now := time.Now()

    // Check current time window
    if totp.Validate(token, secret) {
        return true
    }

    // Check previous and next time windows
    for i := uint(1); i <= skew; i++ {
        // Check previous time window
        pastTime := now.Add(-time.Duration(i) * 30 * time.Second)
        if expectedToken, err := totp.GenerateCode(secret, pastTime); err == nil && expectedToken == token {
            return true
        }

        // Check next time window
        futureTime := now.Add(time.Duration(i) * 30 * time.Second)
        if expectedToken, err := totp.GenerateCode(secret, futureTime); err == nil && expectedToken == token {
            return true
        }
    }

    return false
}

// Backup codes for account recovery
func (s *MFAService) GenerateBackupCodes(count int) ([]string, error) {
    codes := make([]string, count)

    for i := 0; i < count; i++ {
        // Generate 8-character alphanumeric code
        bytes := make([]byte, 5)
        if _, err := rand.Read(bytes); err != nil {
            return nil, fmt.Errorf("failed to generate backup code: %w", err)
        }

        codes[i] = strings.ToLower(base32.StdEncoding.EncodeToString(bytes))[:8]
    }

    return codes, nil
}
```

### Role-Based Access Control (RBAC)

```yaml
# config.yaml - RBAC configuration
auth:
  method: "apikey"

  roles:
    admin:
      permissions:
        - "*"  # Full access
      description: "Full administrative access"

    operator:
      permissions:
        - "movies.*"
        - "queue.*"
        - "history.read"
        - "indexer.read"
        - "downloadclient.read"
        - "system.status"
      description: "Operations access for daily management"

    monitor:
      permissions:
        - "movies.read"
        - "queue.read"
        - "history.read"
        - "system.status"
        - "system.health"
      description: "Read-only monitoring access"

    api_integration:
      permissions:
        - "movies.read"
        - "movies.search"
        - "queue.read"
        - "system.status"
      description: "API integration access"

  # Default permissions for legacy API keys
  default_permissions:
    - "movies.*"
    - "queue.*"
    - "history.*"
    - "system.status"

# Rate limiting per role
rate_limiting:
  admin:
    requests_per_minute: 1000
    burst_size: 100
  operator:
    requests_per_minute: 300
    burst_size: 50
  monitor:
    requests_per_minute: 100
    burst_size: 20
  api_integration:
    requests_per_minute: 200
    burst_size: 30
```

## API Security

### Input Validation and Sanitization

```go
// Example: Comprehensive input validation
package validation

import (
    "fmt"
    "net/url"
    "regexp"
    "strings"
    "time"
)

type Validator struct {
    emailRegex *regexp.Regexp
    pathRegex  *regexp.Regexp
}

func NewValidator() *Validator {
    return &Validator{
        emailRegex: regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
        pathRegex:  regexp.MustCompile(`^[a-zA-Z0-9/_.-]+$`),
    }
}

// Sanitize and validate movie title
func (v *Validator) ValidateMovieTitle(title string) (string, error) {
    // Trim whitespace
    title = strings.TrimSpace(title)

    // Check length
    if len(title) == 0 {
        return "", fmt.Errorf("title cannot be empty")
    }
    if len(title) > 255 {
        return "", fmt.Errorf("title too long (max 255 characters)")
    }

    // Remove potentially dangerous characters
    title = v.sanitizeString(title)

    return title, nil
}

// Validate file paths to prevent directory traversal
func (v *Validator) ValidateFilePath(path string) error {
    // Normalize path
    path = strings.ReplaceAll(path, "\\", "/")

    // Check for directory traversal attempts
    if strings.Contains(path, "..") {
        return fmt.Errorf("path contains directory traversal")
    }

    // Check for absolute paths (should be relative)
    if strings.HasPrefix(path, "/") {
        return fmt.Errorf("absolute paths not allowed")
    }

    // Validate characters
    if !v.pathRegex.MatchString(path) {
        return fmt.Errorf("path contains invalid characters")
    }

    return nil
}

// Validate API pagination parameters
func (v *Validator) ValidatePagination(page, pageSize int) (int, int, error) {
    if page < 1 {
        page = 1
    }
    if page > 10000 { // Prevent excessive page numbers
        return 0, 0, fmt.Errorf("page number too large (max 10000)")
    }

    if pageSize < 1 {
        pageSize = 10
    }
    if pageSize > 1000 { // Prevent excessive page sizes
        return 0, 0, fmt.Errorf("page size too large (max 1000)")
    }

    return page, pageSize, nil
}

// Validate and sanitize search query
func (v *Validator) ValidateSearchQuery(query string) (string, error) {
    query = strings.TrimSpace(query)

    if len(query) == 0 {
        return "", fmt.Errorf("search query cannot be empty")
    }

    if len(query) > 100 {
        return "", fmt.Errorf("search query too long (max 100 characters)")
    }

    // Remove SQL injection attempts
    dangerousPatterns := []string{
        "';",
        "--",
        "/*",
        "*/",
        "xp_",
        "sp_",
        "DROP",
        "DELETE",
        "INSERT",
        "UPDATE",
        "UNION",
        "SELECT",
    }

    upperQuery := strings.ToUpper(query)
    for _, pattern := range dangerousPatterns {
        if strings.Contains(upperQuery, pattern) {
            return "", fmt.Errorf("query contains potentially dangerous content")
        }
    }

    return query, nil
}

// Validate URL for webhooks and notifications
func (v *Validator) ValidateURL(rawURL string) (*url.URL, error) {
    parsedURL, err := url.Parse(rawURL)
    if err != nil {
        return nil, fmt.Errorf("invalid URL format: %w", err)
    }

    // Only allow HTTP and HTTPS
    if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
        return nil, fmt.Errorf("only HTTP and HTTPS URLs are allowed")
    }

    // Prevent localhost/private IP access (SSRF prevention)
    if v.isPrivateHost(parsedURL.Hostname()) {
        return nil, fmt.Errorf("private/local URLs are not allowed")
    }

    return parsedURL, nil
}

func (v *Validator) sanitizeString(input string) string {
    // Remove control characters
    cleaned := strings.Map(func(r rune) rune {
        if r < 32 && r != 9 && r != 10 && r != 13 { // Allow tab, LF, CR
            return -1
        }
        return r
    }, input)

    return cleaned
}

func (v *Validator) isPrivateHost(host string) bool {
    privateHosts := []string{
        "localhost",
        "127.0.0.1",
        "0.0.0.0",
        "::1",
    }

    for _, privateHost := range privateHosts {
        if host == privateHost {
            return true
        }
    }

    // Check private IP ranges (simplified)
    return strings.HasPrefix(host, "10.") ||
           strings.HasPrefix(host, "192.168.") ||
           strings.HasPrefix(host, "172.16.") ||
           strings.HasPrefix(host, "172.17.") ||
           strings.HasPrefix(host, "172.18.") ||
           strings.HasPrefix(host, "172.19.") ||
           strings.HasPrefix(host, "172.20.") ||
           strings.HasPrefix(host, "172.21.") ||
           strings.HasPrefix(host, "172.22.") ||
           strings.HasPrefix(host, "172.23.") ||
           strings.HasPrefix(host, "172.24.") ||
           strings.HasPrefix(host, "172.25.") ||
           strings.HasPrefix(host, "172.26.") ||
           strings.HasPrefix(host, "172.27.") ||
           strings.HasPrefix(host, "172.28.") ||
           strings.HasPrefix(host, "172.29.") ||
           strings.HasPrefix(host, "172.30.") ||
           strings.HasPrefix(host, "172.31.")
}

// Validate time ranges to prevent DoS
func (v *Validator) ValidateTimeRange(start, end time.Time) error {
    if end.Before(start) {
        return fmt.Errorf("end time must be after start time")
    }

    duration := end.Sub(start)
    if duration > 365*24*time.Hour { // Max 1 year range
        return fmt.Errorf("time range too large (max 1 year)")
    }

    return nil
}
```

### Rate Limiting and DDoS Protection

```go
// Example: Advanced rate limiting implementation
package ratelimit

import (
    "context"
    "fmt"
    "net/http"
    "strings"
    "sync"
    "time"

    "golang.org/x/time/rate"
)

type RateLimiter struct {
    limiters map[string]*rate.Limiter
    mu       sync.RWMutex

    // Rate limiting configuration
    globalLimit    rate.Limit
    globalBurst    int
    perIPLimit     rate.Limit
    perIPBurst     int
    perKeyLimit    rate.Limit
    perKeyBurst    int

    // Cleanup
    lastCleanup    time.Time
    cleanupInterval time.Duration
}

func NewRateLimiter() *RateLimiter {
    rl := &RateLimiter{
        limiters:       make(map[string]*rate.Limiter),
        globalLimit:    rate.Limit(1000), // 1000 requests per second globally
        globalBurst:    100,
        perIPLimit:     rate.Limit(10),   // 10 requests per second per IP
        perIPBurst:     20,
        perKeyLimit:    rate.Limit(100),  // 100 requests per second per API key
        perKeyBurst:    50,
        cleanupInterval: 5 * time.Minute,
    }

    // Start cleanup routine
    go rl.cleanup()

    return rl
}

func (rl *RateLimiter) AllowIP(ip string) bool {
    return rl.allow("ip:"+ip, rl.perIPLimit, rl.perIPBurst)
}

func (rl *RateLimiter) AllowAPIKey(apiKey string) bool {
    return rl.allow("key:"+apiKey, rl.perKeyLimit, rl.perKeyBurst)
}

func (rl *RateLimiter) AllowGlobal() bool {
    return rl.allow("global", rl.globalLimit, rl.globalBurst)
}

func (rl *RateLimiter) allow(key string, limit rate.Limit, burst int) bool {
    rl.mu.RLock()
    limiter, exists := rl.limiters[key]
    rl.mu.RUnlock()

    if !exists {
        rl.mu.Lock()
        // Double-check after acquiring write lock
        if limiter, exists = rl.limiters[key]; !exists {
            limiter = rate.NewLimiter(limit, burst)
            rl.limiters[key] = limiter
        }
        rl.mu.Unlock()
    }

    return limiter.Allow()
}

func (rl *RateLimiter) cleanup() {
    ticker := time.NewTicker(rl.cleanupInterval)
    defer ticker.Stop()

    for range ticker.C {
        rl.mu.Lock()
        now := time.Now()

        for key, limiter := range rl.limiters {
            // Remove limiters that haven't been used recently
            if limiter.Tokens() == float64(rl.perIPBurst) &&
               now.Sub(rl.lastCleanup) > rl.cleanupInterval {
                delete(rl.limiters, key)
            }
        }

        rl.lastCleanup = now
        rl.mu.Unlock()
    }
}

// Middleware for HTTP rate limiting
func (rl *RateLimiter) Middleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Check global rate limit first
            if !rl.AllowGlobal() {
                http.Error(w, "Global rate limit exceeded", http.StatusTooManyRequests)
                return
            }

            // Get client IP
            clientIP := rl.getClientIP(r)
            if !rl.AllowIP(clientIP) {
                http.Error(w, "IP rate limit exceeded", http.StatusTooManyRequests)
                return
            }

            // Check API key rate limit if present
            if apiKey := rl.getAPIKey(r); apiKey != "" {
                if !rl.AllowAPIKey(apiKey) {
                    http.Error(w, "API key rate limit exceeded", http.StatusTooManyRequests)
                    return
                }
            }

            next.ServeHTTP(w, r)
        })
    }
}

func (rl *RateLimiter) getClientIP(r *http.Request) string {
    // Check X-Forwarded-For header
    xff := r.Header.Get("X-Forwarded-For")
    if xff != "" {
        ips := strings.Split(xff, ",")
        return strings.TrimSpace(ips[0])
    }

    // Check X-Real-IP header
    if xri := r.Header.Get("X-Real-IP"); xri != "" {
        return xri
    }

    // Fall back to RemoteAddr
    ip := r.RemoteAddr
    if colonIndex := strings.LastIndex(ip, ":"); colonIndex != -1 {
        ip = ip[:colonIndex]
    }

    return ip
}

func (rl *RateLimiter) getAPIKey(r *http.Request) string {
    // Check Authorization header
    if auth := r.Header.Get("Authorization"); auth != "" {
        if strings.HasPrefix(auth, "Bearer ") {
            return strings.TrimPrefix(auth, "Bearer ")
        }
    }

    // Check X-API-Key header
    if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
        return apiKey
    }

    // Check query parameter
    return r.URL.Query().Get("apikey")
}

// Advanced DDoS protection
type DDoSProtection struct {
    requestCounts  map[string]int
    blockList      map[string]time.Time
    mu             sync.RWMutex
    threshold      int
    blockDuration  time.Duration
}

func NewDDoSProtection() *DDoSProtection {
    ddos := &DDoSProtection{
        requestCounts: make(map[string]int),
        blockList:     make(map[string]time.Time),
        threshold:     100, // 100 requests in detection window
        blockDuration: 15 * time.Minute,
    }

    // Start cleanup routine
    go ddos.cleanup()

    return ddos
}

func (ddos *DDoSProtection) IsBlocked(ip string) bool {
    ddos.mu.RLock()
    blockTime, blocked := ddos.blockList[ip]
    ddos.mu.RUnlock()

    if !blocked {
        return false
    }

    // Check if block has expired
    if time.Now().After(blockTime.Add(ddos.blockDuration)) {
        ddos.mu.Lock()
        delete(ddos.blockList, ip)
        ddos.mu.Unlock()
        return false
    }

    return true
}

func (ddos *DDoSProtection) RecordRequest(ip string) {
    if ddos.IsBlocked(ip) {
        return
    }

    ddos.mu.Lock()
    defer ddos.mu.Unlock()

    ddos.requestCounts[ip]++

    // Check if threshold exceeded
    if ddos.requestCounts[ip] >= ddos.threshold {
        ddos.blockList[ip] = time.Now()
        delete(ddos.requestCounts, ip)
    }
}

func (ddos *DDoSProtection) cleanup() {
    ticker := time.NewTicker(1 * time.Minute)
    defer ticker.Stop()

    for range ticker.C {
        ddos.mu.Lock()

        // Reset request counts every minute
        ddos.requestCounts = make(map[string]int)

        // Clean expired blocks
        now := time.Now()
        for ip, blockTime := range ddos.blockList {
            if now.After(blockTime.Add(ddos.blockDuration)) {
                delete(ddos.blockList, ip)
            }
        }

        ddos.mu.Unlock()
    }
}
```

## Container Security

### Secure Container Configuration

```dockerfile
# Dockerfile.secure - Security-hardened container
FROM golang:1.25-alpine AS builder

# Install security updates
RUN apk update && apk upgrade && apk add --no-cache \
    ca-certificates \
    git \
    tzdata \
    && update-ca-certificates

# Create non-root user for build
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build with security flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -trimpath \
    -o radarr ./cmd/radarr

# Final stage - minimal security-focused image
FROM scratch

# Import CA certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd

# Copy application binary
COPY --from=builder --chown=appuser:appuser /build/radarr /radarr

# Copy required files
COPY --from=builder --chown=appuser:appuser /build/migrations /migrations

# Use non-root user
USER appuser

# Set security labels
LABEL \
    org.opencontainers.image.title="Radarr Go" \
    org.opencontainers.image.description="Secure Radarr Go container" \
    org.opencontainers.image.vendor="Radarr Go Team" \
    security.scan="enabled" \
    security.non-root="true"

# Expose port
EXPOSE 7878

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/radarr", "healthcheck"]

# Run as non-root
ENTRYPOINT ["/radarr"]
CMD ["--config", "/config/config.yaml", "--data", "/data"]
```

### Docker Compose Security

```yaml
# docker-compose.secure.yml - Security-hardened compose
version: '3.8'

services:
  radarr-go:
    image: ghcr.io/radarr/radarr-go:latest
    container_name: radarr-go-secure
    restart: unless-stopped

    # Security configurations
    user: "1000:1000"  # Run as non-root user
    read_only: true    # Read-only root filesystem

    # Capability restrictions
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - DAC_OVERRIDE
      - SETGID
      - SETUID

    # Security options
    security_opt:
      - no-new-privileges:true
      - seccomp:unconfined  # or custom seccomp profile
      - apparmor:unconfined # or custom AppArmor profile

    # Resource limits
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '1.0'
        reservations:
          memory: 256M
          cpus: '0.5'

    # Environment variables (use secrets in production)
    environment:
      - RADARR_LOG_LEVEL=info
      - RADARR_SERVER_PORT=7878

    # Volumes with specific mount options
    volumes:
      - radarr_config:/config:rw,noexec,nosuid,nodev
      - radarr_data:/data:rw,noexec,nosuid,nodev
      - movies:/movies:ro,noexec,nosuid,nodev
      - downloads:/downloads:rw,noexec,nosuid,nodev
      # Temporary directories
      - type: tmpfs
        target: /tmp
        tmpfs:
          size: 100M
          mode: 01777
          noexec: true
          nosuid: true
          nodev: true

    # Network configuration
    networks:
      - app-network

    # Health check
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:7878/ping"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 60s

    # Logging configuration
    logging:
      driver: json-file
      options:
        max-size: "10m"
        max-file: "3"
        labels: "service=radarr-go"

  postgres:
    image: postgres:17-alpine
    container_name: radarr-postgres-secure
    restart: unless-stopped

    # Security configurations
    user: "999:999"  # postgres user

    # Capability restrictions
    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - DAC_OVERRIDE
      - FOWNER
      - SETGID
      - SETUID

    security_opt:
      - no-new-privileges:true

    # Environment variables
    environment:
      - POSTGRES_DB=radarr
      - POSTGRES_USER=radarr
      - POSTGRES_PASSWORD_FILE=/run/secrets/postgres_password
      - POSTGRES_INITDB_ARGS=--auth-host=scram-sha-256 --auth-local=scram-sha-256

    # Volumes
    volumes:
      - postgres_data:/var/lib/postgresql/data:rw,noexec,nosuid,nodev
      - ./postgres-init:/docker-entrypoint-initdb.d:ro,noexec,nosuid,nodev

    # Secrets
    secrets:
      - postgres_password

    # Network (backend only)
    networks:
      - db-network

    # Resource limits
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '2.0'

  nginx:
    image: nginx:alpine
    container_name: radarr-nginx-secure
    restart: unless-stopped

    # Security configurations
    user: "nginx:nginx"

    cap_drop:
      - ALL
    cap_add:
      - CHOWN
      - DAC_OVERRIDE
      - SETGID
      - SETUID
      - NET_BIND_SERVICE

    security_opt:
      - no-new-privileges:true

    ports:
      - "80:80"
      - "443:443"

    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro,noexec,nosuid,nodev
      - nginx_cache:/var/cache/nginx:rw,noexec,nosuid,nodev
      - nginx_logs:/var/log/nginx:rw,noexec,nosuid,nodev
      # Read-only tmp for nginx
      - type: tmpfs
        target: /tmp
        tmpfs:
          size: 10M
          noexec: true
          nosuid: true
          nodev: true

    networks:
      - app-network
      - web-network

# Secrets management
secrets:
  postgres_password:
    file: ./secrets/postgres_password.txt

# Network segmentation
networks:
  app-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.1.0/24
  db-network:
    driver: bridge
    internal: true  # No external access
    ipam:
      config:
        - subnet: 172.20.2.0/24
  web-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.3.0/24

# Secure volume configuration
volumes:
  radarr_config:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/radarr/config
  radarr_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/radarr/data
  postgres_data:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/radarr/postgres
  nginx_cache:
    driver: local
  nginx_logs:
    driver: local
```

### Container Security Scanning

```bash
#!/bin/bash
# scripts/security-scan.sh
# Container security scanning and vulnerability assessment

set -euo pipefail

IMAGE_NAME="${1:-ghcr.io/radarr/radarr-go:latest}"
REPORT_DIR="security-reports/$(date +%Y%m%d_%H%M%S)"

mkdir -p "$REPORT_DIR"

log() { echo "[$(date +'%H:%M:%S')] $1"; }

# Scan with multiple tools
scan_with_trivy() {
    log "Scanning with Trivy..."

    if command -v trivy >/dev/null 2>&1; then
        trivy image \
            --format json \
            --output "$REPORT_DIR/trivy-report.json" \
            "$IMAGE_NAME"

        trivy image \
            --format table \
            --output "$REPORT_DIR/trivy-report.txt" \
            "$IMAGE_NAME"

        log "Trivy scan completed"
    else
        log "Trivy not installed, skipping"
    fi
}

scan_with_grype() {
    log "Scanning with Grype..."

    if command -v grype >/dev/null 2>&1; then
        grype -o json "$IMAGE_NAME" > "$REPORT_DIR/grype-report.json"
        grype -o table "$IMAGE_NAME" > "$REPORT_DIR/grype-report.txt"

        log "Grype scan completed"
    else
        log "Grype not installed, skipping"
    fi
}

scan_with_clair() {
    log "Scanning with Clair..."

    if command -v clairctl >/dev/null 2>&1; then
        clairctl analyze "$IMAGE_NAME" > "$REPORT_DIR/clair-report.json"
        log "Clair scan completed"
    else
        log "Clair not installed, skipping"
    fi
}

# Docker security best practices check
check_dockerfile_security() {
    log "Checking Dockerfile security best practices..."

    local dockerfile="${2:-Dockerfile}"

    if [ ! -f "$dockerfile" ]; then
        log "Dockerfile not found at $dockerfile"
        return 1
    fi

    {
        echo "Dockerfile Security Analysis"
        echo "============================"
        echo "File: $dockerfile"
        echo "Date: $(date)"
        echo

        # Check for non-root user
        if grep -q "USER" "$dockerfile"; then
            echo "✓ Non-root user specified"
        else
            echo "⚠ WARNING: No USER instruction found - container runs as root"
        fi

        # Check for COPY --chown usage
        if grep -q "COPY.*--chown" "$dockerfile"; then
            echo "✓ COPY --chown used for proper file ownership"
        else
            echo "⚠ WARNING: Consider using COPY --chown for proper file ownership"
        fi

        # Check for specific base image version
        if grep -q "FROM.*:latest" "$dockerfile"; then
            echo "⚠ WARNING: Using :latest tag - consider pinning to specific version"
        else
            echo "✓ Specific base image version used"
        fi

        # Check for security updates
        if grep -q "apk.*upgrade\|apt.*upgrade\|yum.*update" "$dockerfile"; then
            echo "✓ Security updates applied"
        else
            echo "⚠ WARNING: No explicit security updates found"
        fi

        # Check for CA certificates
        if grep -q "ca-certificates" "$dockerfile"; then
            echo "✓ CA certificates installed"
        else
            echo "⚠ WARNING: CA certificates not explicitly installed"
        fi

        # Check for minimal base image
        if grep -q "FROM scratch\|FROM.*alpine\|FROM.*distroless" "$dockerfile"; then
            echo "✓ Minimal base image used"
        else
            echo "⚠ WARNING: Consider using minimal base image (alpine, distroless, or scratch)"
        fi

    } > "$REPORT_DIR/dockerfile-security.txt"

    log "Dockerfile security check completed"
}

# Runtime security configuration check
check_runtime_security() {
    log "Checking runtime security configuration..."

    {
        echo "Container Runtime Security Check"
        echo "==============================="
        echo "Date: $(date)"
        echo

        # Check running containers
        if docker ps --format "table {{.Names}}\t{{.Image}}\t{{.RunningFor}}" | grep -q radarr; then
            echo "=== Running Radarr Containers ==="
            docker ps --format "table {{.Names}}\t{{.Image}}\t{{.Status}}" | grep radarr
            echo

            # Check container security settings
            for container in $(docker ps --filter "name=radarr" --format "{{.Names}}"); do
                echo "=== Security Settings for $container ==="

                # Check user
                local user=$(docker inspect "$container" --format '{{.Config.User}}')
                if [ -n "$user" ] && [ "$user" != "root" ] && [ "$user" != "0" ]; then
                    echo "✓ Running as non-root user: $user"
                else
                    echo "⚠ WARNING: Running as root user"
                fi

                # Check read-only filesystem
                local readonly=$(docker inspect "$container" --format '{{.HostConfig.ReadonlyRootfs}}')
                if [ "$readonly" = "true" ]; then
                    echo "✓ Read-only root filesystem enabled"
                else
                    echo "⚠ WARNING: Root filesystem is writable"
                fi

                # Check capabilities
                local caps_drop=$(docker inspect "$container" --format '{{.HostConfig.CapDrop}}')
                local caps_add=$(docker inspect "$container" --format '{{.HostConfig.CapAdd}}')

                if [[ "$caps_drop" == *"ALL"* ]]; then
                    echo "✓ All capabilities dropped"
                else
                    echo "⚠ WARNING: Not all capabilities dropped"
                fi

                echo "Capabilities added: $caps_add"
                echo "Capabilities dropped: $caps_drop"

                # Check security options
                local security_opts=$(docker inspect "$container" --format '{{.HostConfig.SecurityOpt}}')
                echo "Security options: $security_opts"

                # Check privileged mode
                local privileged=$(docker inspect "$container" --format '{{.HostConfig.Privileged}}')
                if [ "$privileged" = "false" ]; then
                    echo "✓ Not running in privileged mode"
                else
                    echo "⚠ WARNING: Running in privileged mode"
                fi

                echo
            done
        else
            echo "No running Radarr containers found"
        fi

    } > "$REPORT_DIR/runtime-security.txt"

    log "Runtime security check completed"
}

# Generate security summary
generate_security_summary() {
    log "Generating security summary..."

    {
        echo "Security Scan Summary"
        echo "===================="
        echo "Image: $IMAGE_NAME"
        echo "Scan Date: $(date)"
        echo "Report Directory: $REPORT_DIR"
        echo

        # Count vulnerabilities from Trivy if available
        if [ -f "$REPORT_DIR/trivy-report.json" ]; then
            echo "=== Trivy Vulnerability Summary ==="
            local critical=$(jq -r '.Results[]?.Vulnerabilities[]? | select(.Severity == "CRITICAL") | .VulnerabilityID' "$REPORT_DIR/trivy-report.json" 2>/dev/null | wc -l || echo "0")
            local high=$(jq -r '.Results[]?.Vulnerabilities[]? | select(.Severity == "HIGH") | .VulnerabilityID' "$REPORT_DIR/trivy-report.json" 2>/dev/null | wc -l || echo "0")
            local medium=$(jq -r '.Results[]?.Vulnerabilities[]? | select(.Severity == "MEDIUM") | .VulnerabilityID' "$REPORT_DIR/trivy-report.json" 2>/dev/null | wc -l || echo "0")
            local low=$(jq -r '.Results[]?.Vulnerabilities[]? | select(.Severity == "LOW") | .VulnerabilityID' "$REPORT_DIR/trivy-report.json" 2>/dev/null | wc -l || echo "0")

            echo "Critical: $critical"
            echo "High: $high"
            echo "Medium: $medium"
            echo "Low: $low"
            echo

            if [ "$critical" -gt 0 ] || [ "$high" -gt 0 ]; then
                echo "⚠ ATTENTION: Critical or high severity vulnerabilities found!"
                echo
            fi
        fi

        echo "=== Security Recommendations ==="
        echo "1. Regularly update base images and dependencies"
        echo "2. Use minimal base images (Alpine, distroless, or scratch)"
        echo "3. Run containers as non-root users"
        echo "4. Enable read-only root filesystem"
        echo "5. Drop unnecessary capabilities"
        echo "6. Use security scanning in CI/CD pipeline"
        echo "7. Implement network segmentation"
        echo "8. Monitor containers for runtime security issues"
        echo "9. Use secrets management for sensitive data"
        echo "10. Regular security audits and penetration testing"
        echo

        echo "=== Detailed Reports ==="
        echo "- Trivy Report: $REPORT_DIR/trivy-report.txt"
        echo "- Grype Report: $REPORT_DIR/grype-report.txt"
        echo "- Dockerfile Security: $REPORT_DIR/dockerfile-security.txt"
        echo "- Runtime Security: $REPORT_DIR/runtime-security.txt"

    } > "$REPORT_DIR/security-summary.txt"

    log "Security summary generated: $REPORT_DIR/security-summary.txt"
}

# Main execution
case "${1:-all}" in
    "trivy") scan_with_trivy ;;
    "grype") scan_with_grype ;;
    "clair") scan_with_clair ;;
    "dockerfile") check_dockerfile_security "$@" ;;
    "runtime") check_runtime_security ;;
    "summary") generate_security_summary ;;
    "all"|*)
        scan_with_trivy
        scan_with_grype
        scan_with_clair
        check_dockerfile_security "$@"
        check_runtime_security
        generate_security_summary

        log "Complete security scan finished"
        log "Results available in: $REPORT_DIR"
        cat "$REPORT_DIR/security-summary.txt"
        ;;
esac
```

This security hardening guide provides comprehensive security measures including network security, authentication enhancements, API security, container hardening, and automated security scanning tools. The implementation covers both preventive measures and monitoring capabilities for production deployments.
