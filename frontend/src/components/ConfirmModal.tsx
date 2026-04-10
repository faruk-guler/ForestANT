import React from 'react';
import { X, AlertTriangle } from 'lucide-react';

interface Props {
  show: boolean;
  title: string;
  message: string;
  onConfirm: () => void;
  onCancel: () => void;
  confirmText?: string;
  variant?: 'danger' | 'warning' | 'info';
}

const ConfirmModal: React.FC<Props> = ({ 
  show, 
  title, 
  message, 
  onConfirm, 
  onCancel, 
  confirmText = 'Confirm',
  variant = 'danger'
}) => {
  if (!show) return null;

  const getVariantClass = () => {
    switch (variant) {
      case 'danger': return 'text-danger';
      case 'warning': return 'text-warning';
      default: return 'text-accent';
    }
  };

  const getBtnClass = () => {
    switch (variant) {
      case 'danger': return 'bg-danger text-white';
      case 'warning': return 'bg-warning text-dark';
      default: return 'primary';
    }
  };

  return (
    <div className="modal-overlay">
      <div className="modal card" style={{ maxWidth: '400px' }}>
        <div className="modal-header">
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
            <AlertTriangle className={getVariantClass()} size={20} />
            <h2 style={{ fontSize: '1.25rem' }}>{title}</h2>
          </div>
          <button onClick={onCancel} className="btn-icon">
            <X size={20} />
          </button>
        </div>
        <div className="modal-body" style={{ padding: '1.5rem' }}>
          <p style={{ color: 'var(--text-secondary)', lineHeight: '1.5' }}>{message}</p>
        </div>
        <div className="modal-footer" style={{ borderTop: '1px solid var(--border)', padding: '1rem 1.5rem', display: 'flex', gap: '0.75rem', justifyContent: 'flex-end' }}>
          <button onClick={onCancel} className="secondary">Cancel</button>
          <button onClick={onConfirm} className={getBtnClass()}>{confirmText}</button>
        </div>
      </div>
    </div>
  );
};

export default ConfirmModal;
