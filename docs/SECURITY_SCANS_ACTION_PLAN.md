# Security Scans - Action Plan

**Quick action plan to run all security scans**

---

## ⚡ Quick Start (5 Minutes)

### Step 1: Install Tools
```bash
make install-tools
```

### Step 2: Run Local Scans
```bash
make security-scan
# Or
./scripts/run-security-scans.sh
```

### Step 3: Enable GitHub Scans
1. Go to GitHub repository
2. **Settings** → **Code security and analysis**
3. Enable:
   - ✅ Dependabot alerts
   - ✅ Dependabot security updates  
   - ✅ Secret scanning
   - ✅ Code scanning (CodeQL)

### Step 4: StackHawk Setup
1. Sign up: https://auth.stackhawk.com/login
2. Get API key from dashboard
3. Update `stackhawk.yml` with your config
4. Run: `hawk scan`

---

## 📋 Complete Checklist

### ✅ Setup (One-Time)

#### Local Tools
- [ ] Install Go 1.23+
- [ ] Run `make install-tools`
- [ ] Verify: `gosec --version`, `govulncheck --version`

#### GitHub Configuration
- [ ] Enable Dependabot alerts
- [ ] Enable Secret scanning
- [ ] Enable Code scanning (CodeQL)
- [ ] Verify workflows exist:
  - `.github/workflows/codeql.yml` ✅
  - `.github/workflows/security-scan.yml` ✅
  - `.github/workflows/stackhawk.yml` ✅
- [ ] Verify Dependabot config: `.github/dependabot.yml` ✅

#### StackHawk
- [ ] Create account at https://auth.stackhawk.com/login
- [ ] Get API key from dashboard
- [ ] Update `stackhawk.yml`:
  - [ ] Set `applicationId`
  - [ ] Set `host` (your app URL)
  - [ ] Set `token` (test JWT token)
- [ ] Install StackHawk CLI: `brew install stackhawk/homebrew-tap/hawk`

---

### 🔄 Running Scans

#### Local Scans (Run Now)
```bash
# Option 1: All scans at once
make security-scan

# Option 2: Individual scans
make security-gosec        # gosec scan
make security-vulncheck   # Vulnerability check
make security-deps        # Dependency updates
```

#### GitHub Scans (Automatic)
- [ ] Push code to trigger CodeQL scan
- [ ] Check **Security** tab for results
- [ ] Review Dependabot alerts (runs daily)
- [ ] Review Secret scanning results

#### StackHawk DAST
```bash
# 1. Start your application
make run
# Or: docker-compose up

# 2. Run StackHawk scan
hawk scan --config stackhawk.yml

# 3. View results at https://app.stackhawk.com
```

---

## 📊 Expected Results

### Local Scans
- **Location:** `security-reports/` directory
- **Files:**
  - `gosec-report.json` - gosec findings
  - `gosec-results.sarif` - SARIF format
  - `govulncheck-report.json` - Vulnerability report
  - `dependency-updates.txt` - Dependency updates

### GitHub Scans
- **Location:** Repository **Security** tab
- **Sections:**
  - Code scanning (CodeQL results)
  - Dependabot (dependency alerts)
  - Secret scanning (exposed secrets)

### StackHawk
- **Location:** https://app.stackhawk.com
- **Format:** Web dashboard with detailed findings

---

## 📝 Documenting Results

After running scans, update:

1. **SECURITY_ASSESSMENT.md**
   - CodeQL Scan Results section
   - Dependabot Scan Results section
   - Code Scanning Results section
   - Secret Scanning Results section
   - DAST/StackHawk Scan Results section

2. **Create Issues**
   - One issue per critical/high finding
   - Link to scan results
   - Assign to team members

3. **Track Progress**
   - Update checklist in SECURITY_ASSESSMENT.md
   - Mark findings as resolved
   - Re-scan to verify fixes

---

## 🎯 Priority Order

### 1. Critical Issues (Fix Immediately)
- SQL injection vulnerabilities
- Authentication bypass
- Exposed secrets
- Remote code execution

### 2. High Issues (Fix This Week)
- XSS vulnerabilities
- Authorization flaws
- Insecure deserialization
- Critical dependency vulnerabilities

### 3. Medium Issues (Fix This Month)
- Security misconfigurations
- Information disclosure
- Moderate dependency vulnerabilities

### 4. Low Issues (Fix When Convenient)
- Best practice violations
- Low severity dependency issues
- Informational findings

---

## 🔄 Continuous Scanning

### Automated Scans
- **CodeQL:** Weekly (or on every push)
- **Dependabot:** Daily
- **Secret Scanning:** Real-time (on every push)
- **StackHawk:** Weekly (or before releases)

### Manual Scans
- **Local scans:** Before committing
- **StackHawk:** Before releases
- **Full assessment:** Quarterly

---

## 📚 Documentation Reference

- **[SECURITY_SCAN_GUIDE.md](./SECURITY_SCAN_GUIDE.md)** - Detailed instructions
- **[SECURITY_SCAN_QUICK_START.md](./SECURITY_SCAN_QUICK_START.md)** - Quick reference
- **[SECURITY_ASSESSMENT.md](./SECURITY_ASSESSMENT.md)** - Assessment document
- **[SECURITY_CREDENTIALS_REFERENCE.md](./SECURITY_CREDENTIALS_REFERENCE.md)** - Credentials

---

## 🆘 Troubleshooting

### Scans Not Running?
1. Check GitHub Actions are enabled
2. Verify workflow files exist
3. Check Actions tab for errors
4. Review workflow logs

### No Results?
1. Ensure code is pushed to GitHub
2. Wait for scans to complete (5-10 minutes)
3. Check Security tab
4. Verify permissions are correct

### Local Scans Failing?
1. Run `go mod download` first
2. Check tool installation: `which gosec`
3. Verify Go version: `go version`
4. Check for errors in output

---

## ✅ Success Criteria

You've successfully completed security scans when:

- [ ] All local scans run without errors
- [ ] GitHub scans are enabled and running
- [ ] StackHawk scan completes successfully
- [ ] All results are documented in SECURITY_ASSESSMENT.md
- [ ] Critical issues are identified and tracked
- [ ] Remediation plan is in place

---

## 📞 Support

- **GitHub Issues:** Create issue in repository
- **StackHawk:** https://docs.stackhawk.com/
- **CodeQL:** https://codeql.github.com/docs/
- **Dependabot:** https://docs.github.com/en/code-security/dependabot

---

**Last Updated:** January 2025  
**Status:** Ready to Execute

