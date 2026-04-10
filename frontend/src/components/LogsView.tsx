import React, { useEffect, useState } from 'react';
import axios from 'axios';
import { Clock, CheckCircle, AlertCircle, RefreshCw } from 'lucide-react';

interface ScanLog {
  id: number;
  hostname: string;
  type: string;
  status: string;
  message: string;
  created_at: string;
}

interface Props {
  apiUrl: string;
}

const LogsView: React.FC<Props> = ({ apiUrl }) => {
  const [logs, setLogs] = useState<ScanLog[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchLogs = async () => {
    setLoading(true);
    try {
      const { data } = await axios.get(`${apiUrl}/logs`);
      setLogs(data);
    } catch (error) {
      console.error('Error fetching logs:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchLogs();
  }, []);

  const clearLogs = async () => {
    if (!confirm('Are you sure you want to permanently delete all scan logs?')) return;
    try {
      await axios.delete(`${apiUrl}/logs`);
      setLogs([]);
    } catch (error) {
      console.error('Error clearing logs:', error);
    }
  };

  return (
    <div className="logs-view animate-fade-in">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.5rem' }}>
        <h2 style={{ fontSize: '1.5rem', fontWeight: 700 }}>System Logs</h2>
        <div style={{ display: 'flex', gap: '0.75rem' }}>
          <button className="secondary" style={{ padding: '0.5rem 1rem', fontSize: '0.85rem' }} onClick={clearLogs} disabled={logs.length === 0}>
             Clear Logs
          </button>
          <button className="secondary btn-icon" onClick={fetchLogs} disabled={loading}>
            <RefreshCw className={loading ? 'pulse' : ''} size={18} />
          </button>
        </div>
      </div>

      <div className="card table-container">
        <div className="table-responsive">
          <table>
            <thead>
              <tr>
                <th>Time</th>
                <th>Target</th>
                <th>Check Type</th>
                <th>Status</th>
                <th>Details</th>
              </tr>
            </thead>
            <tbody>
              {logs.length === 0 ? (
                <tr>
                  <td colSpan={5} style={{ textAlign: 'center', padding: '3rem', opacity: 0.5 }}>
                    No system logs available yet.
                  </td>
                </tr>
              ) : logs.map((log) => (
                <tr key={log.id}>
                  <td style={{ whiteSpace: 'nowrap', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
                    {new Date(log.created_at).toLocaleString()}
                  </td>
                  <td style={{ fontWeight: 600 }}>{log.hostname}</td>
                  <td>
                    <span className="status-badge status-pending" style={{ fontSize: '0.7rem' }}>
                      {log.type}
                    </span>
                  </td>
                  <td>
                    {log.status === 'success' ? (
                      <span className="text-success flex items-center gap-1">
                        <CheckCircle size={14} /> Success
                      </span>
                    ) : (
                      <span className="text-danger flex items-center gap-1">
                        <AlertCircle size={14} /> Error
                      </span>
                    )}
                  </td>
                  <td style={{ fontSize: '0.85rem' }}>{log.message}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default LogsView;
