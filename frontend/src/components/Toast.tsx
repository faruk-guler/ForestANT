import React from 'react';
import { useToast } from '../contexts/ToastContext';
import { XCircle, CheckCircle, Info, X } from 'lucide-react';

export const ToastContainer: React.FC = () => {
  const { toasts, removeToast } = useToast();

  if (toasts.length === 0) return null;

  return (
    <div className="toast-container" style={{ position: 'fixed', top: '20px', right: '20px', zIndex: 9999, display: 'flex', flexDirection: 'column', gap: '10px' }}>
      {toasts.map((toast: any) => (
        <div key={toast.id} className={`toast toast-${toast.type}`} style={{ 
          display: 'flex', alignItems: 'center', gap: '12px', padding: '16px 20px',
          background: 'rgba(20, 24, 30, 0.95)', border: '1px solid var(--border)', 
          backdropFilter: 'blur(10px)', borderRadius: '12px', color: 'white',
          boxShadow: '0 10px 30px rgba(0,0,0,0.5)', minWidth: '300px',
          animation: 'fadeInRight 0.3s cubic-bezier(0.16, 1, 0.3, 1)'
        }}>
          {toast.type === 'error' && <XCircle color="var(--danger)" />}
          {toast.type === 'success' && <CheckCircle color="var(--success)" />}
          {toast.type === 'info' && <Info color="var(--accent-color)" />}
          
          <span style={{ flex: 1, fontSize: '0.95rem' }}>{toast.message}</span>
          
          <button onClick={() => removeToast(toast.id)} className="btn-icon" style={{ padding: '4px', margin: 0 }}>
            <X size={16} />
          </button>
        </div>
      ))}
    </div>
  );
};
