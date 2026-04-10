import React, { useMemo } from 'react';
import { Globe, CheckCircle, AlertTriangle } from 'lucide-react';
import type { Domain } from '../types';
import { differenceInDays } from 'date-fns';

interface Props {
  stats: {
    total: number;
    healthy: number;
    critical: number;
    expired: number;
  } | null;
}

const DashboardStats: React.FC<Props> = ({ stats }) => {
  if (!stats) return (
     <div className="summary-grid">
        {[1,2,3,4].map(i => <div key={i} className="card stat-card pulse" style={{ height: '84px', opacity: 0.5 }}></div>)}
     </div>
  );

  return (
    <>
      <div className="summary-grid">
      <div className="card stat-card">
        <div className="stat-icon" style={{ background: 'rgba(59, 130, 246, 0.1)', color: 'var(--accent-color)' }}>
          <Globe size={24} />
        </div>
        <div>
          <div className="stat-label">Total</div>
          <div className="stat-value">{stats?.total || 0}</div>
        </div>
      </div>
      <div className="card stat-card">
        <div className="stat-icon" style={{ background: 'rgba(16, 185, 129, 0.1)', color: 'var(--success)' }}>
          <CheckCircle size={24} />
        </div>
        <div>
          <div className="stat-label">Healthy</div>
          <div className="stat-value">{stats?.healthy || 0}</div>
        </div>
      </div>
      <div className="card stat-card">
        <div className="stat-icon" style={{ background: 'rgba(227, 179, 65, 0.1)', color: 'var(--warning)' }}>
          <AlertTriangle size={24} />
        </div>
        <div>
          <div className="stat-label">Critical</div>
          <div className="stat-value">{stats?.critical || 0}</div>
        </div>
      </div>
      <div className="card stat-card">
        <div className="stat-icon" style={{ background: 'rgba(239, 68, 68, 0.1)', color: 'var(--danger)' }}>
          <AlertTriangle size={24} />
        </div>
        <div>
          <div className="stat-label">Expired</div>
          <div className="stat-value">{stats?.expired || 0}</div>
        </div>
      </div>
    </div>
        </>
  );
};

export default DashboardStats;
