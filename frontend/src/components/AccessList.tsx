import React, { useState } from 'react';
import { Plus, Trash2, Key, Globe, Layout, Shield } from 'lucide-react';
import type { Access } from '../types';
import axios from 'axios';
import { useToast } from '../contexts/ToastContext';

interface Props {
  accessList: Access[];
  fetchAccess: () => void;
  apiUrl: string;
}

const AccessList: React.FC<Props> = ({ accessList, fetchAccess, apiUrl }) => {
  const { addToast } = useToast();
  const [showAdd, setShowAdd] = useState(false);
  const [newName, setNewName] = useState('');
  const [provider, setProvider] = useState<'ssh' | 'dns_aliyun' | 'dns_cloudflare'>('ssh');
  const [config, setConfig] = useState('');

  const handleAdd = async () => {
    try {
      await axios.post(`${apiUrl}/access`, {
        name: newName,
        provider,
        config: config
      });
      addToast('Access created successfully', 'success');
      fetchAccess();
      setShowAdd(false);
      setNewName('');
      setConfig('');
    } catch (error) {
      addToast('Error creating access', 'error');
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure?')) return;
    try {
      await axios.delete(`${apiUrl}/access/${id}`);
      addToast('Access deleted', 'info');
      fetchAccess();
    } catch (error) {
      addToast('Error deleting access', 'error');
    }
  };

  return (
    <div className="card" style={{ marginTop: '1rem' }}>
      <div className="card-header" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '1.5rem' }}>
        <h3 style={{ margin: 0 }}>Saved Credentials</h3>
        <button onClick={() => setShowAdd(!showAdd)} className="primary">
          <Plus size={18} /> {showAdd ? 'Cancel' : 'Add New'}
        </button>
      </div>

      {showAdd && (
        <div style={{ padding: '1.5rem', borderBottom: '1px solid var(--border)', background: 'rgba(59, 130, 246, 0.02)' }}>
          <div className="grid-2">
            <div className="form-group">
              <label>Friendly Name</label>
              <input value={newName} onChange={e => setNewName(e.target.value)} placeholder="e.g. My Production VPS" />
            </div>
            <div className="form-group">
              <label>Provider Type</label>
              <select value={provider} onChange={e => setProvider(e.target.value as any)}>
                <option value="ssh">SSH Server</option>
                <option value="dns_aliyun">Aliyun DNS</option>
                <option value="dns_cloudflare">Cloudflare DNS</option>
              </select>
            </div>
          </div>
          <div className="form-group" style={{ marginTop: '1rem' }}>
            <label>Configuration (JSON)</label>
            <textarea 
              value={config} 
              onChange={e => setConfig(e.target.value)} 
              placeholder={provider === 'ssh' ? '{"host": "1.2.3.4", "user": "root"}' : '{"token": "xxx"}'}
              style={{ minHeight: '100px', fontFamily: 'monospace' }}
            />
          </div>
          <button onClick={handleAdd} disabled={!newName || !config} style={{ marginTop: '1rem' }}>Save Access</button>
        </div>
      )}

      <div className="table-responsive">
        <table>
          <thead>
            <tr>
              <th>Name</th>
              <th>Provider</th>
              <th>Created</th>
              <th style={{ textAlign: 'right' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {accessList.length === 0 ? (
              <tr>
                <td colSpan={4} style={{ textAlign: 'center', padding: '3rem', opacity: 0.5 }}>No credentials stored.</td>
              </tr>
            ) : accessList.map(item => (
              <tr key={item.id}>
                <td style={{ fontWeight: 600 }}>{item.name}</td>
                <td>
                  <div className="flex items-center gap-2">
                    {item.provider === 'ssh' ? <Layout size={16} /> : <Globe size={16} />}
                    {item.provider}
                  </div>
                </td>
                <td style={{ fontSize: '0.85rem' }}>{new Date(item.created_at).toLocaleDateString()}</td>
                <td style={{ textAlign: 'right' }}>
                  <button onClick={() => handleDelete(item.id)} className="btn-icon text-danger">
                    <Trash2 size={16} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default AccessList;
