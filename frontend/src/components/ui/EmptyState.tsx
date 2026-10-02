import React from 'react';
import { FolderOpen } from 'lucide-react';
import { clsx } from 'clsx';
import { Button, type ButtonProps } from './Button';

export interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
    icon?: React.ReactNode;
    variant?: ButtonProps['variant'];
  };
  secondaryAction?: {
    label: string;
    onClick: () => void;
  };
  className?: string;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  icon,
  title,
  description,
  action,
  secondaryAction,
  className,
}) => {
  return (
    <div className={clsx('flex flex-col items-center justify-center text-center p-8 sm:p-12 rounded-xl border border-dashed border-surface-border bg-surface-raised/20', className)}>
      <div className="w-12 h-12 rounded-full bg-primary-500/10 text-primary-400 flex items-center justify-center mb-4 shrink-0 shadow-inner">
        {icon || <FolderOpen className="w-6 h-6" />}
      </div>
      <h4 className="text-base font-semibold text-text-primary mb-1">{title}</h4>
      {description && <p className="text-xs text-text-muted max-w-sm mb-6 leading-relaxed">{description}</p>}
      {(action || secondaryAction) && (
        <div className="flex items-center gap-3">
          {secondaryAction && (
            <Button variant="ghost" size="sm" onClick={secondaryAction.onClick}>
              {secondaryAction.label}
            </Button>
          )}
          {action && (
            <Button
              variant={action.variant || 'primary'}
              size="sm"
              onClick={action.onClick}
              leftIcon={action.icon}
            >
              {action.label}
            </Button>
          )}
        </div>
      )}
    </div>
  );
};
