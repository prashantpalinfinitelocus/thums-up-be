#!/usr/bin/env python3
"""
Code Scanning Results CSV Extractor
Extracts findings from all security scan reports and generates a comprehensive CSV file.
"""

import json
import csv
import os
import sys
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Any

def parse_sarif_file(sarif_path: str, tool_name: str) -> List[Dict[str, Any]]:
    """Parse SARIF file and extract findings."""
    findings = []
    
    try:
        with open(sarif_path, 'r', encoding='utf-8') as f:
            sarif_data = json.load(f)
        
        runs = sarif_data.get('runs', [])
        for run in runs:
            tool = run.get('tool', {})
            driver = tool.get('driver', {})
            tool_name_from_file = driver.get('name', tool_name)
            
            results = run.get('results', [])
            for result in results:
                rule_id = result.get('ruleId', 'N/A')
                message = result.get('message', {})
                text = message.get('text', 'N/A')
                
                level = result.get('level', 'warning')
                severity = 'HIGH' if level == 'error' else 'MEDIUM' if level == 'warning' else 'LOW'
                
                locations = result.get('locations', [])
                for location in locations:
                    physical_location = location.get('physicalLocation', {})
                    artifact_location = physical_location.get('artifactLocation', {})
                    file_path = artifact_location.get('uri', 'N/A')
                    
                    region = physical_location.get('region', {})
                    start_line = region.get('startLine', 0)
                    start_column = region.get('startColumn', 0)
                    
                    findings.append({
                        'Tool': tool_name_from_file,
                        'Severity': severity,
                        'Rule ID': rule_id,
                        'Message': text,
                        'File': file_path,
                        'Line': start_line,
                        'Column': start_column,
                        'Source': 'SARIF',
                        'Timestamp': datetime.now().isoformat()
                    })
    except Exception as e:
        print(f"Error parsing SARIF file {sarif_path}: {e}", file=sys.stderr)
    
    return findings

def parse_gosec_json(gosec_path: str) -> List[Dict[str, Any]]:
    """Parse gosec JSON report."""
    findings = []
    
    try:
        with open(gosec_path, 'r', encoding='utf-8') as f:
            gosec_data = json.load(f)
        
        issues = gosec_data.get('Issues', [])
        for issue in issues:
            severity = issue.get('severity', 'LOW')
            rule_id = issue.get('rule_id', 'N/A')
            details = issue.get('details', 'N/A')
            file_path = issue.get('file', 'N/A')
            line = issue.get('line', 0)
            column = issue.get('column', 0)
            
            findings.append({
                'Tool': 'gosec',
                'Severity': severity,
                'Rule ID': rule_id,
                'Message': details,
                'File': file_path,
                'Line': line,
                'Column': column,
                'Source': 'gosec-json',
                'Timestamp': datetime.now().isoformat()
            })
    except Exception as e:
        print(f"Error parsing gosec JSON file {gosec_path}: {e}", file=sys.stderr)
    
    return findings

def parse_govulncheck_json(govulncheck_path: str) -> List[Dict[str, Any]]:
    """Parse govulncheck JSON report."""
    findings = []
    
    try:
        with open(govulncheck_path, 'r', encoding='utf-8') as f:
            vuln_data = json.load(f)
        
        # govulncheck JSON structure
        vulns = vuln_data.get('Vulns', [])
        for vuln in vulns:
            osv = vuln.get('OSV', {})
            id_val = osv.get('id', 'N/A')
            summary = osv.get('summary', 'N/A')
            severity = 'HIGH'  # Default for vulnerabilities
            
            # Get affected modules
            affected = osv.get('affected', [])
            for aff in affected:
                packages = aff.get('packages', [])
                for pkg in packages:
                    pkg_name = pkg.get('name', 'N/A')
                    
                    findings.append({
                        'Tool': 'govulncheck',
                        'Severity': severity,
                        'Rule ID': id_val,
                        'Message': summary,
                        'File': f"Package: {pkg_name}",
                        'Line': 0,
                        'Column': 0,
                        'Source': 'govulncheck-json',
                        'Timestamp': datetime.now().isoformat()
                    })
    except Exception as e:
        print(f"Error parsing govulncheck JSON file {govulncheck_path}: {e}", file=sys.stderr)
    
    return findings

