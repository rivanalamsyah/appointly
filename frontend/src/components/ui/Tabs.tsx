import React from 'react';
import { clsx } from 'clsx';

export interface TabItem {
  key: string;
  label: React.ReactNode;
  icon?: React.ReactNode;
  badge?: string | number;
  disabled?: boolean;
}

export interface TabsProps {
  tabs: TabItem[];
  activeKey: string;
  onChange: (key: string) => void;
  variant?: 'underline' | 'pills';
  className?: string;
}

export const Tabs: React.FC<TabsProps> = ({
  tabs,
  activeKey,
  onChange,
  variant = 'underline',
  className,
}) => {
  return (
    <div
      role="tablist"
      className={clsx(
        'flex items-center gap-1 overflow-x-auto no-scrollbar',
        variant === 'underline' ? 'border-b border-surface-border pb-px' : 'p-1 bg-surface-overlay rounded-lg border border-surface-border',
        className
      )}
    >
      {tabs.map((tab) => {
        const isActive = tab.key === activeKey;
        return (
          <button
            key={tab.key}
            role="tab"
            aria-selected={isActive}
            disabled={tab.disabled}
            onClick={() => onChange(tab.key)}
            className={clsx(
              'inline-flex items-center gap-2 px-3.5 py-2 text-xs font-medium transition-all duration-200 whitespace-nowrap rounded-md disabled:opacity-40 disabled:cursor-not-allowed focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary-500',
              variant === 'underline'
                ? isActive
                  ? 'text-primary-400 border-b-2 border-primary-500 rounded-b-none font-semibold'
                  : 'text-text-muted hover:text-text-primary'
                : isActive
                ? 'bg-primary-500/20 text-primary-300 shadow-xs border border-primary-500/30 font-semibold'
                : 'text-text-muted hover:text-text-primary hover:bg-white/5'
            )}
          >
            {tab.icon && <span className="w-4 h-4 shrink-0">{tab.icon}</span>}
            <span>{tab.label}</span>
            {tab.badge !== undefined && (
              <span
                className={clsx(
                  'px-1.5 py-0.5 text-[10px] rounded-full font-bold',
                  isActive
                    ? 'bg-primary-500/30 text-primary-200'
                    : 'bg-white/10 text-text-muted'
                )}
              >
                {tab.badge}
              </span>
            )}
          </button>
        );
      })}
    </div>
  );
};
