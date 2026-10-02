import React, { forwardRef } from 'react';
import { ChevronDown } from 'lucide-react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export interface SelectOption {
  value: string;
  label: string;
  disabled?: boolean;
}

export interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  error?: string;
  options?: SelectOption[];
  fullWidth?: boolean;
  leftIcon?: React.ReactNode;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  (
    {
      className,
      error,
      options = [],
      fullWidth = true,
      leftIcon,
      disabled,
      children,
      ...props
    },
    ref
  ) => {
    return (
      <div className={clsx('relative inline-flex items-center', fullWidth && 'w-full')}>
        {leftIcon && (
          <div className="absolute left-3 top-1/2 -translate-y-1/2 text-text-muted pointer-events-none flex items-center justify-center">
            {leftIcon}
          </div>
        )}
        <select
          ref={ref}
          disabled={disabled}
          className={twMerge(
            clsx(
              'w-full appearance-none rounded-lg bg-surface-overlay border border-surface-border text-text-primary text-sm transition-all duration-200 focus:outline-none focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 disabled:opacity-50 disabled:cursor-not-allowed pr-9 py-2 min-h-[40px]',
              leftIcon ? 'pl-10' : 'pl-3.5',
              error && 'border-danger focus:border-danger focus:ring-danger/20',
              className
            )
          )}
          {...props}
        >
          {children ||
            options.map((opt) => (
              <option key={opt.value} value={opt.value} disabled={opt.disabled} className="bg-surface-raised text-text-primary">
                {opt.label}
              </option>
            ))}
        </select>
        <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted pointer-events-none" />
      </div>
    );
  }
);

Select.displayName = 'Select';
