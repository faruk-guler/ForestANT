import React, { useState, useRef } from 'react';
import { Plus, Upload } from 'lucide-react';

interface Props {
  onAddDomain: (hostname: string) => Promise<void>;
  onBulkAdd?: (domains: string[]) => Promise<void>;
}

const DomainForm: React.FC<Props> = ({ onAddDomain, onBulkAdd }) => {
  const [newHostname, setNewHostname] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newHostname.trim()) return;
    await onAddDomain(newHostname.trim());
    setNewHostname('');
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !onBulkAdd) return;

    const reader = new FileReader();
    reader.onload = async (event) => {
      const text = event.target?.result as string;
      if (text) {
        const domains = text.split(/\r?\n/).map(d => d.trim()).filter(d => d.length > 0);
        if (domains.length > 0) {
           await onBulkAdd(domains);
        }
      }
      if (fileInputRef.current) fileInputRef.current.value = '';
    };
    reader.readAsText(file);
  };

  return (
    <div className="card" style={{ marginBottom: '2rem' }}>
      <form onSubmit={handleSubmit} className="form-group" style={{ display: 'flex', gap: '0.75rem' }}>
        <input
          type="text"
          placeholder="Search or add hostname (e.g., example.com)"
          value={newHostname}
          onChange={(e) => setNewHostname(e.target.value)}
        />
        <button type="submit">
          <Plus size={18} /> Add Target
        </button>
        {onBulkAdd && (
          <button type="button" className="secondary" onClick={() => fileInputRef.current?.click()} title="Import from TXT">
            <Upload size={18} /> TXT
          </button>
        )}
        <input 
          type="file" 
          accept=".txt" 
          ref={fileInputRef} 
          style={{ display: 'none' }} 
          onChange={handleFileUpload} 
        />
      </form>
    </div>
  );
};

export default DomainForm;
