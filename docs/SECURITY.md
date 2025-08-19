# Security Guidelines

## Overview

This document outlines the security practices and tools implemented in the GHCP Memory Context Server project to protect against common security vulnerabilities and prevent sensitive data exposure.

## Pre-commit Security Hooks

### Purpose

Pre-commit hooks are automated checks that run before each commit to:
- Prevent secrets and credentials from being committed
- Block sensitive files from being tracked
- Enforce code quality and security standards
- Detect potential security vulnerabilities

### Installation

```bash
# Automated setup
./scripts/setup-precommit.sh

# Manual setup
pip install pre-commit
pre-commit install
pre-commit run --all-files
```

### Security Checks Enabled

#### 1. Secret Detection
- **detect-secrets**: Scans for various types of secrets (API keys, passwords, tokens)
- **gitleaks**: Additional secret detection with custom patterns
- **detect-private-key**: Prevents SSH and other private keys from being committed

#### 2. File Security
- **check-added-large-files**: Blocks files larger than 1MB
- **Large binary detection**: Prevents accidental binary commits
- **Sensitive file patterns**: Blocks `.env`, `.pem`, `.key`, and other sensitive files

#### 3. Code Quality Security
- **gosec**: Go security analyzer for vulnerability detection
- **golangci-lint**: Comprehensive Go linting including security checks
- **go-vet**: Go's built-in static analysis tool

## Sensitive Data Handling

### What NOT to Commit

❌ **Never commit these types of files:**
- API keys and tokens
- Database credentials
- Private keys (SSH, TLS, etc.)
- Environment files (`.env`, `.env.local`)
- Database files (`.db`, `.sqlite`)
- Configuration files with credentials
- JWT secrets
- OAuth client secrets

### Safe Practices

✅ **Do commit:**
- Example/template files (`.env.example`)
- Public configuration
- Documentation
- Test data (non-sensitive)

### Environment Variables

Use environment variables for sensitive configuration:

```go
// Good: Use environment variables
apiKey := os.Getenv("MCP_API_KEY")

// Bad: Hardcoded secrets
apiKey := "sk-1234567890abcdef" // ❌ Never do this
```

### Example Files

Create `.example` files for configuration templates:

```bash
# .env.example (safe to commit)
MCP_API_KEY=your_api_key_here
DATABASE_URL=postgres://user:pass@localhost/dbname  # pragma: allowlist secret

# .env (never commit)
MCP_API_KEY=sk-1234567890abcdef
DATABASE_URL=postgres://user:secretpass@localhost/production  # pragma: allowlist secret
```

## Handling Pre-commit Failures

### Secret Detection Failures

If a secret is detected:

1. **Remove the secret** from your code
2. **Use environment variables** or configuration files
3. **Update the secrets baseline** if it's a false positive:
   ```bash
   detect-secrets scan --baseline .secrets.baseline
   ```

### False Positives

For legitimate strings flagged as secrets:

1. **Add to allowlist** in `.gitleaks.toml`
2. **Update secrets baseline** with verification
3. **Use comments** to mark safe patterns:
   ```go
   // This is safe: example API key format
   exampleKey := "api_key_example_12345"
   ```

### Emergency Commits

In rare cases where you need to bypass hooks:

```bash
# Use with extreme caution
git commit --no-verify -m "Emergency fix"
```

**⚠️ Warning**: Always review and fix security issues before pushing to remote repositories.

## File Patterns

### Blocked by Default

The following patterns are automatically blocked:

```yaml
# Environment files
*.env
.env.*

# Keys and certificates
*.key
*.pem
*.p12
*.pfx

# Database files
*.db
*.sqlite
*.sqlite3

# Sensitive configs
*secret*
*password*
*credential*
```

### Allowed Patterns

These patterns are allowed:

```yaml
# Example and template files
*.example
*.template
*.sample

# Test files
*_test.go
*/testdata/*

# Documentation
*.md
*.txt
```

## CI/CD Integration

Pre-commit hooks also run in CI/CD pipelines to ensure:
- Consistent enforcement across all environments
- Backup protection if local hooks are bypassed
- Security validation on all pull requests

## Security Tools Configuration

### Gitleaks Configuration

Custom patterns for project-specific secrets are defined in `.gitleaks.toml`:
- MCP API keys
- TaskMaster tokens
- Database connection strings

### Detect-Secrets Configuration

Baseline configuration in `.secrets.baseline` includes:
- Plugin configuration for various secret types
- Filters for common false positives
- Allowlist for legitimate patterns

### Golangci-lint Security

Security-focused linters enabled:
- `gosec`: Security vulnerability detection
- `errcheck`: Unchecked error detection
- `staticcheck`: Static analysis for bugs

## Best Practices

1. **Regular Updates**: Keep security tools updated
   ```bash
   pre-commit autoupdate
   ```

2. **Baseline Maintenance**: Regularly review and update secrets baseline
   ```bash
   detect-secrets audit .secrets.baseline
   ```

3. **Team Training**: Ensure all team members understand security practices

4. **Security Reviews**: Include security review in pull request process

5. **Incident Response**: Have a plan for handling accidental secret commits

## Reporting Security Issues

If you discover a security vulnerability:

1. **Do not** create a public issue
2. **Do not** commit the vulnerability details
3. **Contact** the security team privately
4. **Follow** responsible disclosure practices

## Additional Resources

- [OWASP Secure Coding Practices](https://owasp.org/www-project-secure-coding-practices-quick-reference-guide/)
- [Go Security Best Practices](https://github.com/OWASP/Go-SCP)
- [Git Security Best Practices](https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/Git_Cheat_Sheet.md)
