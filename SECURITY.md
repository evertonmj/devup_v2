# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take the security of DevUp seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### How to Report

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: **evertonmj@gmail.com** (or create a private security advisory on GitHub)

You should receive a response within 48 hours. If for some reason you do not, please follow up via email to ensure we received your original message.

### What to Include

Please include the following information in your report:

- Type of issue (e.g., command injection, path traversal, information disclosure)
- Full paths of source file(s) related to the issue
- Location of the affected source code (tag/branch/commit or direct URL)
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Communication**: We will keep you informed about our progress in addressing the issue
- **Resolution**: We aim to resolve critical vulnerabilities within 7 days
- **Credit**: We will credit you in the release notes (unless you prefer to remain anonymous)

## Security Best Practices for Users

### Configuration Files

- **Never commit** `devup.yaml` files containing sensitive information to public repositories
- Use environment variables for secrets rather than hardcoding them
- Keep `.env` files in `.gitignore`

### Running DevUp

- Review configuration files before running `devup install` or `devup setup`
- Be cautious when using configurations from untrusted sources
- Run DevUp with minimal required permissions

### Reporting Issues

If you discover a security issue in a DevUp configuration (not the tool itself), please still report it so we can add safeguards or documentation to prevent similar issues.

## Security Features

DevUp includes the following security considerations:

- Configuration commands are executed with the user's permissions
- Environment files are created with restricted permissions (0600)
- No sensitive data is logged by default
- Path validation prevents directory traversal attacks

## Acknowledgments

We would like to thank the following individuals for responsibly disclosing security issues:

- *No reports yet*
