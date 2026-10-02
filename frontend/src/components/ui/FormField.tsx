import React from 'react';
import { clsx } from 'clsx';
import { AlertCircle } from 'lucide-react';

export interface FormFieldProps {
  label?: string;
  htmlFor?: string;
  required?: boolean;
  error?: string;
  description?: string;
  children: React.ReactNode;
  className?: string;
}

export const FormField: React.FC<FormFieldProps> = ({
  label,
  htmlFor,
  required = false,
  error,
  description,
  children,
  className,
}) => {
  return (
    <div className={clsx('flex flex-col gap-1.5 w-full', className)}>
      {label && (
        <label
          htmlFor={htmlFor}
          className="text-xs font-medium text-text-secondary flex items-center gap-1 select-none"
        >
          {label}
          {required && <span className="text-danger font-semibold">*</span>}
        </label>
      )}
      {children}
      {description && !error && (
        <p className="text-xs text-text-muted leading-normal">{description}</p>
      )}
      {error && (
        <div className="flex items-center gap-1 text-xs text-red-400 mt-0.5 animate-fade-in">
          <AlertCircle className="w-3.5 h-3.5 shrink-0" />
          <span>{error}</span>
        </div>
      )}
    </div>
  );
};
