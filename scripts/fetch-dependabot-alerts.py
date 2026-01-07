#!/usr/bin/env python3
"""
Dependabot Alerts Fetcher
Fetches Dependabot alerts from GitHub API and saves them as JSON file.
"""

import json
import os
import sys
import requests
from pathlib import Path
from typing import List, Dict, Any, Optional

def get_github_token() -> Optional[str]:
    """Get GitHub token from environment variable."""
    token = os.environ.get('GITHUB_TOKEN')
    if not token:
        # Try alternative names
        token = os.environ.get('GH_TOKEN')
    return token

def get_repo_info() -> tuple[str, str]:
    """Extract repository owner and name from environment or git config."""
    # Try GitHub Actions environment variables first
    repo = os.environ.get('GITHUB_REPOSITORY')
    if repo:
        owner, repo_name = repo.split('/', 1)
        return owner, repo_name
    
    # Try to get from git remote
    try:
        import subprocess
        result = subprocess.run(
            ['git', 'config', '--get', 'remote.origin.url'],
            capture_output=True,
            text=True,
            check=True
        )
        url = result.stdout.strip()
        # Handle both https and ssh formats
        if 'github.com' in url:
            if url.startswith('https://'):
                parts = url.replace('https://github.com/', '').replace('.git', '').split('/')
            elif url.startswith('git@'):
                parts = url.replace('git@github.com:', '').replace('.git', '').split('/')
            else:
                parts = url.split('/')
            if len(parts) >= 2:
                return parts[-2], parts[-1]
    except Exception:
        pass
    
    # Default fallback (user should set via environment)
    return os.environ.get('GITHUB_OWNER', ''), os.environ.get('GITHUB_REPO', '')

def fetch_dependabot_alerts(owner: str, repo: str, token: str, state: str = 'open') -> List[Dict[str, Any]]:
    """Fetch Dependabot alerts from GitHub API."""
    headers = {
        'Accept': 'application/vnd.github+json',
        'Authorization': f'Bearer {token}',
        'X-GitHub-Api-Version': '2022-11-28'
    }
    
    url = f'https://api.github.com/repos/{owner}/{repo}/dependabot/alerts'
    params = {
        'state': state,
        'per_page': 100
    }
    
    all_alerts = []
    page = 1
    
    print(f"📡 Fetching Dependabot alerts from {owner}/{repo}...")
    
    while True:
        params['page'] = page
        try:
            response = requests.get(url, headers=headers, params=params, timeout=30)
            response.raise_for_status()
            
            alerts = response.json()
            if not alerts:
                break
            
            all_alerts.extend(alerts)
            print(f"  Fetched page {page}: {len(alerts)} alerts")
            
            # Check if there are more pages
            if len(alerts) < params['per_page']:
                break
            
            page += 1
            
        except requests.exceptions.RequestException as e:
            print(f"❌ Error fetching alerts: {e}", file=sys.stderr)
            if hasattr(e.response, 'status_code'):
                if e.response.status_code == 404:
                    print("  Repository not found or Dependabot not enabled", file=sys.stderr)
                elif e.response.status_code == 403:
                    print("  Permission denied. Check your token permissions.", file=sys.stderr)
                elif e.response.status_code == 401:
                    print("  Authentication failed. Check your token.", file=sys.stderr)
            break
    
    return all_alerts

def save_alerts_json(alerts: List[Dict[str, Any]], output_path: str):
    """Save alerts to JSON file."""
    output_file = Path(output_path)
    output_file.parent.mkdir(parents=True, exist_ok=True)
    
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(alerts, f, indent=2, ensure_ascii=False)
    
    print(f"✅ Saved {len(alerts)} alerts to {output_path}")

