import React, { createContext, useContext, useState, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { CheckCircle2, AlertTriangle, AlertCircle, Info, X } from 'lucide-react';
import { clsx } from 'clsx';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

export interface ToastMessage {
  id: string;
  type: ToastType;
  title: string;
  description?: string;
}

interface ToastContextType {
  toast: (title: string, options?: { type?: ToastType; description?: string }) => void;
  success: (title: string, description?: string) => void;
  error: (title: string, description?: string) => void;
  warning: (title: string, description?: string) => void;
  info: (title: string, description?: string) => void;
}

const ToastContext = createContext<ToastContextType | null>(null);

export const ToastProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [toasts, setToasts] = useState<ToastMessage[]>([]);

  const removeToast = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const addToast = useCallback(
    (title: string, options?: { type?: ToastType; description?: string }) => {
      const id = Math.random().toString(36).substring(2, 9);
      const type = options?.type || 'info';
      const description = options?.description;

      setToasts((prev) => [...prev, { id, type, title, description }]);

      setTimeout(() => {
        removeToast(id);
      }, 4000);
    },
    [removeToast]
  );

  const success = useCallback((title: string, description?: string) => addToast(title, { type: 'success', description }), [addToast]);
  const error = useCallback((title: string, description?: string) => addToast(title, { type: 'error', description }), [addToast]);
  const warning = useCallback((title: string, description?: string) => addToast(title, { type: 'warning', description }), [addToast]);
  const info = useCallback((title: string, description?: string) => addToast(title, { type: 'info', description }), [addToast]);

  const icons = {
    success: <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0" />,
    error: <AlertCircle className="w-5 h-5 text-rose-400 shrink-0" />,
    warning: <AlertTriangle className="w-5 h-5 text-amber-400 shrink-0" />,
    info: <Info className="w-5 h-5 text-sky-400 shrink-0" />,
  };

  const borders = {
    success: 'border-emerald-500/30 bg-emerald-950/40',
    error: 'border-rose-500/30 bg-rose-950/40',
    warning: 'border-amber-500/30 bg-amber-950/40',
    info: 'border-sky-500/30 bg-sky-950/40',
  };

  return (
    <ToastContext.Provider value={{ toast: addToast, success, error, warning, info }}>
      {children}
      {typeof document !== 'undefined' &&
        createPortal(
          <div className="fixed bottom-5 right-5 z-50 flex flex-col gap-2 max-w-sm w-full pointer-events-none">
            {toasts.map((t) => (
              <div
                key={t.id}
                className={clsx(
                  'pointer-events-auto flex items-start gap-3 p-4 rounded-xl border backdrop-blur-md shadow-2xl animate-fade-in transition-all',
                  borders[t.type]
                )}
              >
                {icons[t.type]}
                <div className="flex-1 min-w-0">
                  <h5 className="text-xs font-semibold text-text-primary">{t.title}</h5>
                  {t.description && <p className="text-[11px] text-text-muted mt-0.5">{t.description}</p>}
                </div>
                <button
                  onClick={() => removeToast(t.id)}
                  className="text-text-muted hover:text-text-primary p-0.5 rounded"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            ))}
          </div>,
          document.body
        )}
    </ToastContext.Provider>
  );
};

export const useToast = () => {
  const context = useContext(ToastContext);
  if (!context) {
    // Return fallback noop if outside provider
    return {
      toast: () => {},
      success: () => {},
      error: () => {},
      warning: () => {},
      info: () => {},
    };
  }
  return context;
};
