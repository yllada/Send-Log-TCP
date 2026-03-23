# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.6.x   | :white_check_mark: |
| 1.5.x   | :white_check_mark: |
| < 1.5   | :x:                |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, report them via:

1. **Email**: security@y3lcorp.com (preferred)
2. **Private GitHub Advisory**: [Create a private security advisory](https://github.com/Y3LCorp/Send-Log-TCP/security/advisories/new)

### What to Include

Please include the following information:

- **Description** of the vulnerability
- **Steps to reproduce** the issue
- **Potential impact** of the vulnerability
- **Suggested fix** (if you have one)
- **Your contact information** for follow-up questions

### Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Depends on severity (see below)

| Severity | Response Time |
|----------|--------------|
| Critical | 24-48 hours |
| High | 7 days |
| Medium | 30 days |
| Low | 90 days |

## Security Considerations

### Network Security

SendLog Syslog handles network connections and transmits log data. Key security features:

#### TLS/SSL Support
- **TLS 1.2 and 1.3** support for encrypted connections
- **Certificate verification** (configurable)
- **mTLS (Mutual TLS)** for client authentication
- Custom CA certificate support

#### Secure Defaults
- TLS verification is **enabled by default**
- Minimum TLS version is 1.2
- InsecureSkipVerify only when explicitly disabled

### Data Handling

#### Local Storage
- Configuration stored in user's config directory
- File permissions set to `0644` (user read/write)
- No encryption for stored profiles (by design - local config)

#### Sensitive Data
- Certificate paths are stored, not certificate contents
- No credentials are stored by the application
- Connection details are stored locally only

### What We Don't Store

- **No analytics or telemetry** are collected
- **No data sent to third parties**
- **No cloud connectivity** required

## Security Best Practices for Users

### Certificate Management

1. **Use trusted certificates** in production
2. **Rotate certificates** regularly
3. **Store private keys securely** with proper file permissions
4. **Don't disable TLS verification** in production

### Network Security

1. **Use TLS** for all production syslog connections
2. **Verify server certificates** to prevent MITM attacks
3. **Use firewall rules** to restrict syslog server access
4. **Monitor connection logs** for unusual activity

### Operational Security

1. **Keep the application updated** to get security patches
2. **Review connection profiles** periodically
3. **Use separate profiles** for test vs production
4. **Remove unused profiles** and templates

## Vulnerability Disclosure Policy

We follow responsible disclosure practices:

1. **Reporter submits** vulnerability privately
2. **We acknowledge** receipt within 48 hours
3. **We investigate** and validate the issue
4. **We develop** a fix
5. **We coordinate** disclosure timing with reporter
6. **We release** the fix and publish an advisory
7. **Reporter credited** (unless anonymity requested)

## Security Acknowledgments

We thank the following individuals for responsibly disclosing security issues:

*No vulnerabilities reported yet.*

---

## Questions?

For security-related questions that are not vulnerabilities, please open a [Discussion](https://github.com/yllada/Send-Log-TCP/discussions).
