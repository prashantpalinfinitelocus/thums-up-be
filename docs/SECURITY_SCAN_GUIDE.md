# Security Scan Execution Guide

**Document Version:** 1.0  
**Last Updated:** January 2025  
**Application:** Thums Up Backend API

This guide provides step-by-step instructions for running all required security scans.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [CodeQL Scan](#codeql-scan)
3. [Dependabot Scan](#dependabot-scan)
4. [Code Scanning](#code-scanning)
5. [Secret Scanning](#secret-scanning)
6. [DAST/StackHawk Scan](#daststackhawk-scan)
7. [Local Security Scans](#local-security-scans)
8. [Interpreting Results](#interpreting-results)

---

## Prerequisites

### Required Tools
- **Git** - Version control
- **GitHub Account** - For GitHub-based scans
- **Docker** (optional) - For local CodeQL scans
- **Go 1.23+** - For local scans
- **StackHawk Account** - For DAST scanning

### GitHub Repository Setup
- Repository must be on GitHub (GitHub.com or GitHub Enterprise)
- GitHub Actions enabled
- Appropriate permissions for security scanning

### Install Local Tools
```bash
# Install gosec (Go security scanner)
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Install govulncheck (Go vulnerability checker)
go install golang.org/x/vuln/cmd/govulncheck@latest

# Install CodeQL CLI (optional, for local scans)
# Download from: https://github.com/github/codeql-cli-binaries/releases
```

---

## CodeQL Scan

CodeQL is GitHub's semantic code analysis engine that finds security vulnerabilities.

### Option 1: GitHub Actions (Recommended)

#### Step 1: Create Workflow File
Create `.github/workflows/codeql.yml` (see provided file in this guide)

#### Step 2: Enable CodeQL on GitHub
1. Go to your repository on GitHub
2. Navigate to **Settings** → **Code security and analysis**
3. Under **Code scanning**, click **Set up** next to CodeQL analysis
4. Select **Go** as the language
5. Choose workflow file: `.github/workflows/codeql.yml`
6. Click **Start commit** and commit the workflow

#### Step 3: View Results
1. Go to **Security** tab in your repository
2. Click **Code scanning** in the left sidebar
3. Results will appear after the workflow runs (usually 5-10 minutes)

### Option 2: Local CodeQL Scan

#### Step 1: Install CodeQL CLI
```bash
# Download CodeQL CLI
wget https://github.com/github/codeql-cli-binaries/releases/latest/download/codeql-bundle-linux64.tar.gz
tar -xzf codeql-bundle-linux64.tar.gz
export PATH=$PATH:$(pwd)/codeql
```

#### Step 2: Create CodeQL Database
```bash
cd /Users/prashantpal/coke/thums-up-be
codeql database create codeql-database --language=go --source-root=.
```

#### Step 3: Run Analysis
```bash
codeql database analyze codeql-database \
  --format=sarif-latest \
  --output=codeql-results.sarif \
  codeql-go-queries
```

#### Step 4: Upload Results (Optional)
```bash
# Upload to GitHub Code Scanning
gh codeql upload-results codeql-results.sarif \
  --ref=main \
  --commit=$(git rev-parse HEAD)
```

### CodeQL Configuration

The workflow file (`.github/workflows/codeql.yml`) is configured to:
- Scan on push to main branch
- Scan on pull requests
- Scan on schedule (weekly)
- Use Go language pack
- Upload results to GitHub Security tab

---

## Dependabot Scan

Dependabot automatically scans dependencies for known vulnerabilities.

### Option 1: GitHub Dependabot (Recommended)

#### Step 1: Enable Dependabot Alerts
1. Go to repository **Settings** → **Code security and analysis**
2. Under **Dependabot alerts**, click **Enable**
3. GitHub will automatically scan your dependencies

#### Step 2: Configure Dependabot
Create `.github/dependabot.yml` (see provided file in this guide)

#### Step 3: View Results
1. Go to **Security** tab
2. Click **Dependabot** in the left sidebar
3. View alerts and update recommendations

### Option 2: Manual Dependency Check

#### Using govulncheck (Go Official)
```bash
cd /Users/prashantpal/coke/thums-up-be

# Check for vulnerabilities
govulncheck ./...

# Check specific module
govulncheck -mod ./...

# Generate JSON report
govulncheck -json ./... > vuln-report.json
```

#### Using go list
```bash
# List all dependencies
go list -m all

# Check for updates
go list -m -u all

# Update dependencies
go get -u ./...
go mod tidy
```

### Dependabot Configuration

The `.github/dependabot.yml` file configures:
- Go module updates (daily)
- Docker updates (weekly)
- GitHub Actions updates (weekly)

---

## Code Scanning

GitHub Advanced Security provides comprehensive code scanning beyond CodeQL.

### Step 1: Enable Advanced Security
1. Go to repository **Settings** → **Code security and analysis**
2. Under **Code scanning**, ensure **CodeQL analysis** is enabled
3. Under **Secret scanning**, click **Enable** (see Secret Scanning section)

### Step 2: Configure Additional Scanners
You can add third-party scanners via GitHub Actions:

#### Example: SonarCloud Integration
```yaml
# .github/workflows/sonarcloud.yml
name: SonarCloud Scan
on:
  push:
    branches: [main]
  pull_request:
    types: [opened, synchronize, reopened]
jobs:
  sonarcloud:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: SonarCloud Scan
        uses: SonarSource/sonarcloud-github-action@master
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          SONAR_TOKEN: ${{ secrets.SONAR_TOKEN }}
```

### Step 3: View Results
1. Go to **Security** tab
2. Click **Code scanning** in the left sidebar
3. All scan results from different tools appear here

---

## Secret Scanning

GitHub automatically scans for secrets in your code.

### Step 1: Enable Secret Scanning
1. Go to repository **Settings** → **Code security and analysis**
2. Under **Secret scanning**, click **Enable**
3. GitHub will scan all commits automatically

### Step 2: Configure Custom Patterns (Optional)
1. Go to **Settings** → **Security** → **Secret scanning**
2. Click **New pattern**
3. Add custom secret patterns if needed

### Step 3: View Results
1. Go to **Security** tab
2. Click **Secret scanning** in the left sidebar
3. Review detected secrets

### Step 4: Remediate Secrets
If secrets are found:
1. **Rotate the secret immediately** - Generate new credentials
2. **Remove from code** - Delete the secret from codebase
3. **Update references** - Update all places using the secret
4. **Use Secret Manager** - Store secrets in Google Secret Manager (for GCP)

### Local Secret Scanning

#### Using GitGuardian CLI
```bash
# Install GitGuardian CLI
pip install ggshield

# Scan repository
ggshield scan repo /Users/prashantpal/coke/thums-up-be

# Scan specific files
ggshield scan path /Users/prashantpal/coke/thums-up-be/config/

# Scan git history
ggshield scan commit-range HEAD~10..HEAD
```

#### Using TruffleHog
```bash
# Install TruffleHog
docker run -it -v "$PWD:/pwd" trufflesecurity/trufflehog:latest \
  github --repo=https://github.com/your-org/your-repo
```

---

## DAST/StackHawk Scan

StackHawk performs dynamic application security testing (DAST) by testing the running application.

### Step 1: Create StackHawk Account
1. Go to https://auth.stackhawk.com/login
2. Sign up or log in
3. Create a new application

### Step 2: Install StackHawk CLI
```bash
# Install via Homebrew (macOS)
brew install stackhawk/homebrew-tap/hawk

# Or download from: https://docs.stackhawk.com/hawk-cli/installation
```

### Step 3: Configure StackHawk
Create `stackhawk.yml` in the project root:

```yaml
app:
  applicationId: thums-up-backend
  env: Production
  host: https://tccc-tja-test-cloudrun-backend-<hash>-<region>.a.run.app
  openApiSpec: ./docs/swagger.yaml
authentication:
  type: Bearer
  token: <your-test-jwt-token>
```

### Step 4: Start Application
```bash
# Make sure your application is running
cd /Users/prashantpal/coke/thums-up-be
make run

# Or if using Docker
docker-compose up -d
```

### Step 5: Run StackHawk Scan
```bash
# Run scan
hawk scan

# Run with specific configuration
hawk scan --config stackhawk.yml

# Run and save results
hawk scan --output stackhawk-results.json
```

### Step 6: View Results
1. Log in to StackHawk dashboard: https://app.stackhawk.com
2. Navigate to your application
3. View scan results, vulnerabilities, and recommendations

### Step 7: Integrate with CI/CD (Optional)
Add to `.github/workflows/stackhawk.yml` (see provided file)

### StackHawk Configuration Options

#### For Development Environment
```yaml
app:
  applicationId: thums-up-backend-dev
  env: Development
  host: http://localhost:8080
  openApiSpec: ./docs/swagger.yaml
authentication:
  type: Bearer
  token: <dev-jwt-token>
```

#### For Staging Environment
```yaml
app:
  applicationId: thums-up-backend-staging
  env: Staging
  host: https://<staging-url>.run.app
  openApiSpec: ./docs/swagger.yaml
authentication:
  type: Bearer
  token: <staging-jwt-token>
```

---

## Local Security Scans

### Using gosec (Go Security Checker)

```bash
cd /Users/prashantpal/coke/thums-up-be

# Run gosec
gosec ./...

# Generate JSON report
gosec -fmt json -out gosec-report.json ./...

# Generate HTML report
gosec -fmt json -out gosec-report.json ./...
# Then convert to HTML using a tool or script

# Include tests
gosec -tests ./...

# Exclude specific rules
gosec -exclude-dir=vendor,node_modules ./...
```

### Using govulncheck

```bash
# Check for known vulnerabilities
govulncheck ./...

# Check with verbose output
govulncheck -v ./...

# Generate JSON report
govulncheck -json ./... > govulncheck-report.json
```

### Using Makefile

```bash
# Run security checks (uses gosec)
make security

# Run all checks
make ci
```

---

## Interpreting Results

### CodeQL Results

**Severity Levels:**
- **Error** - Critical security issue
- **Warning** - Potential security issue
- **Note** - Informational

**Common Findings:**
- SQL injection vulnerabilities
- XSS vulnerabilities
- Insecure deserialization
- Hardcoded secrets
- Weak cryptography

### Dependabot Results

**Severity Levels:**
- **Critical** - Immediate action required
- **High** - Fix as soon as possible
- **Moderate** - Fix when convenient
- **Low** - Optional fix

**Actions:**
- Click on alert to see details
- Click **Create Dependabot security update** to auto-fix
- Or manually update dependency

### Secret Scanning Results

**Actions Required:**
1. **Rotate secret immediately**
2. **Remove from code**
3. **Mark as resolved** in GitHub

### StackHawk Results

**OWASP Top 10 Categories:**
- A01:2021 – Broken Access Control
- A02:2021 – Cryptographic Failures
- A03:2021 – Injection
- A04:2021 – Insecure Design
- A05:2021 – Security Misconfiguration
- A06:2021 – Vulnerable Components
- A07:2021 – Authentication Failures
- A08:2021 – Software and Data Integrity Failures
- A09:2021 – Security Logging Failures
- A10:2021 – Server-Side Request Forgery

---

## Quick Start Checklist

### Initial Setup
- [ ] Enable GitHub Actions
- [ ] Enable Dependabot alerts
- [ ] Enable Secret scanning
- [ ] Create StackHawk account
- [ ] Install local tools (gosec, govulncheck)

### Run Scans
- [ ] CodeQL scan (via GitHub Actions)
- [ ] Dependabot scan (automatic)
- [ ] Code scanning (via GitHub Actions)
- [ ] Secret scanning (automatic)
- [ ] StackHawk DAST scan (manual/CI)

### Review Results
- [ ] Review CodeQL findings
- [ ] Review Dependabot alerts
- [ ] Review secret scanning results
- [ ] Review StackHawk vulnerabilities
- [ ] Document findings in SECURITY_ASSESSMENT.md

### Remediation
- [ ] Fix critical issues
- [ ] Update vulnerable dependencies
- [ ] Rotate exposed secrets
- [ ] Address DAST findings
- [ ] Re-scan to verify fixes

---

## Automation

### GitHub Actions Workflow

All scans can be automated via GitHub Actions:
- CodeQL runs automatically on push/PR
- Dependabot checks dependencies daily
- Secret scanning runs on every push
- StackHawk can be integrated into CI/CD

### Scheduled Scans

Recommended schedule:
- **CodeQL:** Weekly (or on every push)
- **Dependabot:** Daily
- **Secret Scanning:** Real-time (on every push)
- **StackHawk:** Weekly (or before releases)

---

## Troubleshooting

### CodeQL Issues
- **Database creation fails:** Ensure Go is properly installed
- **No results:** Check that queries are running correctly
- **Timeout:** Increase timeout in workflow file

### Dependabot Issues
- **No alerts:** Check that dependencies are in go.mod
- **False positives:** Mark as dismissed with reason

### StackHawk Issues
- **Connection refused:** Ensure application is running
- **Authentication fails:** Check JWT token is valid
- **No endpoints found:** Verify OpenAPI spec is correct

---

## Next Steps

After running all scans:

1. **Document Results** - Update `SECURITY_ASSESSMENT.md` with actual findings
2. **Prioritize Fixes** - Address critical and high severity issues first
3. **Create Issues** - Track remediation in GitHub Issues
4. **Re-scan** - Verify fixes after remediation
5. **Submit to Security Team** - Share complete assessment

---

## Resources

- **CodeQL Documentation:** https://codeql.github.com/docs/
- **Dependabot Documentation:** https://docs.github.com/en/code-security/dependabot
- **Secret Scanning:** https://docs.github.com/en/code-security/secret-scanning
- **StackHawk Documentation:** https://docs.stackhawk.com/
- **gosec Documentation:** https://github.com/securecodewarrior/gosec
- **govulncheck Documentation:** https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck

---

**Last Updated:** January 2025  
**Maintained By:** Development Team

