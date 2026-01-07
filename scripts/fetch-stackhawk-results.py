#!/usr/bin/env python3
"""
StackHawk Results Fetcher and PDF Generator
Fetches StackHawk scan results and generates PDF report.
"""

import json
import os
import sys
import requests
from pathlib import Path
from typing import Dict, Any, Optional, List
from datetime import datetime

def get_stackhawk_api_key() -> Optional[str]:
    """Get StackHawk API key from environment variable."""
    api_key = os.environ.get('STACKHAWK_API_KEY') or os.environ.get('HAWK_API_KEY')
    if api_key:
        # Validate key format (StackHawk API keys are typically UUIDs or long strings)
        if len(api_key) < 10:
            print("⚠️  Warning: API key seems too short. Please verify it's correct.", file=sys.stderr)
    return api_key

def get_application_id() -> Optional[str]:
    """Get application ID from stackhawk.yml or environment."""
    # Try environment first
    app_id = os.environ.get('STACKHAWK_APPLICATION_ID')
    if app_id:
        return app_id
    
    # Try to read from stackhawk.yml
    try:
        import yaml
        with open('stackhawk.yml', 'r') as f:
            config = yaml.safe_load(f)
            return config.get('app', {}).get('applicationId')
    except Exception:
        pass
    
    return None

def fetch_latest_scan(api_key: str, application_id: str) -> Optional[Dict[str, Any]]:
    """Fetch the latest scan from StackHawk API."""
    headers = {
        'Authorization': f'Bearer {api_key}',
        'Content-Type': 'application/json',
        'Accept': 'application/json'
    }
    
    # Try different API endpoint formats
    # Format 1: Direct application scans endpoint
    url1 = f'https://api.stackhawk.com/api/v1/app/{application_id}/scans/latest'
    # Format 2: Scans endpoint with application filter
    url2 = f'https://api.stackhawk.com/api/v1/scans?applicationId={application_id}&limit=1'
    # Format 3: Original format
    url3 = f'https://api.stackhawk.com/api/v1/scans/{application_id}/latest'
    
    urls_to_try = [
        (url1, "application scans endpoint"),
        (url2, "scans list endpoint"),
        (url3, "direct scans endpoint")
    ]
    
    print(f"📡 Fetching latest scan for application {application_id}...")
    
    for url, description in urls_to_try:
        try:
            print(f"  Trying {description}...")
            response = requests.get(url, headers=headers, timeout=30)
            
            if response.status_code == 200:
                data = response.json()
                # Handle different response formats
                if isinstance(data, list) and len(data) > 0:
                    scan = data[0]  # Get first scan from list
                elif isinstance(data, dict):
                    scan = data
                else:
                    continue
                
                scan_id = scan.get('id') or scan.get('scanId')
                print(f"✅ Found scan: {scan_id}")
                return scan
            elif response.status_code == 401:
                # If we get 401, try next endpoint format
                error_detail = response.text
                print(f"  Authentication failed with {description}")
                if url == urls_to_try[-1][0]:  # Last URL
                    print(f"  Response: {error_detail[:200]}", file=sys.stderr)
                    print("  Authentication failed. Please verify:", file=sys.stderr)
                    print("  1. STACKHAWK_API_KEY is correct", file=sys.stderr)
                    print("  2. API key has proper permissions", file=sys.stderr)
                    print("  3. Application ID is correct", file=sys.stderr)
                continue
            elif response.status_code == 404:
                print(f"  No scans found with {description}")
                continue
            else:
                print(f"  Unexpected status {response.status_code} with {description}")
                continue
                
        except requests.exceptions.RequestException as e:
            if url == urls_to_try[-1][0]:  # Last URL, show error
                print(f"❌ Error fetching scan: {e}", file=sys.stderr)
                if hasattr(e, 'response') and e.response is not None:
                    print(f"  Status: {e.response.status_code}", file=sys.stderr)
                    print(f"  Response: {e.response.text[:200]}", file=sys.stderr)
            continue
    
    print("⚠️  Could not fetch scan from any endpoint", file=sys.stderr)
    return None

