import React from 'react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'default' | 'glass' | 'bordered';
  hoverable?: boolean;
}

export const Card: React.FC<CardProps> = ({
  children,
  variant = 'default',
  hoverable = false,
  className,
  ...props
}) => {
  const base = 'rounded-xl overflow-hidden transition-all duration-200';

  const variants = {
    default: 'bg-surface-raised border border-surface-border shadow-md',
    glass: 'bg-surface-raised/70 backdrop-blur-md border border-white/[0.08] shadow-xl',
    bordered: 'bg-transparent border border-surface-border',
  };

  const hover = hoverable ? 'hover:-translate-y-0.5 hover:border-primary-500/40 hover:shadow-lg hover:shadow-primary-500/5' : '';

  return (
    <div className={twMerge(clsx(base, variants[variant], hover, className))} {...props}>
      {children}
    </div>
  );
};

export const CardHeader: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...props
}) => (
  <div className={twMerge(clsx('p-5 border-b border-surface-border/60 flex items-center justify-between', className))} {...props}>
    {children}
  </div>
);

export const CardBody: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...props
}) => (
  <div className={twMerge(clsx('p-5', className))} {...props}>
    {children}
  </div>
);

export const CardFooter: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  children,
  className,
  ...props
}) => (
  <div className={twMerge(clsx('px-5 py-4 border-t border-surface-border/60 bg-surface-overlay/30 flex items-center justify-between', className))} {...props}>
    {children}
  </div>
);
