# Security Policy

## Reporting Security Vulnerabilities

We take the security of Environment Manager seriously. If you believe you have found a security vulnerability in our project, please report it to us as described below.


### How to Report a Vulnerability

if you have enabled GitHub Security Advisories, you can report directly through:
- Navigate to the Security tab in this repository
- Click on "Report a vulnerability"

### What to Include in Your Report

Please include the following information to help us understand the nature and scope of the possible issue:

- Type of issue (e.g., buffer overflow, SQL injection, cross-site scripting, etc.)
- Full paths of source file(s) related to the manifestation of the issue
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the issue, including how an attacker might exploit it

### Response Timeline

- You will receive an acknowledgment of your report within **48 hours**
- We will provide a detailed response within **5 business days** indicating the next steps in handling your submission
- We will keep you informed about the progress towards a fix and full announcement
- We may ask for additional information or guidance during the resolution process

## Supported Versions

We release security patches for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

*Note: Please update this table based on your actual versioning scheme*

## Security Update Policy

### Disclosure Process

1. **Initial Assessment**: The security report is received and assigned to a primary handler who coordinates the fix and release process
2. **Validation**: The issue is confirmed and we identify all affected versions
3. **Fix Development**: Fixes are prepared for all supported releases
4. **CVE Assignment**: A CVE (Common Vulnerabilities and Exposures) identifier is requested if applicable
5. **Embargo Period**: We typically set a 7-14 day embargo period depending on severity
6. **Release**: On the embargo date, we will:
   - Push fixes to the repository
   - Publish new releases
   - Announce the vulnerability

### Security Notifications

Security updates will be distributed through:

- GitHub Security Advisories
- Release notes
- Email notifications to registered users (if applicable)
- Project blog/website (if applicable)

## Scope

### In Scope

The following are considered valid security vulnerabilities:

- **Authentication/Authorization bypass** in the application
- **Remote Code Execution (RCE)** vulnerabilities
- **SQL Injection** in database queries
- **Cross-Site Scripting (XSS)** in the frontend
- **Cross-Site Request Forgery (CSRF)** attacks
- **Server-Side Request Forgery (SSRF)** vulnerabilities
- **Privilege escalation** within the application
- **Information disclosure** of sensitive data
- **Denial of Service (DoS)** vulnerabilities in critical paths
- **Insecure cryptographic practices** in the codebase
- **Vulnerable dependencies** with exploitable attack vectors

### Out of Scope

The following are **not** considered security vulnerabilities:

- Vulnerabilities in dependencies without a proven exploit in our application
- Issues that require physical access to the server
- Social engineering attacks
- Denial of Service attacks that require significant resources
- Issues in end-of-life versions
- "Theoretical" vulnerabilities without practical exploitation
- Missing security headers that don't lead to exploitable vulnerabilities
- Self-XSS or scenarios requiring significant user interaction
- Issues related to software or protocols not under our control

## Safe Harbor

We support safe harbor for security researchers who:

- Make a good faith effort to avoid privacy violations, destruction of data, and interruption or degradation of our services
- Only interact with accounts you own or with explicit permission from the account holder
- Do not exploit a vulnerability beyond the minimal amount of testing required to prove its existence
- Report vulnerabilities in accordance with this policy

We will not pursue civil action or initiate a complaint to law enforcement for accidental, good faith violations of this policy.

## Security Best Practices for Users

While using Environment Manager, we recommend:

1. **Keep your installation updated** to the latest version
2. **Use strong authentication** mechanisms
3. **Enable HTTPS** for all production deployments
4. **Regularly review** access logs and audit trails
5. **Follow the principle of least privilege** when configuring access
6. **Regularly backup** your configuration and data
