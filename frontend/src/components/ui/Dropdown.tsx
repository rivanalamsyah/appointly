import React, { useState, useRef, useEffect } from 'react';
import { clsx } from 'clsx';

export interface DropdownItem {
  key: string;
  label: React.ReactNode;
  icon?: React.ReactNode;
  onClick?: () => void;
  danger?: boolean;
  disabled?: boolean;
  divider?: boolean;
}

export interface DropdownProps {
  trigger: React.ReactNode;
  items: DropdownItem[];
  align?: 'left' | 'right';
  className?: string;
}

export const Dropdown: React.FC<DropdownProps> = ({
  trigger,
  items,
  align = 'right',
  className,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  return (
    <div ref={containerRef} className={clsx('relative inline-block text-left', className)}>
      <div onClick={() => setIsOpen(!isOpen)} role="button" tabIndex={0}>
        {trigger}
      </div>

      {isOpen && (
        <div
          className={clsx(
            'absolute z-50 mt-1 min-w-[180px] rounded-lg bg-surface-overlay border border-surface-border shadow-xl backdrop-blur-md p-1 animate-fade-in focus:outline-none',
            align === 'right' ? 'right-0' : 'left-0'
          )}
        >
          {items.map((item: DropdownItem, index: number) => {
            if (item.divider) {
              return <div key={`divider-${index}`} className="my-1 border-t border-surface-border" />;
            }
            return (
              <button
                key={item.key}
                type="button"
                disabled={item.disabled}
                onClick={() => {
                  if (item.onClick) item.onClick();
                  setIsOpen(false);
                }}
                className={clsx(
                  'w-full flex items-center gap-2.5 px-3 py-2 text-xs rounded-md font-medium transition-colors duration-150 text-left disabled:opacity-50 disabled:cursor-not-allowed',
                  item.danger
                    ? 'text-red-400 hover:bg-danger/15'
                    : 'text-text-primary hover:bg-white/5 hover:text-white'
                )}
              >
                {item.icon && <span className="w-4 h-4 shrink-0 opacity-80">{item.icon}</span>}
                <span>{item.label}</span>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
};