def generate_csv(alerts: List[Dict[str, Any]], output_path: str):
    """Generate CSV file from alerts."""
    import csv
    
    if not alerts:
        print("No alerts to write to CSV", file=sys.stderr)
        return
    
    fieldnames = [
        'Alert Number',
        'State',
        'Package Name',
        'Ecosystem',
        'Manifest Path',
        'Scope',
        'Relationship',
        'Severity',
        'GHSA ID',
        'CVE ID',
        'Summary',
        'Vulnerable Version Range',
        'First Patched Version',
        'CVSS Score',
        'Published At',
        'Updated At',
        'HTML URL'
    ]
    
    output_file = Path(output_path)
    output_file.parent.mkdir(parents=True, exist_ok=True)
    
    with open(output_file, 'w', newline='', encoding='utf-8') as csvfile:
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        writer.writeheader()
        
        for alert in alerts:
            dependency = alert.get('dependency', {})
            package = dependency.get('package', {})
            security_advisory = alert.get('security_advisory', {})
            security_vulnerability = alert.get('security_vulnerability', {})
            cvss = security_advisory.get('cvss', {})
            
            row = {
                'Alert Number': alert.get('number', ''),
                'State': alert.get('state', ''),
                'Package Name': package.get('name', ''),
                'Ecosystem': package.get('ecosystem', ''),
                'Manifest Path': dependency.get('manifest_path', ''),
                'Scope': dependency.get('scope', ''),
                'Relationship': dependency.get('relationship', ''),
                'Severity': security_vulnerability.get('severity', ''),
                'GHSA ID': security_advisory.get('ghsa_id', ''),
                'CVE ID': security_advisory.get('cve_id', ''),
                'Summary': security_advisory.get('summary', ''),
                'Vulnerable Version Range': security_vulnerability.get('vulnerable_version_range', ''),
                'First Patched Version': security_vulnerability.get('first_patched_version', {}).get('identifier', ''),
                'CVSS Score': cvss.get('score', ''),
                'Published At': security_advisory.get('published_at', ''),
                'Updated At': alert.get('updated_at', ''),
                'HTML URL': alert.get('html_url', '')
            }
            writer.writerow(row)
    
    print(f"✅ Generated CSV file: {output_path}")

def main():
    """Main function."""
    print("🔍 Dependabot Alerts Fetcher")
    print("=" * 50)
    
    # Get GitHub token
    token = get_github_token()
    if not token:
        print("❌ Error: GITHUB_TOKEN environment variable not set", file=sys.stderr)
        print("   Set it with: export GITHUB_TOKEN=your_token", file=sys.stderr)
        return 1
    
    # Get repository info
    owner, repo = get_repo_info()
    if not owner or not repo:
        print("❌ Error: Could not determine repository owner and name", file=sys.stderr)
        print("   Set GITHUB_OWNER and GITHUB_REPO environment variables", file=sys.stderr)
        return 1
    
    print(f"📦 Repository: {owner}/{repo}")
    
    # Fetch alerts
    alerts = fetch_dependabot_alerts(owner, repo, token, state='open')
    
    if not alerts:
        print("⚠️  No Dependabot alerts found")
        # Create empty file
        alerts = []
    
    # Save JSON file
    output_dir = Path('results')
    output_dir.mkdir(exist_ok=True)
    
    json_path = output_dir / 'dependabot-alerts-backend.json'
    save_alerts_json(alerts, str(json_path))
    
    # Generate CSV
    csv_path = output_dir / 'dependabot-alerts-backend.csv'
    generate_csv(alerts, str(csv_path))
    
    # Summary
    print("\n" + "=" * 50)
    print("📈 Summary:")
    print(f"  Total alerts: {len(alerts)}")
    
    if alerts:
        severity_counts = {}
        ecosystem_counts = {}
        for alert in alerts:
            severity = alert.get('security_vulnerability', {}).get('severity', 'unknown')
            ecosystem = alert.get('dependency', {}).get('package', {}).get('ecosystem', 'unknown')
            severity_counts[severity] = severity_counts.get(severity, 0) + 1
            ecosystem_counts[ecosystem] = ecosystem_counts.get(ecosystem, 0) + 1
        
        print("\n  By Severity:")
        for severity, count in sorted(severity_counts.items(), key=lambda x: x[1], reverse=True):
            print(f"    {severity}: {count}")
        
        print("\n  By Ecosystem:")
        for ecosystem, count in sorted(ecosystem_counts.items(), key=lambda x: x[1], reverse=True):
            print(f"    {ecosystem}: {count}")
    
    print(f"\n✅ Files generated:")
    print(f"  - {json_path}")
    print(f"  - {csv_path}")
    
    return 0

if __name__ == '__main__':
    sys.exit(main())

