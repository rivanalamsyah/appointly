import React from 'react';
import { clsx } from 'clsx';
import { twMerge } from 'tailwind-merge';

export interface SkeletonProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'text' | 'circular' | 'rectangular';
  width?: string | number;
  height?: string | number;
}

export const Skeleton: React.FC<SkeletonProps> = ({
  variant = 'text',
  width,
  height,
  className,
  style,
  ...props
}) => {
  const base = 'animate-pulse bg-surface-overlay/80 shrink-0';

  const variants = {
    text: 'h-4 rounded w-full my-1',
    circular: 'rounded-full',
    rectangular: 'rounded-lg w-full h-24',
  };

  return (
    <div
      className={twMerge(clsx(base, variants[variant], className))}
      style={{
        width: width !== undefined ? width : undefined,
        height: height !== undefined ? height : undefined,
        ...style,
      }}
      {...props}
    />
  );
};
