import React, { useState, useEffect, useCallback } from 'react';
import axios from 'axios';
import DashboardStats from './components/DashboardStats';
import DomainForm from './components/DomainForm';
import DomainTable from './components/DomainTable';
import StatusCharts from './components/StatusCharts';
import LogsView from './components/LogsView';
import AccessList from './components/AccessList';
import WorkflowList from './components/WorkflowList';
import SettingsModal from './components/SettingsModal';
import ConfirmModal from './components/ConfirmModal';
import { useToast } from './contexts/ToastContext';
import type { Domain, Settings, Access, Workflow } from './types';
import { useTheme } from './contexts/ThemeContext';
import { 
  LayoutDashboard, Globe, ListTree, RefreshCw, Settings as SettingsIcon, 
  Sun, Moon, Info, ShieldCheck, Code 
} from 'lucide-react';
import logo from './assets/logo.png';

const API_URL = (import.meta.env.VITE_API_URL || '/api');

function App() {
  const { addToast } = useToast();
  const { theme, toggleTheme } = useTheme();
  const [domains, setDomains] = useState<Domain[]>([]);
  const [loading, setLoading] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [activeTab, setActiveTab] = useState<'dashboard' | 'domains' | 'logs' | 'access' | 'workflows' | 'about'>('dashboard');
  const [settings, setSettings] = useState<Settings>({ webhook_url: '', notification_threshold: '30' });
  const [accessList, setAccessList] = useState<Access[]>([]);
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [stats, setStats] = useState<any>(null);
  const [confirmModal, setConfirmModal] = useState<{
    show: boolean;
    title: string;
    message: string;
    onConfirm: () => void;
    confirmText?: string;
  }>({ show: false, title: '', message: '', onConfirm: () => {} });

  const fetchDashboardData = useCallback(async (isSilent = false) => {
    if (!isSilent) setLoading(true);
    try {
      const { data } = await axios.get(`${API_URL}/dashboard`);
      setDomains(data.domains);
      setSettings(data.settings);
      setStats(data.stats);
    } catch (error) {
      console.error('Fetch error:', error);
    } finally {
      if (!isSilent) setLoading(false);
    }
  }, []);

  const fetchAccess = useCallback(async () => {
    try {
      const { data } = await axios.get(`${API_URL}/access`);
      setAccessList(data);
    } catch (error) {
      console.error('Access fetch error:', error);
    }
  }, []);

  const fetchWorkflows = useCallback(async () => {
    try {
      const { data } = await axios.get(`${API_URL}/workflows`);
      setWorkflows(data);
    } catch (error) {
      console.error('Workflows fetch error:', error);
    }
  }, []);

  useEffect(() => {
    fetchDashboardData();
    fetchAccess();
    fetchWorkflows();

    const eventSource = new EventSource(`${API_URL}/events`);
    eventSource.onmessage = (event) => {
      if (event.data === 'reload') fetchDashboardData(true);
    };
    return () => eventSource.close();
  }, [fetchDashboardData, fetchAccess, fetchWorkflows]);

  const addDomain = async (hostname: string) => {
    const domainRegex = /^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$/;
    if (!domainRegex.test(hostname)) {
      addToast('Invalid domain name format.', 'error');
      return;
    }
    try {
      await axios.post(`${API_URL}/domains`, { hostname });
      addToast(`Added ${hostname} successfully.`, 'success');
      fetchDashboardData(true);
    } catch (error: any) {
      addToast(error.response?.data?.error || 'Error adding domain', 'error');
    }
  };

  const bulkAddDomains = async (domainsToUpload: string[]) => {
    try {
      setLoading(true);
      const { data } = await axios.post(`${API_URL}/domains/bulk`, { domains: domainsToUpload });
      addToast(`Imported ${data.added} domains.`, 'success');
      fetchDashboardData(true);
    } catch (error) {
      addToast('Error importing domains in bulk.', 'error');
    } finally {
      setLoading(false);
    }
  };

  const updateDomain = async (id: number, hostname: string) => {
    try {
      await axios.put(`${API_URL}/domains/${id}`, { hostname });
      addToast(`Updated domain to ${hostname}`, 'success');
      fetchDashboardData(true);
    } catch (error) {
      addToast('Error updating domain', 'error');
    }
  };

  const deleteDomain = async (id: number) => {
    setConfirmModal({
      show: true,
      title: 'Delete Domain',
      message: 'Are you sure you want to delete this domain? This action cannot be undone.',
      confirmText: 'Delete',
      onConfirm: async () => {
        try {
          await axios.delete(`${API_URL}/domains/${id}`);
          addToast('Domain deleted.', 'info');
          fetchDashboardData(true);
        } catch (error) {
          addToast('Error deleting domain', 'error');
        } finally {
          setConfirmModal(prev => ({ ...prev, show: false }));
        }
      }
    });
  };

  const bulkDeleteDomains = async (ids: number[]) => {
    setConfirmModal({
      show: true,
      title: 'Bulk Delete',
      message: `Are you sure you want to delete ${ids.length} selected domains? This action cannot be undone.`,
      confirmText: 'Delete All',
      onConfirm: async () => {
        try {
          setLoading(true);
          await axios.post(`${API_URL}/domains/bulk-delete`, { ids });
          addToast(`Deleted ${ids.length} domains.`, 'info');
          fetchDashboardData(true);
        } catch (error) {
          addToast('Error performing bulk delete.', 'error');
        } finally {
          setLoading(false);
          setConfirmModal(prev => ({ ...prev, show: false }));
        }
      }
    });
  };

  const triggerScan = async (id?: number) => {
    setLoading(true);
    try {
      const url = id ? `${API_URL}/scan/${id}` : `${API_URL}/scan`;
      const { data } = await axios.post(url);
      addToast(data.message, 'success');
      fetchDashboardData(true);
    } catch (error) {
      addToast('Error triggering scan', 'error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app-layout">
      <aside className="sidebar">
        <div className="sidebar-logo" onClick={() => setActiveTab('dashboard')} style={{ cursor: 'pointer' }}>
          <img src={logo} alt="ForestANT" className="logo-img" />
          <div className="logo-text">
            <h1>ForestANT</h1>
            <span>Version 3.0</span>
          </div>
        </div>
        
        <div style={{ marginBottom: '1.25rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between', paddingLeft: '0.2rem' }}>
           <p style={{ color: 'var(--text-secondary)', fontSize: '0.8rem', fontWeight: '500', opacity: 0.7 }}>SSL & Domain Automation</p>
           <div 
             className="nm-toggle-wrapper" 
             onClick={toggleTheme}
             data-theme-state={theme}
             title={theme === 'dark' ? 'Switch to Light Mode' : 'Switch to Dark Mode'}
           >
             <div className="nm-toggle-knob">
               {theme === 'dark' ? <Moon size={16} /> : <Sun size={16} />}
             </div>
           </div>
        </div>

        <nav className="sidebar-nav">
          <button className={`nav-item ${activeTab === 'dashboard' ? 'active' : ''}`} onClick={() => setActiveTab('dashboard')}><LayoutDashboard size={18} /> Dashboard</button>
          <button className={`nav-item ${activeTab === 'domains' ? 'active' : ''}`} onClick={() => setActiveTab('domains')}><Globe size={18} /> Domains</button>
          <button className={`nav-item ${activeTab === 'logs' ? 'active' : ''}`} onClick={() => setActiveTab('logs')}><ListTree size={18} /> System Logs</button>
          <button className={`nav-item ${activeTab === 'access' ? 'active' : ''}`} onClick={() => setActiveTab('access')}><SettingsIcon size={18} /> Credentials</button>
          <button className={`nav-item ${activeTab === 'workflows' ? 'active' : ''}`} onClick={() => setActiveTab('workflows')}><RefreshCw size={18} /> Workflows</button>
          <button className={`nav-item ${activeTab === 'about' ? 'active' : ''}`} onClick={() => setActiveTab('about')}><Info size={18} /> About</button>
        </nav>
        
        <div className="sidebar-footer" style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
          <button onClick={() => triggerScan()} className="secondary pulse-hover" style={{ width: '100%', justifyContent: 'center' }} disabled={loading}>
            <RefreshCw className={loading ? 'pulse' : ''} size={16} />
            {loading ? 'Scanning...' : 'Scan All Now'}
          </button>
          <button onClick={() => setShowSettings(true)} className="secondary" style={{ width: '100%', justifyContent: 'center' }}>
            <SettingsIcon size={16} /> Settings
          </button>
        </div>
      </aside>

      <main className="main-content">
        {activeTab === 'dashboard' ? (
          <div className="tab-fade-in dashboard-overview">
            <div className="fleet-grid">
              <DashboardStats stats={stats} />
              <StatusCharts stats={stats} domains={domains} />
            </div>
            <div style={{ marginTop: '2rem' }}>
              <div className="section-header" style={{ marginBottom: '1rem' }}>
                <h3>Asset Status Reports</h3>
                <p>Full monitoring report for all assets.</p>
              </div>
              <DomainTable domains={domains} settingsThreshold={settings.notification_threshold} updateDomain={updateDomain} triggerScan={triggerScan} showActions={false} />
            </div>
          </div>
        ) : activeTab === 'domains' ? (
          <div className="tab-fade-in content-container">
            <div className="section-header" style={{ marginBottom: '2rem' }}>
              <h2>Domain Management</h2>
              <p>Add, edit, delete or bulk import assets.</p>
            </div>
            <div style={{ marginBottom: '2rem' }}>
              <DomainForm onAddDomain={addDomain} onBulkAdd={bulkAddDomains} />
            </div>
            <DomainTable domains={domains} settingsThreshold={settings.notification_threshold} updateDomain={updateDomain} deleteDomain={deleteDomain} bulkDeleteDomains={bulkDeleteDomains} triggerScan={triggerScan} />
          </div>
        ) : activeTab === 'logs' ? (
          <LogsView apiUrl={API_URL} />
        ) : activeTab === 'access' ? (
          <div style={{ padding: '0 2rem 2rem 2rem' }}>
            <div className="section-header"><h2>Access Management</h2><p>Securely store target credentials.</p></div>
            <AccessList accessList={accessList} fetchAccess={fetchAccess} apiUrl={API_URL} />
          </div>
        ) : activeTab === 'workflows' ? (
          <div style={{ padding: '0 2rem 2rem 2rem' }}>
            <div className="section-header"><h2>Automation Pipelines</h2><p>Automated certificate & domain workflows.</p></div>
            <WorkflowList workflows={workflows} domains={domains} accessList={accessList} fetchWorkflows={fetchWorkflows} apiUrl={API_URL} />
          </div>
        ) : (
          <div className="about-container tab-fade-in">
            <div className="card about-card-minimal">
              <div className="about-content">
                <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem', marginBottom: '2rem' }}>
                  <ShieldCheck size={32} style={{ color: 'var(--accent-color)' }} />
                  <h2 style={{ fontSize: '2rem', fontWeight: '800', letterSpacing: '-0.5px' }}>ForestANT v3.0</h2>
                </div>
                <p style={{ fontSize: '1.1rem', color: 'var(--text-primary)', marginBottom: '1.5rem', fontWeight: '500' }}>
                  A high-performance monitoring ecosystem built with Go and React.
                </p>
                <div className="mission-text-minimal">
                  ForestANT provides real-time oversight of SSL/TLS certificate lifecycles 
                  and domain fleet health, ensuring absolute continuity for your digital infrastructure.
                </div>
                <div style={{ margin: '2.5rem 0', display: 'flex', flexDirection: 'column', gap: '1rem' }}>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.8rem', color: 'var(--text-secondary)' }}>
                    <div style={{ width: '6px', height: '6px', borderRadius: '50%', background: 'var(--accent-color)' }}></div>
                    <span>Backend powered by <strong>Go Fiber</strong></span>
                  </div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.8rem', color: 'var(--text-secondary)' }}>
                    <div style={{ width: '6px', height: '6px', borderRadius: '50%', background: 'var(--accent-color)' }}></div>
                    <span>Native High-Availability Windows Deployment</span>
                  </div>
                </div>
                <div className="developer-section-minimal">
                  <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '1rem' }}>
                    <Code size={18} />
                    <span style={{ fontWeight: '600' }}>Developer: faruk-guler</span>
                  </div>
                  <div className="social-links-minimal" style={{ flexDirection: 'column', gap: '0.75rem' }}>
                    <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                      <span style={{ fontWeight: '600', minWidth: '80px' }}>GitHub:</span>
                      <a href="https://github.com/faruk-guler" target="_blank" rel="noopener noreferrer" className="social-link-item">github.com/faruk-guler</a>
                    </div>
                    <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
                      <span style={{ fontWeight: '600', minWidth: '80px' }}>Website:</span>
                      <a href="#" target="_blank" rel="noopener noreferrer" className="social-link-item">Kendi Sitem</a>
                    </div>
                  </div>
                </div>
                <div style={{ marginTop: '4rem', paddingTop: '2rem', borderTop: '1px solid var(--border)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', opacity: 0.7 }}>
                  <div style={{ fontSize: '0.85rem', color: 'var(--text-secondary)' }}>© 2026 ForestANT Monitoring • All Rights Reserved</div>
                  <div style={{ display: 'flex', gap: '0.5rem' }}>
                     <div style={{ width: '8px', height: '8px', background: 'var(--success)', borderRadius: '50%' }}></div>
                     <span style={{ fontSize: '0.75rem', fontWeight: '700', textTransform: 'uppercase', letterSpacing: '1px' }}>System Operational</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}
      </main>

      {showSettings && (
        <SettingsModal settings={settings} setSettings={setSettings} onClose={() => setShowSettings(false)} apiUrl={API_URL} />
      )}

      <ConfirmModal 
        show={confirmModal.show}
        title={confirmModal.title}
        message={confirmModal.message}
        confirmText={confirmModal.confirmText}
        onConfirm={confirmModal.onConfirm}
        onCancel={() => setConfirmModal(prev => ({ ...prev, show: false }))}
      />
    </div>
  );
}

export default App;
