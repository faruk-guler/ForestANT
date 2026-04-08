import React from 'react';
import { Domain } from '../types';
import { AlertCircle, CheckCircle2, Clock } from 'lucide-react';

interface Props {
  domains: Domain[];
}

const DomainMiniGrid: React.FC<Props> = ({ domains }) => {
  // Sorting: Expired first, then Pending, then Healthy
  const sortedDomains = [...domains].sort((a, b) => {
    const priority = (d: Domain) => {
      if (d.status === 'expired' || d.status === 'critical') return 0;
      if (d.status === 'pending') return 1;
      return 2;
    };
    return priority(a) - priority(b);
  });

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'var(--status-green)';
      case 'critical': return 'var(--status-yellow)';
      case 'expired': return 'var(--status-red)';
      default: return 'var(--text-secondary)';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active': return <CheckCircle2 size={12} style={{ color: 'var(--status-green)' }} />;
      case 'critical': return <AlertCircle size={12} style={{ color: 'var(--status-yellow)' }} />;
      case 'expired': return <AlertCircle size={12} style={{ color: 'var(--status-red)' }} />;
      default: return <Clock size={12} style={{ color: 'var(--text-secondary)' }} />;
    }
  };

  return (
    <div className="mini-grid">
      {sortedDomains.map((d) => (
        <div key={d.id} className={`mini-card status-border-${d.status}`} title={`${d.hostname}\nStatus: ${d.status}`}>
          <div className="mini-card-header">
            {getStatusIcon(d.status)}
            <span className="mini-hostname">{d.hostname}</span>
          </div>
        </div>
      ))}
      {domains.length === 0 && (
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9rem', padding: '2rem', textAlign: 'center', gridColumn: '1 / -1' }}>
          No domains found. Add some in the Domains tab!
        </p>
      )}
    </div>
  );
};

export default DomainMiniGrid;
