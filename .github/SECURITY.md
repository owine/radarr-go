# Security Policy

## Supported Versions

We currently support the following versions of Radarr Go with security updates:

| Version | Supported          |
| ------- | ------------------ |
| main    | :white_check_mark: |
| develop | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. Please follow these steps to report security issues:

### Private Disclosure Process

1. **Do NOT create a public issue** for security vulnerabilities
2. Send an email to: **security@radarr-go.example.com** (replace with actual email)
3. Include the following information:
   - Description of the vulnerability
   - Steps to reproduce the issue
   - Potential impact
   - Suggested fix (if you have one)

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your report within 24 hours
- **Investigation**: We will investigate and validate the vulnerability within 72 hours
- **Fix Timeline**: Critical vulnerabilities will be patched within 7 days, others within 30 days
- **Disclosure**: We will coordinate with you on public disclosure timing

### Security Features

Radarr Go includes several security features:

- **API Key Authentication**: Optional API key protection for all endpoints
- **Input Validation**: All user inputs are validated and sanitized
- **SQL Injection Protection**: GORM ORM protects against SQL injection
- **Dependency Scanning**: Automated scanning with Gosec and Nancy
- **Container Security**: Non-root container execution
- **CORS Protection**: Configurable cross-origin resource sharing

### Security Best Practices

When deploying Radarr Go:

1. **Use API Keys**: Always enable API key authentication in production
2. **HTTPS**: Use HTTPS in production environments
3. **Database Security**: Use strong database passwords and restrict access
4. **Container Security**: Run containers as non-root user
5. **Network Security**: Restrict network access using firewalls
6. **Updates**: Keep Radarr Go and dependencies updated
7. **Monitoring**: Monitor logs for suspicious activity

### Scope

This security policy applies to:

- Radarr Go application code
- Docker containers and images
- CI/CD pipeline security
- Dependencies and third-party libraries

### Out of Scope

The following are considered out of scope:

- Social engineering attacks
- Physical access to servers
- DDoS attacks
- Issues in third-party services (Docker Hub, GitHub, etc.)

## Security Hall of Fame

We recognize security researchers who help improve Radarr Go security:

<!-- Contributors will be listed here after responsible disclosure -->

## Contact

For security-related questions or concerns:
- Email: security@radarr-go.example.com
- PGP Key: [Link to public key]

Thank you for helping keep Radarr Go secure!
