'use client';

import React, { useState } from 'react';
import { 
  Calendar as CalendarIcon, 
  ChevronLeft, 
  ChevronRight, 
  Plus, 
  Filter, 
  User, 
  MapPin, 
  Clock, 
  CheckCircle2, 
  XCircle, 
  AlertCircle
} from 'lucide-react';
import type { Appointment, Staff, Location } from '@/types/api';

interface CalendarViewProps {
  initialAppointments?: Appointment[];
  staffList?: Staff[];
  locations?: Location[];
}

export default function CalendarView({ 
  initialAppointments = [], 
  staffList = [], 
  locations = [] 
}: CalendarViewProps) {
  const [viewMode, setViewMode] = useState<'day' | 'week' | 'month'>('week');
  const [selectedDate, setSelectedDate] = useState<Date>(new Date());
  const [selectedStaff, setSelectedStaff] = useState<string>('all');
  const [selectedLocation, setSelectedLocation] = useState<string>('all');

  // Dummy calendar dates calculation for week view
  const getDaysOfWeek = (date: Date) => {
    const start = new Date(date);
    const day = start.getDay();
    const diff = start.getDate() - day + (day === 0 ? -6 : 1); // Monday start
    const monday = new Date(start.setDate(diff));

    const days = [];
    for (let i = 0; i < 7; i++) {
      const nextDay = new Date(monday);
      nextDay.setDate(monday.getDate() + i);
      days.push(nextDay);
    }
    return days;
  };

  const weekDays = getDaysOfWeek(selectedDate);
  const timeSlots = Array.from({ length: 12 }, (_, i) => `${i + 8}:00`); // 08:00 - 19:00

  const navigateDate = (direction: 'prev' | 'next') => {
    const newDate = new Date(selectedDate);
    if (viewMode === 'day') {
      newDate.setDate(selectedDate.getDate() + (direction === 'next' ? 1 : -1));
    } else if (viewMode === 'week') {
      newDate.setDate(selectedDate.getDate() + (direction === 'next' ? 7 : -7));
    } else {
      newDate.setMonth(selectedDate.getMonth() + (direction === 'next' ? 1 : -1));
    }
    setSelectedDate(newDate);
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'CONFIRMED':
        return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20';
      case 'PENDING_DEPOSIT':
        return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
      case 'COMPLETED':
        return 'bg-blue-500/10 text-blue-400 border-blue-500/20';
      case 'CANCELLED':
        return 'bg-rose-500/10 text-rose-400 border-rose-500/20';
      default:
        return 'bg-slate-500/10 text-slate-400 border-slate-500/20';
    }
  };

  return (
    <div className="space-y-6">
      {/* Calendar Header Controls */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-slate-900/60 backdrop-blur-md border border-slate-800 rounded-2xl p-4 shadow-xl">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-1 bg-slate-800/80 rounded-xl p-1 border border-slate-700/50">
            <button
              onClick={() => navigateDate('prev')}
              className="p-2 text-slate-400 hover:text-slate-100 hover:bg-slate-700/50 rounded-lg transition-colors"
            >
              <ChevronLeft className="w-5 h-5" />
            </button>
            <button
              onClick={() => setSelectedDate(new Date())}
              className="px-3 py-1.5 text-xs font-semibold text-slate-300 hover:text-white transition-colors"
            >
              Today
            </button>
            <button
              onClick={() => navigateDate('next')}
              className="p-2 text-slate-400 hover:text-slate-100 hover:bg-slate-700/50 rounded-lg transition-colors"
            >
              <ChevronRight className="w-5 h-5" />
            </button>
          </div>

          <h2 className="text-xl font-bold text-slate-100 tracking-tight">
            {selectedDate.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
          </h2>
        </div>

        {/* View mode switcher & filters */}
        <div className="flex items-center gap-3 flex-wrap">
          {/* Staff filter */}
          <div className="flex items-center gap-2 bg-slate-800/50 border border-slate-700/50 rounded-xl px-3 py-1.5 text-xs text-slate-300">
            <User className="w-4 h-4 text-violet-400" />
            <select
              value={selectedStaff}
              onChange={(e) => setSelectedStaff(e.target.value)}
              className="bg-transparent text-slate-200 outline-none cursor-pointer"
            >
              <option value="all" className="bg-slate-900">All Staff</option>
              {staffList.map((s) => (
                <option key={s.id} value={s.id} className="bg-slate-900">{s.name}</option>
              ))}
            </select>
          </div>

          {/* View mode selector */}
          <div className="flex bg-slate-800/80 p-1 rounded-xl border border-slate-700/50 text-xs font-medium">
            {(['day', 'week', 'month'] as const).map((mode) => (
              <button
                key={mode}
                onClick={() => setViewMode(mode)}
                className={`px-3 py-1.5 rounded-lg capitalize transition-all ${
                  viewMode === mode
                    ? 'bg-gradient-to-r from-violet-600 to-indigo-600 text-white font-semibold shadow-md'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
              >
                {mode}
              </button>
            ))}
          </div>

          {/* New Appointment button */}
          <button className="flex items-center gap-2 bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white text-xs font-bold px-4 py-2 rounded-xl transition-all shadow-lg shadow-violet-900/30 hover:shadow-violet-900/50 active:scale-95">
            <Plus className="w-4 h-4" />
            Book Appointment
          </button>
        </div>
      </div>

      {/* Calendar Grid View */}
      <div className="bg-slate-900/40 border border-slate-800/80 rounded-2xl overflow-hidden shadow-2xl backdrop-blur-md">
        {/* Days Header */}
        <div className="grid grid-cols-8 border-b border-slate-800 text-center bg-slate-950/60 font-medium text-xs text-slate-400">
          <div className="p-3 border-r border-slate-800 flex items-center justify-center">
            <Clock className="w-4 h-4 text-slate-500" />
          </div>
          {weekDays.map((day, idx) => {
            const isToday = new Date().toDateString() === day.toDateString();
            return (
              <div
                key={idx}
                className={`p-3 border-r border-slate-800/60 transition-colors ${
                  isToday ? 'bg-violet-950/30 text-violet-300 font-bold' : ''
                }`}
              >
                <div className="text-[10px] uppercase tracking-wider text-slate-500">
                  {day.toLocaleDateString('en-US', { weekday: 'short' })}
                </div>
                <div className={`text-base font-semibold mt-0.5 inline-flex items-center justify-center w-7 h-7 rounded-full ${
                  isToday ? 'bg-violet-600 text-white shadow-md shadow-violet-900/50' : 'text-slate-200'
                }`}>
                  {day.getDate()}
                </div>
              </div>
            );
          })}
        </div>

        {/* Time Grid Rows */}
        <div className="divide-y divide-slate-800/40 max-h-[650px] overflow-y-auto custom-scrollbar">
          {timeSlots.map((time, timeIdx) => (
            <div key={timeIdx} className="grid grid-cols-8 min-h-[72px]">
              {/* Time Label Column */}
              <div className="p-2 border-r border-slate-800/80 text-xs font-mono text-slate-500 text-right pr-3 pt-2 bg-slate-950/20 select-none">
                {time}
              </div>

              {/* Day Slot Columns */}
              {weekDays.map((day, dayIdx) => (
                <div
                  key={dayIdx}
                  className="border-r border-slate-800/40 p-1 relative group hover:bg-slate-800/20 transition-colors cursor-pointer"
                >
                  {/* Sample interactive slot hover cue */}
                  <div className="opacity-0 group-hover:opacity-100 transition-opacity absolute inset-1 border border-dashed border-violet-500/40 rounded-lg flex items-center justify-center bg-violet-500/5">
                    <Plus className="w-4 h-4 text-violet-400" />
                  </div>
                </div>
              ))}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
