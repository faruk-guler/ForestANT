export interface Domain {
  id: number;
  hostname: string;
  ssl_expiry: string | null;
  domain_expiry: string | null;
  last_scan: string | null;
  status: string;
  nameservers: string | null;
  security_rating: string | null;
  status_availability: string | null;
  last_whois_raw: string | null;
}

export interface Settings {
  webhook_url: string;
  notification_threshold: string;
  telegram_token?: string;
  telegram_chat_id?: string;
  discord_webhook_url?: string;
  slack_webhook_url?: string;
  smtp_host?: string;
  smtp_port?: string;
  smtp_user?: string;
  smtp_pass?: string;
  smtp_from?: string;
  smtp_to?: string;
  scan_interval?: string;
}

export interface Access {
  id: number;
  name: string;
  provider: 'ssh' | 'dns_aliyun' | 'dns_cloudflare';
  config: string; // JSON string
  created_at: string;
}

export interface Workflow {
  id: number;
  name: string;
  domain_id: number;
  certificate_id?: number;
  access_id: number;
  type: 'deploy_ssh' | 'acme_http' | 'acme_dns';
  status: 'pending' | 'running' | 'success' | 'failed';
  last_run?: string;
  config?: string;
}
