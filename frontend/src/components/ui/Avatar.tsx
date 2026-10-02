import React from 'react';
import { clsx } from 'clsx';

export interface AvatarProps {
  name: string;
  src?: string;
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  status?: 'online' | 'busy' | 'offline';
  className?: string;
}

export const Avatar: React.FC<AvatarProps> = ({
  name,
  src,
  size = 'md',
  status,
  className,
}) => {
  const initials = name
    .split(' ')
    .filter(Boolean)
    .slice(0, 2)
    .map((n: string) => n[0])
    .join('')
    .toUpperCase() || 'U';

  const sizes: Record<'xs' | 'sm' | 'md' | 'lg' | 'xl', string> = {
    xs: 'w-6 h-6 text-[10px]',
    sm: 'w-8 h-8 text-xs',
    md: 'w-10 h-10 text-sm',
    lg: 'w-12 h-12 text-base',
    xl: 'w-16 h-16 text-xl',
  };

  const statusColors: Record<'online' | 'busy' | 'offline', string> = {
    online: 'bg-emerald-500 shadow-emerald-500/50',
    busy: 'bg-amber-500 shadow-amber-500/50',
    offline: 'bg-slate-500 shadow-slate-500/50',
  };

  return (
    <div className={clsx('relative inline-flex shrink-0 select-none', className)}>
      {src ? (
        <img
          src={src}
          alt={name}
          className={clsx(
            'rounded-full object-cover border border-surface-border',
            sizes[size]
          )}
        />
      ) : (
        <div
          className={clsx(
            'rounded-full bg-gradient-to-br from-primary-600 to-indigo-800 text-white font-semibold flex items-center justify-center border border-white/10 shadow-sm',
            sizes[size]
          )}
        >
          {initials}
        </div>
      )}
      {status && (
        <span
          className={clsx(
            'absolute bottom-0 right-0 w-2.5 h-2.5 rounded-full ring-2 ring-surface-base shadow-xs',
            statusColors[status]
          )}
        />
      )}
    </div>
  );
};
