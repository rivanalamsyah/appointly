import React from 'react';

export interface SeparatorProps {
  orientation?: 'horizontal' | 'vertical';
  className?: string;
}

export const Separator: React.FC<SeparatorProps> = ({
  orientation = 'horizontal',
  className = '',
}) => {
  return (
    <div
      role="separator"
      aria-orientation={orientation}
      className={`${
        orientation === 'horizontal'
          ? 'h-[1px] w-full bg-[var(--color-surface-border)]'
          : 'w-[1px] h-full bg-[var(--color-surface-border)]'
      } ${className}`}
    />
  );
};
