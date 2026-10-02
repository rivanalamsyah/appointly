import React from 'react';
import { Calendar as CalendarIcon } from 'lucide-react';
import { clsx } from 'clsx';
import { Input } from './Input';

export interface DatePickerProps {
  value: string; // YYYY-MM-DD
  onChange: (value: string) => void;
  minDate?: string;
  maxDate?: string;
  label?: string;
  error?: string;
  disabled?: boolean;
  className?: string;
}

export const DatePicker: React.FC<DatePickerProps> = ({
  value,
  onChange,
  minDate,
  maxDate,
  error,
  disabled,
  className,
}) => {
  return (
    <Input
      type="date"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      min={minDate}
      max={maxDate}
      error={error}
      disabled={disabled}
      className={clsx('cursor-pointer', className)}
      leftIcon={<CalendarIcon className="w-4 h-4 text-text-muted" />}
    />
  );
};
