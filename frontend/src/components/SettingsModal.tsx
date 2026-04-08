import React, { useState } from 'react';
import { X, Send } from 'lucide-react';
import axios from 'axios';
import type { Settings } from '../types';
import { useToast } from '../contexts/ToastContext';

interface Props {
  settings: Settings;
  setSettings: (s: Settings) => void;
  onClose: () => void;
  apiUrl: string;
}

const SettingsModal: React.FC<Props> = ({ settings, setSettings, onClose, apiUrl }) => {
  const { addToast } = useToast();
  const [testing, setTesting] = useState(false);
  const [activeTab, setActiveTab] = useState<'general' | 'im' | 'notifications'>('general');

  const saveSettings = async () => {
    try {
      await axios.post(`${apiUrl}/settings`, settings);
      onClose();
      addToast('Settings saved successfully!', 'success');
    } catch (error) {
      addToast('Error saving settings', 'error');
    }
  };

  const testWebhook = async () => {
    if (!settings.webhook_url) {
      addToast('Please enter a webhook URL first and save it.', 'info');
      return;
    }
    setTesting(true);
    try {
      const { data } = await axios.post(`${apiUrl}/settings/test-webhook`);
      addToast(data.message, 'success');
    } catch (error) {
      addToast('Error sending test webhook', 'error');
    } finally {
      setTesting(false);
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal card" style={{ maxWidth: '650px' }}>
        <div className="modal-header">
          <h2>System Configuration</h2>
          <button onClick={onClose} className="btn-icon">
            <X size={24} />
          </button>
        </div>

        <div className="tabs-header">
          <button 
            className={`tab-btn ${activeTab === 'general' ? 'active' : ''}`} 
            onClick={() => setActiveTab('general')}
          >
            General
          </button>
          <button 
            className={`tab-btn ${activeTab === 'im' ? 'active' : ''}`} 
            onClick={() => setActiveTab('im')}
          >
            IM / Social
          </button>
          <button 
            className={`tab-btn ${activeTab === 'notifications' ? 'active' : ''}`} 
            onClick={() => setActiveTab('notifications')}
          >
            Email & Webhook
          </button>
        </div>

        <div className="modal-body">
          {activeTab === 'general' && (
            <div className="tab-content">
              <div className="input-group">
                <label style={{ fontWeight: 'bold' }}>Alert Threshold (Days)</label>
                <input 
                  type="number" 
                  value={settings.notification_threshold}
                  onChange={e => setSettings({...settings, notification_threshold: e.target.value})}
                />
                <p className="input-hint">Notify when certificate has less than this many days left.</p>
              </div>

              <div className="input-group">
                <label style={{ fontWeight: 'bold' }}>Auto-Scan Interval</label>
                <select 
                  value={settings.scan_interval || '24'}
                  onChange={e => setSettings({...settings, scan_interval: e.target.value})}
                  style={{ padding: '0.6rem', borderRadius: '8px', border: '1px solid var(--border)', background: 'var(--input-bg)', color: 'var(--text-primary)', width: '100%' }}
                >
                  <option value="5">Every 5 Minutes (TEST ONLY)</option>
                  <option value="1">Every 1 Hour (Aggressive)</option>
                  <option value="6">Every 6 Hours (Recommended)</option>
                  <option value="12">Every 12 Hours (Standard)</option>
                  <option value="24">Every 24 Hours (Conservative)</option>
                </select>
                <p className="input-hint">How often the engine should scan all domains automatically.</p>
              </div>
            </div>
          )}

          {activeTab === 'im' && (
            <div className="tab-content">
              <div className="input-group">
                <label style={{ color: '#0088cc', fontWeight: 'bold' }}>Telegram Bot</label>
                <input 
                  type="password" 
                  placeholder="Bot Token" 
                  value={settings.telegram_token || ''}
                  onChange={e => setSettings({...settings, telegram_token: e.target.value})}
                  style={{ marginBottom: '8px' }}
                />
                <input 
                  type="text" 
                  placeholder="Chat ID" 
                  value={settings.telegram_chat_id || ''}
                  onChange={e => setSettings({...settings, telegram_chat_id: e.target.value})}
                />
              </div>

              <div className="input-group">
                <label style={{ color: '#5865F2', fontWeight: 'bold' }}>Discord Webhook</label>
                <input 
                  type="text" 
                  placeholder="https://discord.com/api/webhooks/..." 
                  value={settings.discord_webhook_url || ''}
                  onChange={e => setSettings({...settings, discord_webhook_url: e.target.value})}
                />
              </div>

              <div className="input-group">
                <label style={{ color: '#E01E5A', fontWeight: 'bold' }}>Slack Webhook</label>
                <input 
                  type="text" 
                  placeholder="https://hooks.slack.com/services/..." 
                  value={settings.slack_webhook_url || ''}
                  onChange={e => setSettings({...settings, slack_webhook_url: e.target.value})}
                />
              </div>
            </div>
          )}

          {activeTab === 'notifications' && (
            <div className="tab-content">
              <div className="input-group">
                <label style={{ color: '#EA4335', fontWeight: 'bold' }}>Email (SMTP) Alerts</label>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 80px', gap: '8px', marginBottom: '8px' }}>
                  <input 
                    type="text" 
                    placeholder="SMTP Host" 
                    value={settings.smtp_host || ''}
                    onChange={e => setSettings({...settings, smtp_host: e.target.value})}
                  />
                  <input 
                    type="text" 
                    placeholder="Port" 
                    value={settings.smtp_port || ''}
                    onChange={e => setSettings({...settings, smtp_port: e.target.value})}
                  />
                </div>
                <input 
                  type="text" 
                  placeholder="Username / Email" 
                  value={settings.smtp_user || ''}
                  onChange={e => setSettings({...settings, smtp_user: e.target.value})}
                  style={{ marginBottom: '8px' }}
                />
                <input 
                  type="password" 
                  placeholder="Password" 
                  value={settings.smtp_pass || ''}
                  onChange={e => setSettings({...settings, smtp_pass: e.target.value})}
                  style={{ marginBottom: '8px' }}
                />
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '8px' }}>
                  <input 
                    type="text" 
                    placeholder="From Address" 
                    value={settings.smtp_from || ''}
                    onChange={e => setSettings({...settings, smtp_from: e.target.value})}
                  />
                  <input 
                    type="text" 
                    placeholder="To Address" 
                    value={settings.smtp_to || ''}
                    onChange={e => setSettings({...settings, smtp_to: e.target.value})}
                  />
                </div>
              </div>

              <div style={{ height: '1px', background: 'var(--border)', margin: '1rem 0' }}></div>

              <div className="input-group">
                <label>Fallback Webhook</label>
                <div style={{ display: 'flex', gap: '8px' }}>
                  <input 
                    type="text" 
                    placeholder="https://webhook-url.com/..." 
                    value={settings.webhook_url}
                    onChange={e => setSettings({...settings, webhook_url: e.target.value})}
                  />
                  <button 
                    onClick={testWebhook} 
                    className="secondary" 
                    style={{ padding: '0 1rem' }}
                    disabled={testing}
                  >
                     <Send size={18} />
                  </button>
                </div>
              </div>
            </div>
          )}

          <div className="modal-footer">
            <button onClick={onClose} className="secondary">Cancel</button>
            <button onClick={saveSettings}>Save Configuration</button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SettingsModal;
