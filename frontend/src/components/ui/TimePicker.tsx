import React from 'react';
import { Clock } from 'lucide-react';
import { clsx } from 'clsx';
import { Select } from './Select';

export interface TimePickerProps {
  value: string; // HH:mm
  onChange: (value: string) => void;
  intervalMinutes?: number; // e.g., 15 or 30
  error?: string;
  disabled?: boolean;
  className?: string;
}

export const TimePicker: React.FC<TimePickerProps> = ({
  value,
  onChange,
  intervalMinutes = 30,
  error,
  disabled,
  className,
}) => {
  // Generate time slots 00:00 -> 23:30 based on interval
  const options = React.useMemo(() => {
    const times: { value: string; label: string }[] = [];
    for (let h = 0; h < 24; h++) {
      for (let m = 0; m < 60; m += intervalMinutes) {
        const hh = String(h).padStart(2, '0');
        const mm = String(m).padStart(2, '0');
        const timeStr = `${hh}:${mm}`;
        times.push({
          value: timeStr,
          label: timeStr,
        });
      }
    }
    return times;
  }, [intervalMinutes]);

  return (
    <Select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      options={options}
      error={error}
      disabled={disabled}
      className={clsx('cursor-pointer', className)}
      leftIcon={<Clock className="w-4 h-4 text-text-muted" />}
    />
  );
};
