import React, { useMemo, useState } from 'react';
import { ChevronUp, ChevronDown, Globe, Edit2, RefreshCw, Trash2, CheckCircle, Info, Shield, AlertCircle, Clock } from 'lucide-react';
import { differenceInDays } from 'date-fns';
import type { Domain } from '../types';

interface Props {
  domains: Domain[];
  settingsThreshold: string;
  updateDomain: (id: number, hostname: string) => void;
  triggerScan: (id: number) => void;
  deleteDomain?: (id: number) => void;
  bulkDeleteDomains?: (ids: number[]) => void;
  showActions?: boolean;
}

const DomainTable: React.FC<Props> = ({ 
  domains, 
  settingsThreshold, 
  updateDomain, 
  triggerScan, 
  deleteDomain, 
  bulkDeleteDomains,
  showActions = true 
}) => {
  const [sortConfig, setSortConfig] = useState<{ key: keyof Domain, direction: 'asc' | 'desc' } | null>(null);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [editingHostname, setEditingHostname] = useState('');
  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [filter, setFilter] = useState<'all' | 'expired' | 'critical' | 'healthy'>('all');
  const [searchQuery, setSearchQuery] = useState('');

  const requestSort = (key: keyof Domain) => {
    let direction: 'asc' | 'desc' = 'asc';
    if (sortConfig && sortConfig.key === key && sortConfig.direction === 'asc') direction = 'desc';
    setSortConfig({ key, direction });
  };

  const filteredAndSortedDomains = useMemo(() => {
    let items = domains.filter(d => 
      d.hostname.toLowerCase().includes(searchQuery.toLowerCase())
    );
    
    // Applying Filter
    if (filter !== 'all') {
      items = items.filter(d => {
        const sslDays = d.ssl_expiry ? differenceInDays(new Date(d.ssl_expiry), new Date()) : 100;
        const domDays = d.domain_expiry ? differenceInDays(new Date(d.domain_expiry), new Date()) : 100;
        
        if (filter === 'expired') return sslDays < 0 || domDays < 0;
        if (filter === 'critical') return (sslDays >= 0 && sslDays < 30) || (domDays >= 0 && domDays < 30);
        if (filter === 'healthy') return sslDays >= 30 && domDays >= 30;
        return true;
      });
    }

    // Applying Sort
    if (sortConfig !== null) {
      items.sort((a: any, b: any) => {
        let aValue = a[sortConfig.key];
        let bValue = b[sortConfig.key];

        if (aValue === null) return 1;
        if (bValue === null) return -1;
        if (aValue < bValue) return sortConfig.direction === 'asc' ? -1 : 1;
        if (aValue > bValue) return sortConfig.direction === 'asc' ? 1 : -1;
        return 0;
      });
    }
    return items;
  }, [domains, sortConfig, filter, searchQuery]); // FIX: searchQuery bağımlılığı eklendi

  const handleSelectAll = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.checked) {
      setSelectedIds(filteredAndSortedDomains.map(d => d.id));
    } else {
      setSelectedIds([]);
    }
  };

  const handleSelectOne = (id: number) => {
    if (selectedIds.includes(id)) {
      setSelectedIds(selectedIds.filter(selectedId => selectedId !== id));
    } else {
      setSelectedIds([...selectedIds, id]);
    }
  };

  const getStatusClass = (expiry: string | null) => {
    if (!expiry) return 'status-pending';
    const days = differenceInDays(new Date(expiry), new Date());
    if (days < 0) return 'status-red';
    if (days < parseInt(settingsThreshold)) return 'status-yellow';
    return 'status-green';
  };

  const formatExpiry = (expiry: string | null) => {
    if (!expiry) return 'Unknown';
    const date = new Date(expiry);
    const days = differenceInDays(date, new Date());
    const isExpired = days < 0;
    
    return (
      <div className="expiry-info">
        <div className="date">{date.toLocaleDateString()}</div>
        <div className={`days-left ${isExpired ? 'text-danger' : ''}`}>
          {isExpired ? (
            <span className="flex items-center gap-1"><AlertCircle size={12} /> Expired {Math.abs(days)} days ago</span>
          ) : (
            <span>{days} days remaining</span>
          )}
        </div>
      </div>
    );
  };

  return (
    <div className="card table-container">
      {showActions && selectedIds.length > 0 && (
         <div style={{ padding: '1rem 1.5rem', borderBottom: '1px solid var(--border)', background: 'rgba(239, 68, 68, 0.05)', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <span style={{ fontSize: '0.9rem', fontWeight: 600 }}>{selectedIds.length} item(s) selected</span>
            <button className="secondary text-danger" style={{ padding: '0.5rem 1rem' }} onClick={() => { bulkDeleteDomains?.(selectedIds); setSelectedIds([]); }}>
              <Trash2 size={16} /> Delete Selected
            </button>
         </div>
      )}

      {/* Filter Tabs */}
      <div style={{ padding: '1rem 1.5rem', borderBottom: '1px solid var(--border)', display: 'flex', gap: '1rem', flexWrap: 'wrap', alignItems: 'center' }}>
        <div style={{ position: 'relative', flex: '1', minWidth: '240px' }}>
          <input 
            type="text" 
            placeholder="Search domains..." 
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{ paddingLeft: '2.8rem', borderRadius: '12px', fontSize: '0.9rem' }}
          />
          <Globe size={18} style={{ position: 'absolute', left: '1rem', top: '50%', transform: 'translateY(-50%)', opacity: 0.4 }} />
        </div>
        
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <button className={`filter-btn ${filter === 'all' ? 'active' : ''}`} onClick={() => setFilter('all')}>All ({domains.length})</button>
          <button className={`filter-btn ${filter === 'expired' ? 'active' : ''}`} onClick={() => setFilter('expired')}>Expired</button>
          <button className={`filter-btn ${filter === 'critical' ? 'active' : ''}`} onClick={() => setFilter('critical')}>Critical</button>
          <button className={`filter-btn ${filter === 'healthy' ? 'active' : ''}`} onClick={() => setFilter('healthy')}>Healthy</button>
        </div>
      </div>
      <div className="table-responsive">
        <table>
          <thead>
            <tr>
              {showActions && (
                <th style={{ width: '40px', paddingRight: '0' }}>
                  <input 
                    type="checkbox" 
                    checked={selectedIds.length === filteredAndSortedDomains.length && filteredAndSortedDomains.length > 0} 
                    onChange={handleSelectAll} 
                    style={{ width: '16px', height: '16px', cursor: 'pointer' }}
                  />
                </th>
              )}
              <th onClick={() => requestSort('hostname')} className="sortable">
                <div className="th-content">
                  Domain {sortConfig?.key === 'hostname' && (sortConfig.direction === 'asc' ? <ChevronUp size={14} /> : <ChevronDown size={14} />)}
                </div>
              </th>
              <th onClick={() => requestSort('ssl_expiry')} className="sortable">
                <div className="th-content">
                  SSL Status {sortConfig?.key === 'ssl_expiry' && (sortConfig.direction === 'asc' ? <ChevronUp size={14} /> : <ChevronDown size={14} />)}
                </div>
              </th>
              <th onClick={() => requestSort('domain_expiry')} className="sortable">
                <div className="th-content">
                  Domain Registry {sortConfig?.key === 'domain_expiry' && (sortConfig.direction === 'asc' ? <ChevronUp size={14} /> : <ChevronDown size={14} />)}
                </div>
              </th>
              {showActions && <th style={{ textAlign: 'right' }}>Actions</th>}
            </tr>
          </thead>
          <tbody>
            {filteredAndSortedDomains.length === 0 ? (
              <tr>
                <td colSpan={showActions ? 5 : 3} style={{ textAlign: 'center', padding: '3rem', opacity: 0.5 }}>
                  <Info size={40} style={{ marginBottom: '0.75rem', margin: '0 auto', opacity: 0.3 }} />
                  <div style={{ fontSize: '0.9rem' }}>No targets discovered yet.</div>
                </td>
              </tr>
            ) : filteredAndSortedDomains.map((domain) => (
              <tr key={domain.id} className={`domain-row ${selectedIds.includes(domain.id) ? 'selected' : ''}`}>
                {showActions && (
                  <td style={{ width: '40px', paddingRight: '0' }}>
                    <input 
                      type="checkbox" 
                      checked={selectedIds.includes(domain.id)}
                      onChange={() => handleSelectOne(domain.id)}
                      style={{ width: '16px', height: '16px', cursor: 'pointer', accentColor: 'var(--accent-color)' }}
                    />
                  </td>
                )}
                <td>
                  {editingId === domain.id ? (
                    <div className="form-group" style={{ margin: 0 }}>
                      <input
                        type="text"
                        value={editingHostname}
                        onChange={(e) => setEditingHostname(e.target.value)}
                        autoFocus
                        style={{ padding: '0.5rem 0.75rem', fontSize: '0.9rem' }}
                        onBlur={() => setEditingId(null)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') {
                            updateDomain(domain.id, editingHostname);
                            setEditingId(null);
                          }
                        }}
                      />
                    </div>
                  ) : (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '2px' }}>
                      <div className="hostname" style={{ fontWeight: 600, fontSize: '0.95rem' }}>{domain.hostname}</div>
                      <div className="scan-time" style={{ display: 'flex', alignItems: 'center', gap: '4px', fontSize: '0.75rem', opacity: 0.6 }}>
                        <Clock size={10} />
                        {domain.last_scan ? new Date(domain.last_scan).toLocaleString(undefined, { dateStyle: 'short', timeStyle: 'short' }) : 'Pending'}
                      </div>
                    </div>
                  )}
                </td>
                <td style={{ verticalAlign: 'middle' }}>
                  <div className={`status-badge ${getStatusClass(domain.ssl_expiry)}`} style={{ padding: '0.2rem 0.6rem', borderRadius: '6px', fontSize: '0.75rem' }}>
                    <div className="flex items-center gap-1">
                      <Shield size={10} />
                      {domain.ssl_expiry ? (differenceInDays(new Date(domain.ssl_expiry), new Date()) < 0 ? 'EXPIRED' : 'ACTIVE') : 'PENDING'}
                    </div>
                  </div>
                  <div style={{ marginTop: '0.35rem' }}>{formatExpiry(domain.ssl_expiry)}</div>
                </td>
                <td style={{ verticalAlign: 'middle' }}>
                  <div className={`status-badge ${getStatusClass(domain.domain_expiry)}`} style={{ padding: '0.2rem 0.6rem', borderRadius: '6px', fontSize: '0.75rem' }}>
                    <div className="flex items-center gap-1">
                      <Globe size={10} />
                      {domain.domain_expiry ? (differenceInDays(new Date(domain.domain_expiry), new Date()) < 0 ? 'EXPIRED' : 'VALID') : 'PENDING'}
                    </div>
                  </div>
                  <div style={{ marginTop: '0.35rem' }}>{formatExpiry(domain.domain_expiry)}</div>
                </td>
                {showActions && (
                  <td style={{ textAlign: 'right', verticalAlign: 'middle' }}>
                    <div className="actions" style={{ justifyContent: 'flex-end', display: 'flex', gap: '4px' }}>
                      {editingId === domain.id ? (
                        <button onClick={() => { updateDomain(domain.id, editingHostname); setEditingId(null); }} className="btn-icon text-success" style={{ padding: '4px' }}>
                          <CheckCircle size={18} />
                        </button>
                      ) : (
                        <>
                          <button onClick={() => { setEditingId(domain.id); setEditingHostname(domain.hostname); }} className="btn-icon" style={{ padding: '6px' }} title="Edit">
                            <Edit2 size={14} /> 
                          </button>
                          <button onClick={() => triggerScan(domain.id)} className="btn-icon" style={{ padding: '6px' }} title="Rescan">
                            <RefreshCw size={14} />
                          </button>
                          <button onClick={() => deleteDomain?.(domain.id)} className="btn-icon hover-danger" style={{ padding: '6px' }} title="Delete">
                            <Trash2 size={14} />
                          </button>
                        </>
                      )}
                    </div>
                  </td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export default DomainTable;
