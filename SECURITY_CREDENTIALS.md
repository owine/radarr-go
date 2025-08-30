# Security and Credential Management

This document outlines security best practices and credential management for Radarr Go.

## 🔐 Credential Security

### Environment Variables (Recommended)

All sensitive configuration should use environment variables instead of hardcoded values:

```yaml
# ✅ SECURE - Using environment variables
database:
  password: "${RADARR_DATABASE_PASSWORD:-}"

auth:
  api_key: "${RADARR_AUTH_API_KEY:-}"

tmdb:
  api_key: "${RADARR_TMDB_API_KEY:-}"
```

```bash
# Set environment variables
export RADARR_DATABASE_PASSWORD="$(openssl rand -base64 32)"
export RADARR_AUTH_API_KEY="$(openssl rand -base64 48)"
export RADARR_TMDB_API_KEY="your_actual_tmdb_api_key"
```

### ❌ What NOT to Do

Never commit these patterns to version control:

```yaml
# ❌ INSECURE - Hardcoded credentials
database:
  password: "password123"

auth:
  api_key: "my-secret-key"
```

## 🛡️ Security Best Practices

### 1. Strong Password Generation

Generate cryptographically secure passwords:

```bash
# Generate 32-character password
openssl rand -base64 32 | tr -d "=+/" | cut -c1-32

# Alternative method
cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1
```

### 2. API Key Security

- **Length**: Minimum 32 characters for API keys
- **Complexity**: Use mixed alphanumeric characters
- **Rotation**: Rotate API keys regularly
- **Scope**: Use different keys for different purposes

```bash
# Generate secure API key
openssl rand -hex 32
```

### 3. Database Credentials

- Use dedicated database users with minimal privileges
- Avoid using database admin accounts
- Enable SSL/TLS for database connections
- Regularly rotate database passwords

## 🐳 Docker Security

### Secure Docker Configuration

```yaml
# docker-compose.yml
version: '3.8'
services:
  radarr-go:
    image: ghcr.io/radarr/radarr-go:latest
    environment:
      - RADARR_DATABASE_PASSWORD_FILE=/run/secrets/db_password
      - RADARR_AUTH_API_KEY_FILE=/run/secrets/api_key
    secrets:
      - db_password
      - api_key

secrets:
  db_password:
    external: true
  api_key:
    external: true
```

### Docker Secrets

```bash
# Create Docker secrets
echo "your_secure_db_password" | docker secret create db_password -
echo "your_secure_api_key" | docker secret create api_key -
```

## 🔄 Development Environment

### Secure Development Setup

The development setup script automatically generates secure credentials:

```bash
# Run development setup (generates random credentials)
./scripts/dev-setup.sh

# Override with your own credentials
export RADARR_DEV_DB_PASSWORD="your_dev_password"
export RADARR_DEV_API_KEY="your_dev_api_key"
```

### Test Credentials

Test environments should use:

```bash
# Set test-specific credentials
export RADARR_TEST_DB_PASSWORD="$(openssl rand -base64 24)"
export RADARR_TEST_API_KEY="$(openssl rand -hex 32)"
```

## 📊 Security Scanning

### GitGuardian Integration

This repository includes GitGuardian configuration (`.gitguardian.yaml`) that:

- ✅ Allows environment variable patterns
- ✅ Ignores documentation examples
- ✅ Excludes test files with mock data
- ❌ Detects real hardcoded secrets

### CI/CD Security Checks

The following security checks run automatically:

```bash
# Security vulnerability scanning
govulncheck ./...

# Credential scanning
gitguardian scan --recursive .

# Dependency scanning
nancy audit go.sum
```

## 🚨 Incident Response

### If Credentials Are Compromised

1. **Immediate Actions**:
   - Rotate all affected credentials immediately
   - Review access logs for unauthorized usage
   - Update all configuration files and deployments

2. **Investigation**:
   - Determine scope of exposure
   - Check if credentials were used maliciously
   - Document the incident

3. **Prevention**:
   - Implement additional monitoring
   - Review security practices
   - Update documentation and training

## 📋 Security Checklist

### Production Deployment

- [ ] All passwords use environment variables
- [ ] API keys are minimum 32 characters
- [ ] Database users have minimal privileges
- [ ] SSL/TLS enabled for all connections
- [ ] Regular credential rotation schedule
- [ ] Security monitoring configured
- [ ] Backup encryption enabled
- [ ] Access logs retention configured

### Development

- [ ] No hardcoded credentials in code
- [ ] Development credentials are unique
- [ ] Test databases are isolated
- [ ] Pre-commit hooks enabled
- [ ] Security scanning in CI/CD

## 📚 Additional Resources

- [OWASP Application Security Verification Standard](https://owasp.org/www-project-application-security-verification-standard/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [GitGuardian Best Practices](https://docs.gitguardian.com/secrets-detection/detectors/generics/generic_password)
- [Database Security Best Practices](https://owasp.org/www-project-cheat-sheets/cheatsheets/Database_Security_Cheat_Sheet.html)

## 🤝 Reporting Security Issues

If you discover a security vulnerability, please report it responsibly:

1. **DO NOT** create a public GitHub issue
2. Email security concerns to: [security@radarr.video]
3. Include detailed information about the vulnerability
4. Allow time for the issue to be addressed before disclosure

We take security seriously and appreciate responsible disclosure.
