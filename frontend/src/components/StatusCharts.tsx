import React, { useMemo } from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import { Domain } from '../types';
import { differenceInDays } from 'date-fns';

interface Props {
  domains: Domain[];
  stats: {
    healthy: number;
    critical: number;
    expired: number;
  } | null;
}

const StatusCharts: React.FC<Props> = ({ domains, stats }) => {
  const data = useMemo(() => {
    if (!stats) return [];
    return [
      { name: 'Healthy', value: stats?.healthy || 0, color: '#3fb950' },
      { name: 'Critical', value: stats?.critical || 0, color: '#e3b341' },
      { name: 'Expired', value: stats?.expired || 0, color: '#ff4d4d' },
    ].filter(item => item.value > 0);
  }, [stats]);

  if (domains.length === 0) return null;

  return (
    <div className="card" style={{ height: '100%', display: 'flex', flexDirection: 'column', minHeight: '160px' }}>
      <h3 style={{ fontSize: '0.9rem', color: 'var(--text-secondary)', marginBottom: '1rem', textTransform: 'uppercase', letterSpacing: '1px' }}>Fleet Health Overview</h3>
      <div style={{ flex: 1, width: '100%' }}>
        <ResponsiveContainer width="100%" height="100%">
          <PieChart margin={{ top: 5, bottom: 5 }}>
            <Pie
              data={data}
              innerRadius={45}
              outerRadius={65}
              paddingAngle={5}
              dataKey="value"
              cx="35%"
              cy="50%"
              animationBegin={0}
              animationDuration={800}
            >
              {data.map((entry, index) => (
                <Cell key={`cell-${index}`} fill={entry.color} stroke="none" />
              ))}
            </Pie>
            <Tooltip 
              contentStyle={{ background: 'var(--card-bg)', border: '1px solid var(--border)', borderRadius: '12px', fontSize: '0.8rem' }}
              itemStyle={{ color: '#fff' }}
            />
            <Legend 
              layout="vertical"
              verticalAlign="middle" 
              align="right"
              iconType="circle"
              iconSize={8}
              wrapperStyle={{ paddingLeft: '20px', fontSize: '0.8rem', lineHeight: '2' }}
            />
          </PieChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
};

export default StatusCharts;
