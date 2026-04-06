import requests
import ssl
import socket
import datetime
import json
import os
import time
import whois
from dateutil import parser
from concurrent.futures import ThreadPoolExecutor, as_completed

DOMAINS_FILE = "domains.txt"
DATA_FILE = "data.json"

def get_domain_expiration(domain):
    # Try RDAP first (Standard)
    try:
        r = requests.get(f"https://rdap.org/domain/{domain}", timeout=5)
        if r.status_code == 200:
            events = r.json().get("events", [])
            for e in events:
                if e.get("eventAction") == "expiration":
                    dt = parser.parse(e.get("eventDate")).replace(tzinfo=None)
                    return dt
    except Exception:
        pass

    # Try Backup WHOIS API
    try:
        r = requests.get(f"https://networkcalc.com/api/dns/whois/{domain}", timeout=5)
        if r.status_code == 200:
            data = r.json()
            if data.get("status") == "OK" and data.get("whois"):
                exp = data["whois"].get("registry_expiration_date") or data["whois"].get("expires")
                if exp:
                    dt = parser.parse(exp).replace(tzinfo=None)
                    return dt
    except Exception:
        pass

    # Try .tr Specific Fallback
    try:
        if domain.endswith('.tr'):
            r = requests.get(f"https://whois.enis.org.tr/whois?domain={domain}", timeout=10)
            if r.status_code == 200:
                import re
                match = re.search(r"Expires on\.+:\s*([0-9]{4}-[0-9]{2}-[0-9]{2})", r.text)
                if match:
                    return parser.parse(match.group(1)).replace(tzinfo=None)
    except Exception:
        pass

    # Try Standard whois library (Port 43)
    try:
        w = whois.whois(domain)
        d = w.expiration_date
        if isinstance(d, list):
            for item in d:
                if isinstance(item, datetime.datetime):
                    return item.replace(tzinfo=None)
            return None
        if isinstance(d, datetime.datetime):
            return d.replace(tzinfo=None)
    except Exception:
        pass

    return None

def get_ssl_expiration(domain):
    try:
        context = ssl.create_default_context()
        with socket.create_connection((domain, 443), timeout=10) as sock:
            with context.wrap_socket(sock, server_hostname=domain) as ssock:
                cert = ssock.getpeercert()
                expiration_date = datetime.datetime.strptime(cert['notAfter'], '%b %d %H:%M:%S %Y %Z')
                return expiration_date.replace(tzinfo=None)
    except Exception:
        pass
    return None

def process_domain(domain):
    print(f"Checking: {domain}...")
    
    domain_expiry = None
    ssl_expiry = None
    
    try:
        domain_expiry = get_domain_expiration(domain)
        if not isinstance(domain_expiry, datetime.datetime):
            domain_expiry = None
    except Exception:
        pass
        
    try:
        ssl_expiry = get_ssl_expiration(domain)
        if not isinstance(ssl_expiry, datetime.datetime):
            ssl_expiry = None
    except Exception:
        pass
        
    now = datetime.datetime.now(datetime.timezone.utc).replace(tzinfo=None)
    
    return {
        "domain_name": domain,
        "domain": {
            "expiry": domain_expiry.isoformat() if domain_expiry else None,
            "days_left": (domain_expiry - now).days if domain_expiry else "Error"
        },
        "ssl": {
            "expiry": ssl_expiry.isoformat() if ssl_expiry else None,
            "days_left": (ssl_expiry - now).days if ssl_expiry else "Error"
        }
    }

def main():
    if not os.path.exists(DOMAINS_FILE):
        print(f"Error: {DOMAINS_FILE} not found.")
        return

    with open(DOMAINS_FILE, "r") as f:
        domains = [line.strip() for line in f if line.strip()]

    print(f"Turbo Mode Enabled. Total {len(domains)} domains will be scanned in parallel.")
    print("---------------------------------")
    
    start_time = time.time()
    results = []

    # Parallel processing with 10 threads
    with ThreadPoolExecutor(max_workers=10) as executor:
        future_to_domain = {executor.submit(process_domain, domain): domain for domain in domains}
        for future in as_completed(future_to_domain):
            domain = future_to_domain[future]
            try:
                result = future.result()
                results.append(result)
            except Exception as e:
                print(f"{domain} generated an exception: {e}")

    end_time = time.time()
    total_time = round(end_time - start_time, 2)

    # Sort results to match the order in domains.txt
    results.sort(key=lambda x: domains.index(x['domain_name']))

    data = {
        "items": results,
        "last_updated": datetime.datetime.now(datetime.timezone.utc).isoformat()
    }
    
    with open(DATA_FILE, "w") as f:
        json.dump(data, f, indent=4)
        
    print("---------------------------------")
    print(f"Scan complete. {len(results)} domains saved in {total_time}s.")

if __name__ == "__main__":
    main()
