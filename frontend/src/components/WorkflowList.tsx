import React, { useState } from 'react';
import { Plus, Play, Trash2, Zap, Settings, ShieldCheck } from 'lucide-react';
import type { Workflow, Domain, Access } from '../types';
import axios from 'axios';
import { useToast } from '../contexts/ToastContext';

interface Props {
  workflows: Workflow[];
  domains: Domain[];
  accessList: Access[];
  fetchWorkflows: () => void;
  apiUrl: string;
}

const WorkflowList: React.FC<Props> = ({ workflows, domains, accessList, fetchWorkflows, apiUrl }) => {
  const { addToast } = useToast();
  const [showAdd, setShowAdd] = useState(false);
  const [newName, setNewName] = useState('');
  const [domainId, setDomainId] = useState<number>(0);
  const [accessId, setAccessId] = useState<number>(0);
  const [type, setType] = useState<'deploy_ssh' | 'acme_http'>('deploy_ssh');

  const handleAdd = async () => {
    try {
      await axios.post(`${apiUrl}/workflows`, {
        name: newName,
        domain_id: domainId,
        access_id: accessId,
        type: type,
        config: JSON.stringify({ remote_path: '/etc/nginx/ssl' })
      });
      addToast('Workflow created', 'success');
      fetchWorkflows();
      setShowAdd(false);
    } catch (error) {
      addToast('Error creating workflow', 'error');
    }
  };

  const handleRun = async (id: number) => {
    try {
      await axios.post(`${apiUrl}/workflows/${id}/run`);
      addToast('Workflow started in the background', 'success');
      fetchWorkflows();
    } catch (error) {
      addToast('Error starting workflow', 'error');
    }
  };

  return (
    <div className="card" style={{ marginTop: '1rem' }}>
      <div className="card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1.5rem' }}>
        <h3 style={{ margin: 0 }}>Automation Workflows</h3>
        <button onClick={() => setShowAdd(!showAdd)} className="primary">
          <Plus size={18} /> {showAdd ? 'Cancel' : 'Create Workflow'}
        </button>
      </div>

      {showAdd && (
        <div style={{ padding: '1.5rem', borderBottom: '1px solid var(--border)', background: 'rgba(59, 130, 246, 0.02)' }}>
          <div className="grid-2">
            <div className="form-group">
              <label>Workflow Name</label>
              <input value={newName} onChange={e => setNewName(e.target.value)} placeholder="e.g. Deploy to Nginx" />
            </div>
            <div className="form-group">
              <label>Service Type</label>
              <select value={type} onChange={e => setType(e.target.value as any)}>
                <option value="deploy_ssh">SSH Deployment</option>
                <option value="acme_http">ACME HTTP-01 (Auto-Renew)</option>
              </select>
            </div>
          </div>
          <div className="grid-2" style={{ marginTop: '1rem' }}>
            <div className="form-group">
              <label>Target Domain</label>
              <select value={domainId} onChange={e => setDomainId(Number(e.target.value))}>
                <option value={0}>Select a domain...</option>
                {domains.map(d => <option key={d.id} value={d.id}>{d.hostname}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label>Used Credentials</label>
              <select value={accessId} onChange={e => setAccessId(Number(e.target.value))}>
                <option value={0}>Select access...</option>
                {accessList.map(a => <option key={a.id} value={a.id}>{a.name} ({a.provider})</option>)}
              </select>
            </div>
          </div>
          <button onClick={handleAdd} disabled={!newName || !domainId || !accessId} style={{ marginTop: '1.5rem' }}>Save & Enable</button>
        </div>
      )}

      <div className="table-responsive">
        <table>
          <thead>
            <tr>
              <th>Status</th>
              <th>Workflow Name</th>
              <th>Type</th>
              <th>Target</th>
              <th style={{ textAlign: 'right' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {workflows.length === 0 ? (
              <tr>
                <td colSpan={5} style={{ textAlign: 'center', padding: '3rem', opacity: 0.5 }}>No active workflows.</td>
              </tr>
            ) : workflows.map(w => (
              <tr key={w.id}>
                <td>
                  <span className={`status-badge ${w.status === 'success' ? 'status-green' : 'status-pending'}`}>
                    {w.status.toUpperCase()}
                  </span>
                </td>
                <td style={{ fontWeight: 600 }}>{w.name}</td>
                <td>
                  <div className="flex items-center gap-2">
                    <Zap size={14} className="text-accent" />
                    {w.type === 'deploy_ssh' ? 'SSH Web Server' : 'ACME Auto-Renew'}
                  </div>
                </td>
                <td>{domains.find(d => d.id === w.domain_id)?.hostname || 'Unknown'}</td>
                <td style={{ textAlign: 'right' }}>
                  <div className="actions" style={{ justifyContent: 'flex-end' }}>
                    <button onClick={() => handleRun(w.id)} className="btn-icon" title="Execute Workflow">
                      <Play size={16} />
                    </button>
                    <button className="btn-icon">
                      <Settings size={16} />
                    </button>
                    <button className="btn-icon text-danger">
                      <Trash2 size={16} />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default WorkflowList;
