# Security Scan Quick Start Guide

**Quick reference for running all security scans**

---

## 🚀 Quick Start (5 Minutes)

### 1. Local Scans (Run Now)
```bash
# Install tools (one-time)
make install-tools

# Run all local security scans
./scripts/run-security-scans.sh

# Or use Makefile
make security-scan
```

### 2. GitHub-Based Scans (Automatic)

#### Enable on GitHub:
1. Go to your repository → **Settings** → **Code security and analysis**
2. Enable:
   - ✅ **Dependabot alerts**
   - ✅ **Dependabot security updates**
   - ✅ **Secret scanning**
   - ✅ **Code scanning** (CodeQL)

#### That's it! Scans run automatically.

---

## 📋 Detailed Steps

### CodeQL Scan

**Option 1: Automatic (Recommended)**
- Already configured in `.github/workflows/codeql.yml`
- Runs automatically on push/PR
- View results: **Security** tab → **Code scanning**

**Option 2: Manual**
```bash
# Install CodeQL CLI
wget https://github.com/github/codeql-cli-binaries/releases/latest/download/codeql-bundle-linux64.tar.gz
tar -xzf codeql-bundle-linux64.tar.gz
export PATH=$PATH:$(pwd)/codeql

# Create database and scan
codeql database create codeql-db --language=go --source-root=.
codeql database analyze codeql-db --format=sarif-latest --output=codeql-results.sarif
```

---

### Dependabot Scan

**Automatic (No action needed)**
- Configured in `.github/dependabot.yml`
- Runs daily automatically
- View results: **Security** tab → **Dependabot**

**Manual Check:**
```bash
# Check for vulnerabilities
govulncheck ./...

# Check for updates
go list -m -u all
```

---

### Code Scanning

**Automatic (No action needed)**
- Enabled with CodeQL
- View results: **Security** tab → **Code scanning**

**Additional Scanners:**
- gosec (via GitHub Actions workflow)
- View: **Actions** tab → **Security Scans** workflow

---

### Secret Scanning

**Automatic (No action needed)**
- Enabled in GitHub Settings
- Scans every push automatically
- View results: **Security** tab → **Secret scanning**

**Manual Check:**
```bash
# Using GitGuardian CLI
pip install ggshield
ggshield scan repo .

# Or use the script
./scripts/run-security-scans.sh
```

---

### StackHawk DAST Scan

**Step 1: Create Account**
- Go to: https://auth.stackhawk.com/login
- Sign up / Log in

**Step 2: Get API Key**
- Dashboard → Settings → API Keys
- Copy API Key

**Step 3: Configure**
```bash
# Edit stackhawk.yml
# Update: host, applicationId, token
```

**Step 4: Run Scan**
```bash
# Install StackHawk CLI
brew install stackhawk/homebrew-tap/hawk  # macOS
# Or download from: https://docs.stackhawk.com/hawk-cli/installation

# Start your application
make run  # or docker-compose up

# Run scan
hawk scan --config stackhawk.yml
```

**Step 5: View Results**
- Dashboard: https://app.stackhawk.com

**Or via GitHub Actions:**
- Add secrets: `STACKHAWK_API_KEY`, `STACKHAWK_APP_ID`, `STACKHAWK_ENV_ID`
- Workflow runs automatically (configured in `.github/workflows/stackhawk.yml`)

---

## 🛠️ Installation Commands

```bash
# Install all tools at once
make install-tools

# Or individually:
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest

# StackHawk CLI
brew install stackhawk/homebrew-tap/hawk  # macOS
# Or: https://docs.stackhawk.com/hawk-cli/installation

# GitGuardian (optional)
pip install ggshield
```

---

## 📊 View Results

### GitHub
- **Security Tab** → All scan results
- **Actions Tab** → Workflow runs and artifacts

### Local
- Reports in: `security-reports/` directory
- JSON/SARIF files for integration

### StackHawk
- Dashboard: https://app.stackhawk.com
- Login and view application scans

---

## ✅ Checklist

### Initial Setup
- [ ] Install local tools (`make install-tools`)
- [ ] Enable GitHub security features
- [ ] Create StackHawk account
- [ ] Configure `stackhawk.yml`

### Run Scans
- [ ] Run local scans (`./scripts/run-security-scans.sh`)
- [ ] Push code to trigger GitHub scans
- [ ] Run StackHawk scan (`hawk scan`)

### Review Results
- [ ] Check GitHub Security tab
- [ ] Review local reports
- [ ] Check StackHawk dashboard
- [ ] Document findings in `SECURITY_ASSESSMENT.md`

### Remediation
- [ ] Fix critical issues
- [ ] Update vulnerable dependencies
- [ ] Rotate exposed secrets
- [ ] Re-scan to verify fixes

---

## 🔧 Troubleshooting

### CodeQL not running?
- Check GitHub Actions are enabled
- Verify workflow file exists: `.github/workflows/codeql.yml`
- Check Actions tab for errors

### Dependabot not working?
- Verify `.github/dependabot.yml` exists
- Check repository settings → Dependabot enabled
- Ensure `go.mod` is present

### StackHawk connection failed?
- Ensure application is running
- Check `stackhawk.yml` configuration
- Verify JWT token is valid
- Check network/firewall settings

### Local scans failing?
- Ensure Go 1.23+ is installed
- Run `go mod download` first
- Check tool installation: `which gosec`

---

## 📚 Full Documentation

For detailed instructions, see:
- **[SECURITY_SCAN_GUIDE.md](./SECURITY_SCAN_GUIDE.md)** - Complete guide
- **[SECURITY_ASSESSMENT.md](./SECURITY_ASSESSMENT.md)** - Assessment document

---

## 🆘 Need Help?

- **GitHub Issues:** Create an issue in the repository
- **StackHawk Support:** https://docs.stackhawk.com/
- **CodeQL Docs:** https://codeql.github.com/docs/

---

**Last Updated:** January 2025