def fetch_scan_findings(api_key: str, scan_id: str) -> List[Dict[str, Any]]:
    """Fetch findings for a specific scan."""
    headers = {
        'Authorization': f'Bearer {api_key}',
        'Content-Type': 'application/json'
    }
    
    url = f'https://api.stackhawk.com/api/v1/scans/{scan_id}/findings'
    
    print(f"📊 Fetching findings for scan {scan_id}...")
    
    all_findings = []
    page = 1
    per_page = 100
    
    while True:
        params = {'page': page, 'perPage': per_page}
        try:
            response = requests.get(url, headers=headers, params=params, timeout=30)
            response.raise_for_status()
            
            data = response.json()
            findings = data.get('findings', [])
            if not findings:
                break
            
            all_findings.extend(findings)
            print(f"  Fetched page {page}: {len(findings)} findings")
            
            if len(findings) < per_page:
                break
            
            page += 1
        except requests.exceptions.RequestException as e:
            print(f"❌ Error fetching findings: {e}", file=sys.stderr)
            break
    
    return all_findings

def generate_pdf_report(scan: Dict[str, Any], findings: List[Dict[str, Any]], output_path: str):
    """Generate PDF report from scan and findings data."""
    try:
        from reportlab.lib import colors
        from reportlab.lib.pagesizes import letter, A4
        from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer, PageBreak
        from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
        from reportlab.lib.units import inch
    except ImportError:
        print("⚠️  reportlab not installed. Installing...")
        import subprocess
        subprocess.check_call([sys.executable, '-m', 'pip', 'install', 'reportlab'])
        from reportlab.lib import colors
        from reportlab.lib.pagesizes import letter, A4
        from reportlab.platypus import SimpleDocTemplate, Table, TableStyle, Paragraph, Spacer, PageBreak
        from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
        from reportlab.lib.units import inch
    
    doc = SimpleDocTemplate(output_path, pagesize=A4)
    story = []
    styles = getSampleStyleSheet()
    
    # Title
    title_style = ParagraphStyle(
        'CustomTitle',
        parent=styles['Heading1'],
        fontSize=24,
        textColor=colors.HexColor('#1a1a1a'),
        spaceAfter=30,
        alignment=1  # Center
    )
    story.append(Paragraph("StackHawk Security Scan Report", title_style))
    story.append(Spacer(1, 0.2*inch))
    
    # Scan Information
    scan_info_style = ParagraphStyle(
        'ScanInfo',
        parent=styles['Normal'],
        fontSize=10,
        textColor=colors.HexColor('#666666')
    )
    
    scan_id = scan.get('id', 'N/A')
    scan_status = scan.get('status', 'N/A')
    scan_started = scan.get('startedAt', 'N/A')
    scan_completed = scan.get('completedAt', scan.get('updatedAt', 'N/A'))
    
    story.append(Paragraph(f"<b>Scan ID:</b> {scan_id}", styles['Normal']))
    story.append(Paragraph(f"<b>Status:</b> {scan_status}", styles['Normal']))
    story.append(Paragraph(f"<b>Started:</b> {scan_started}", styles['Normal']))
    story.append(Paragraph(f"<b>Completed:</b> {scan_completed}", styles['Normal']))
    story.append(Spacer(1, 0.3*inch))
    
    # Summary
    story.append(Paragraph("Summary", styles['Heading2']))
    story.append(Spacer(1, 0.1*inch))
    
    # Count findings by severity
    severity_counts = {}
    for finding in findings:
        severity = finding.get('severity', 'UNKNOWN')
        severity_counts[severity] = severity_counts.get(severity, 0) + 1
    
    summary_data = [['Severity', 'Count']]
    for severity in ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'INFO']:
        count = severity_counts.get(severity, 0)
        if count > 0:
            summary_data.append([severity, str(count)])
    
    summary_table = Table(summary_data, colWidths=[3*inch, 1*inch])
    summary_table.setStyle(TableStyle([
        ('BACKGROUND', (0, 0), (-1, 0), colors.HexColor('#2c3e50')),
        ('TEXTCOLOR', (0, 0), (-1, 0), colors.whitesmoke),
        ('ALIGN', (0, 0), (-1, -1), 'LEFT'),
        ('FONTNAME', (0, 0), (-1, 0), 'Helvetica-Bold'),
        ('FONTSIZE', (0, 0), (-1, 0), 12),
        ('BOTTOMPADDING', (0, 0), (-1, 0), 12),
        ('BACKGROUND', (0, 1), (-1, -1), colors.beige),
        ('GRID', (0, 0), (-1, -1), 1, colors.black)
    ]))
    story.append(summary_table)
    story.append(Spacer(1, 0.3*inch))
    
    # Findings Details
    story.append(Paragraph("Security Findings", styles['Heading2']))
    story.append(Spacer(1, 0.1*inch))
    
    # Group findings by severity
    for severity in ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'INFO']:
        severity_findings = [f for f in findings if f.get('severity') == severity]
        if not severity_findings:
            continue
        
        story.append(Paragraph(f"{severity} Severity ({len(severity_findings)} findings)", styles['Heading3']))
        story.append(Spacer(1, 0.1*inch))
        
        for idx, finding in enumerate(severity_findings[:20], 1):  # Limit to 20 per severity
            finding_id = finding.get('id', 'N/A')
            title = finding.get('title', finding.get('name', 'N/A'))
            description = finding.get('description', finding.get('message', 'N/A'))
            url = finding.get('url', finding.get('requestUrl', 'N/A'))
            cwe = finding.get('cwe', 'N/A')
            
            story.append(Paragraph(f"<b>Finding {idx}:</b> {title}", styles['Normal']))
            story.append(Paragraph(f"<b>ID:</b> {finding_id}", scan_info_style))
            story.append(Paragraph(f"<b>URL:</b> {url}", scan_info_style))
            if cwe and cwe != 'N/A':
                story.append(Paragraph(f"<b>CWE:</b> {cwe}", scan_info_style))
            story.append(Paragraph(f"<b>Description:</b> {description[:200]}...", styles['Normal']))
            story.append(Spacer(1, 0.15*inch))
        
        if len(severity_findings) > 20:
            story.append(Paragraph(f"... and {len(severity_findings) - 20} more {severity} findings", scan_info_style))
        
        story.append(Spacer(1, 0.2*inch))
    
    # Footer
    story.append(Spacer(1, 0.3*inch))
    story.append(Paragraph(f"Report generated: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}", scan_info_style))
    story.append(Paragraph("Generated by StackHawk Security Scan", scan_info_style))
    
    doc.build(story)
    print(f"✅ PDF report generated: {output_path}")