def find_report_files(base_dir: str = '.') -> Dict[str, List[str]]:
    """Find all report files in the repository."""
    reports = {
        'sarif': [],
        'gosec': [],
        'govulncheck': []
    }
    
    base_path = Path(base_dir)
    
    # Search in current directory and results folder
    search_paths = [base_path, base_path / 'results']
    
    for search_path in search_paths:
        if not search_path.exists():
            continue
            
        # Find SARIF files
        for sarif_file in search_path.glob('**/*.sarif'):
            reports['sarif'].append(str(sarif_file))
        
        # Find gosec JSON files
        for gosec_file in search_path.glob('**/gosec*.json'):
            if 'gosec-report.json' in str(gosec_file) or 'gosec-report' in str(gosec_file):
                reports['gosec'].append(str(gosec_file))
        
        # Find govulncheck JSON files
        for vuln_file in search_path.glob('**/govulncheck*.json'):
            reports['govulncheck'].append(str(vuln_file))
    
    return reports

def generate_csv(findings: List[Dict[str, Any]], output_path: str):
    """Generate CSV file from findings."""
    if not findings:
        print("No findings to write to CSV", file=sys.stderr)
        return
    
    fieldnames = ['Tool', 'Severity', 'Rule ID', 'Message', 'File', 'Line', 'Column', 'Source', 'Timestamp']
    
    with open(output_path, 'w', newline='', encoding='utf-8') as csvfile:
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(findings)
    
    print(f"✓ Generated CSV file: {output_path}")
    print(f"  Total findings: {len(findings)}")

def main():
    """Main function to extract and generate CSV."""
    print("🔍 Code Scanning Results CSV Extractor")
    print("=" * 50)
    
    # Find all report files
    print("\n📂 Searching for report files...")
    reports = find_report_files()
    
    print(f"  Found {len(reports['sarif'])} SARIF file(s)")
    print(f"  Found {len(reports['gosec'])} gosec JSON file(s)")
    print(f"  Found {len(reports['govulncheck'])} govulncheck JSON file(s)")
    
    all_findings = []
    
    # Parse SARIF files
    print("\n📊 Parsing SARIF files...")
    for sarif_file in reports['sarif']:
        print(f"  Processing: {sarif_file}")
        # Determine tool name from filename
        tool_name = 'CodeQL'
        if 'gosec' in sarif_file.lower():
            tool_name = 'gosec'
        elif 'stackhawk' in sarif_file.lower():
            tool_name = 'StackHawk'
        
        findings = parse_sarif_file(sarif_file, tool_name)
        all_findings.extend(findings)
        print(f"    Found {len(findings)} findings")
    
    # Parse gosec JSON files
    print("\n📊 Parsing gosec JSON files...")
    for gosec_file in reports['gosec']:
        print(f"  Processing: {gosec_file}")
        findings = parse_gosec_json(gosec_file)
        all_findings.extend(findings)
        print(f"    Found {len(findings)} findings")
    
    # Parse govulncheck JSON files
    print("\n📊 Parsing govulncheck JSON files...")
    for vuln_file in reports['govulncheck']:
        print(f"  Processing: {vuln_file}")
        findings = parse_govulncheck_json(vuln_file)
        all_findings.extend(findings)
        print(f"    Found {len(findings)} findings")
    
    # Generate CSV
    output_dir = Path('results')
    output_dir.mkdir(exist_ok=True)
    output_path = output_dir / 'code-scanning-files-extracted.csv'
    
    print(f"\n💾 Generating CSV file...")
    generate_csv(all_findings, str(output_path))
    
    # Summary
    print("\n" + "=" * 50)
    print("📈 Summary:")
    print(f"  Total findings: {len(all_findings)}")
    
    if all_findings:
        severity_counts = {}
        tool_counts = {}
        for finding in all_findings:
            severity = finding.get('Severity', 'UNKNOWN')
            tool = finding.get('Tool', 'UNKNOWN')
            severity_counts[severity] = severity_counts.get(severity, 0) + 1
            tool_counts[tool] = tool_counts.get(tool, 0) + 1
        
        print("\n  By Severity:")
        for severity, count in sorted(severity_counts.items(), key=lambda x: x[1], reverse=True):
            print(f"    {severity}: {count}")
        
        print("\n  By Tool:")
        for tool, count in sorted(tool_counts.items(), key=lambda x: x[1], reverse=True):
            print(f"    {tool}: {count}")
    
    print(f"\n✅ CSV file generated: {output_path}")
    return 0

if __name__ == '__main__':
    sys.exit(main())

