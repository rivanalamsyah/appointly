import React from 'react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?:
    | 'pending'
    | 'confirmed'
    | 'completed'
    | 'cancelled'
    | 'rescheduled'
    | 'no_show'
    | 'active'
    | 'inactive'
    | 'primary'
    | 'outline'
    | 'info';
  size?: 'sm' | 'md';
  dot?: boolean;
}

export const Badge: React.FC<BadgeProps> = ({
  children,
  variant = 'primary',
  size = 'md',
  dot = false,
  className,
  ...props
}) => {
  const base =
    'inline-flex items-center gap-1.5 font-medium rounded-full whitespace-nowrap transition-colors';

  const variants = {
    pending: 'bg-amber-500/15 text-amber-400 border border-amber-500/30',
    confirmed: 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30',
    completed: 'bg-indigo-500/15 text-indigo-400 border border-indigo-500/30',
    cancelled: 'bg-rose-500/15 text-rose-400 border border-rose-500/30',
    rescheduled: 'bg-blue-500/15 text-blue-400 border border-blue-500/30',
    no_show: 'bg-slate-500/15 text-slate-400 border border-slate-500/30',
    active: 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30',
    inactive: 'bg-slate-500/15 text-slate-400 border border-slate-500/30',
    primary: 'bg-primary-500/15 text-primary-300 border border-primary-500/30',
    info: 'bg-sky-500/15 text-sky-400 border border-sky-500/30',
    outline: 'bg-transparent text-text-secondary border border-surface-border',
  };

  const sizes = {
    sm: 'px-2 py-0.5 text-[10px]',
    md: 'px-2.5 py-1 text-xs',
  };

  const dotColors = {
    pending: 'bg-amber-400',
    confirmed: 'bg-emerald-400',
    completed: 'bg-indigo-400',
    cancelled: 'bg-rose-400',
    rescheduled: 'bg-blue-400',
    no_show: 'bg-slate-400',
    active: 'bg-emerald-400',
    inactive: 'bg-slate-400',
    primary: 'bg-primary-400',
    info: 'bg-sky-400',
    outline: 'bg-text-secondary',
  };

  return (
    <span
      className={twMerge(clsx(base, variants[variant], sizes[size], className))}
      {...props}
    >
      {dot && <span className={clsx('w-1.5 h-1.5 rounded-full shrink-0', dotColors[variant])} />}
      {children}
    </span>
  );
};