def save_json_report(scan: Dict[str, Any], findings: List[Dict[str, Any]], output_path: str):
    """Save scan and findings as JSON."""
    report = {
        'scan': scan,
        'findings': findings,
        'summary': {
            'total_findings': len(findings),
            'severity_counts': {}
        },
        'generated_at': datetime.now().isoformat()
    }
    
    # Count by severity
    for finding in findings:
        severity = finding.get('severity', 'UNKNOWN')
        report['summary']['severity_counts'][severity] = report['summary']['severity_counts'].get(severity, 0) + 1
    
    with open(output_path, 'w', encoding='utf-8') as f:
        json.dump(report, f, indent=2, ensure_ascii=False)
    
    print(f"✅ JSON report saved: {output_path}")

def main():
    """Main function."""
    print("🔍 StackHawk Results Fetcher and PDF Generator")
    print("=" * 50)
    
    # Get API key
    api_key = get_stackhawk_api_key()
    if not api_key:
        print("❌ Error: STACKHAWK_API_KEY environment variable not set", file=sys.stderr)
        print("\n📝 How to get your StackHawk API key:", file=sys.stderr)
        print("  1. Log in to StackHawk platform: https://app.stackhawk.com", file=sys.stderr)
        print("  2. Go to Settings → API Keys", file=sys.stderr)
        print("  3. Create a new API key or copy an existing one", file=sys.stderr)
        print("  4. Set it as a GitHub secret: STACKHAWK_API_KEY", file=sys.stderr)
        return 1
    
    # Get application ID
    application_id = get_application_id()
    if not application_id:
        print("❌ Error: Could not determine StackHawk application ID", file=sys.stderr)
        print("   Set STACKHAWK_APPLICATION_ID or ensure stackhawk.yml exists", file=sys.stderr)
        return 1
    
    print(f"📦 Application ID: {application_id}")
    
    # Create output directory
    output_dir = Path('results')
    output_dir.mkdir(exist_ok=True)
    
    json_path = output_dir / 'stackhawk-results.json'
    pdf_path = output_dir / 'stackhawk-security-report.pdf'
    sarif_path = output_dir / 'stackhawk-results.sarif'
    
    # Try to parse SARIF file first (if scan was just run)
    scan = None
    findings = []
    
    if sarif_path.exists():
        print("📄 Found SARIF file, parsing it for results...")
        try:
            with open(sarif_path, 'r', encoding='utf-8') as f:
                sarif_data = json.load(f)
            
            runs = sarif_data.get('runs', [])
            for run in runs:
                results = run.get('results', [])
                if results:
                    print(f"✅ Found {len(results)} findings in SARIF file")
                    # Convert SARIF to StackHawk format
                    for result in results:
                        rule_id = result.get('ruleId', 'N/A')
                        message = result.get('message', {})
                        text = message.get('text', 'N/A')
                        level = result.get('level', 'warning')
                        
                        # Map SARIF levels to StackHawk severity
                        severity_map = {
                            'error': 'HIGH',
                            'warning': 'MEDIUM',
                            'note': 'LOW',
                            'none': 'INFO'
                        }
                        severity = severity_map.get(level.lower(), 'MEDIUM')
                        
                        locations = result.get('locations', [])
                        url = 'N/A'
                        if locations:
                            location = locations[0]
                            physical_location = location.get('physicalLocation', {})
                            artifact_location = physical_location.get('artifactLocation', {})
                            url = artifact_location.get('uri', 'N/A')
                        
                        findings.append({
                            'id': rule_id,
                            'title': text[:100] if text else rule_id,
                            'description': text,
                            'severity': severity,
                            'url': url,
                            'cwe': rule_id
                        })
                    
                    scan = {
                        'id': 'sarif-import',
                        'status': 'completed',
                        'startedAt': datetime.now().isoformat(),
                        'completedAt': datetime.now().isoformat()
                    }
                    break
        except Exception as e:
            print(f"⚠️  Could not parse SARIF file: {e}")
            print("  Will try to fetch from API instead...")
    
    # If no SARIF data, try API
    if not scan:
        scan = fetch_latest_scan(api_key, application_id)
    if not scan:
        print("⚠️  No scan found. Creating empty report.", file=sys.stderr)
        # Create empty JSON report
        empty_report = {
            'scan': None,
            'findings': [],
            'summary': {
                'total_findings': 0,
                'severity_counts': {}
            },
            'error': 'No scan found. Make sure a scan has been completed in StackHawk platform.',
            'generated_at': datetime.now().isoformat()
        }
        with open(json_path, 'w', encoding='utf-8') as f:
            json.dump(empty_report, f, indent=2, ensure_ascii=False)
        print(f"✅ Empty JSON report created: {json_path}")
        return 0
    
    scan_id = scan.get('id')
    if not scan_id:
        print("❌ Error: Scan ID not found. Creating report with available data.", file=sys.stderr)
        # Create JSON report with scan data but no findings
        save_json_report(scan, [], str(json_path))
        return 0
    
    # Fetch findings
    findings = fetch_scan_findings(api_key, scan_id)
    
    if not findings:
        print("⚠️  No findings found for this scan")
        findings = []
    
    # Save JSON report
    try:
        save_json_report(scan, findings, str(json_path))
    except Exception as e:
        print(f"⚠️  Error saving JSON report: {e}", file=sys.stderr)
        # Create minimal JSON report
        minimal_report = {
            'scan': scan,
            'findings': [],
            'summary': {'total_findings': 0, 'severity_counts': {}},
            'error': str(e),
            'generated_at': datetime.now().isoformat()
        }
        with open(json_path, 'w', encoding='utf-8') as f:
            json.dump(minimal_report, f, indent=2, ensure_ascii=False)
    
    # Generate PDF report
    try:
        generate_pdf_report(scan, findings, str(pdf_path))
    except Exception as e:
        print(f"⚠️  Error generating PDF: {e}", file=sys.stderr)
        print("  PDF generation failed, but JSON report is available", file=sys.stderr)
        # PDF file won't exist, which is fine - the workflow will skip uploading it
    
    # Summary
    print("\n" + "=" * 50)
    print("📈 Summary:")
    print(f"  Total findings: {len(findings)}")
    
    if findings:
        severity_counts = {}
        for finding in findings:
            severity = finding.get('severity', 'UNKNOWN')
            severity_counts[severity] = severity_counts.get(severity, 0) + 1
        
        print("\n  By Severity:")
        for severity in ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW', 'INFO']:
            count = severity_counts.get(severity, 0)
            if count > 0:
                print(f"    {severity}: {count}")
    
    print(f"\n✅ Files generated:")
    print(f"  - {json_path}")
    if pdf_path.exists():
        print(f"  - {pdf_path}")
    
    return 0

if __name__ == '__main__':
    sys.exit(main())

