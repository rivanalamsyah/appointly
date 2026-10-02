'use client';

import React, { useState } from 'react';
import { 
  Clock, 
  Calendar as CalendarIcon, 
  Save, 
  Plus, 
  CheckCircle2, 
  AlertCircle, 
  Globe, 
  User, 
  Palmtree, 
  Sliders, 
  Trash2 
} from 'lucide-react';
import type { BusinessHours, StaffSchedule, StaffTimeOff } from '@/types/api';

interface ScheduleManagerProps {
  timezone?: string;
}

const WEEKDAYS = [
  { key: 'monday', label: 'Monday' },
  { key: 'tuesday', label: 'Tuesday' },
  { key: 'wednesday', label: 'Wednesday' },
  { key: 'thursday', label: 'Thursday' },
  { key: 'friday', label: 'Friday' },
  { key: 'saturday', label: 'Saturday' },
  { key: 'sunday', label: 'Sunday' },
];

export default function ScheduleManager({ timezone = 'America/New_York' }: ScheduleManagerProps) {
  const [activeTab, setActiveTab] = useState<'business' | 'staff' | 'timeoff'>('business');

  // Business Hours State
  const [businessHours, setBusinessHours] = useState<Record<string, { isOpen: boolean; openTime: string; closeTime: string }>>({
    monday: { isOpen: true, openTime: '09:00', closeTime: '18:00' },
    tuesday: { isOpen: true, openTime: '09:00', closeTime: '18:00' },
    wednesday: { isOpen: true, openTime: '09:00', closeTime: '18:00' },
    thursday: { isOpen: true, openTime: '09:00', closeTime: '18:00' },
    friday: { isOpen: true, openTime: '09:00', closeTime: '18:00' },
    saturday: { isOpen: true, openTime: '10:00', closeTime: '16:00' },
    sunday: { isOpen: false, openTime: '09:00', closeTime: '17:00' },
  });

  // Staff Schedule State
  const [staffSchedules, setStaffSchedules] = useState<Record<string, { isWorking: boolean; startTime: string; endTime: string; breakStart: string; breakEnd: string }>>({
    monday: { isWorking: true, startTime: '09:00', endTime: '17:00', breakStart: '12:00', breakEnd: '13:00' },
    tuesday: { isWorking: true, startTime: '09:00', endTime: '17:00', breakStart: '12:00', breakEnd: '13:00' },
    wednesday: { isWorking: true, startTime: '09:00', endTime: '17:00', breakStart: '12:00', breakEnd: '13:00' },
    thursday: { isWorking: true, startTime: '09:00', endTime: '17:00', breakStart: '12:00', breakEnd: '13:00' },
    friday: { isWorking: true, startTime: '09:00', endTime: '17:00', breakStart: '12:00', breakEnd: '13:00' },
    saturday: { isWorking: false, startTime: '09:00', endTime: '17:00', breakStart: '12:00', breakEnd: '13:00' },
    sunday: { isWorking: false, startTime: '09:00', endTime: '17:00', breakStart: '12:00', breakEnd: '13:00' },
  });

  // Time Off Exceptions State
  const [timeOffs, setTimeOffs] = useState<{ id: string; startDate: string; endDate: string; reason: string; isAllDay: boolean }[]>([
    { id: 'to-1', startDate: '2026-11-26', endDate: '2026-11-27', reason: 'Thanksgiving Holiday', isAllDay: true },
  ]);

  // New Time Off Form State
  const [newStartDate, setNewStartDate] = useState('');
  const [newEndDate, setNewEndDate] = useState('');
  const [newReason, setNewReason] = useState('');
  const [newIsAllDay, setNewIsAllDay] = useState(true);

  const [toast, setToast] = useState<string | null>(null);
  const [validationError, setValidationError] = useState<string | null>(null);

  const handleSaveBusinessHours = (e: React.FormEvent) => {
    e.preventDefault();
    setValidationError(null);

    // Validate interval start < end
    for (const day of WEEKDAYS) {
      const bh = businessHours[day.key];
      if (bh?.isOpen && bh.openTime >= bh.closeTime) {
        setValidationError(`Invalid business hours on ${day.label}: Open time (${bh.openTime}) must be before Close time (${bh.closeTime}).`);
        return;
      }
    }

    setToast('Weekly business hours updated successfully!');
    setTimeout(() => setToast(null), 3000);
  };

  const handleSaveStaffSchedule = (e: React.FormEvent) => {
    e.preventDefault();
    setValidationError(null);

    for (const day of WEEKDAYS) {
      const ss = staffSchedules[day.key];
      if (ss?.isWorking) {
        if (ss.startTime >= ss.endTime) {
          setValidationError(`Invalid working hours on ${day.label}: Start time (${ss.startTime}) must be before End time (${ss.endTime}).`);
          return;
        }
        if (ss.breakStart && ss.breakEnd) {
          if (ss.breakStart < ss.startTime || ss.breakEnd > ss.endTime || ss.breakStart >= ss.breakEnd) {
            setValidationError(`Invalid break interval on ${day.label}: Break (${ss.breakStart}-${ss.breakEnd}) must be inside working hours (${ss.startTime}-${ss.endTime}).`);
            return;
          }
        }
      }
    }

    setToast('Staff recurring schedule and break intervals saved!');
    setTimeout(() => setToast(null), 3000);
  };

  const handleAddTimeOff = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newStartDate || !newEndDate) return;
    if (newEndDate < newStartDate) {
      setValidationError('End date cannot be before start date.');
      return;
    }

    const item = {
      id: `to-${Date.now()}`,
      startDate: newStartDate,
      endDate: newEndDate,
      reason: newReason || 'Vacation / Leave',
      isAllDay: newIsAllDay,
    };

    setTimeOffs((prev) => [...prev, item]);
    setNewStartDate('');
    setNewEndDate('');
    setNewReason('');
    setValidationError(null);
    setToast('Staff time-off exception created');
    setTimeout(() => setToast(null), 3000);
  };

  const handleDeleteTimeOff = (id: string) => {
    setTimeOffs((prev) => prev.filter((t) => t.id !== id));
    setToast('Time-off exception removed');
    setTimeout(() => setToast(null), 3000);
  };

  return (
    <div className="space-y-6 max-w-5xl">
      {toast && (
        <div className="p-4 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-semibold flex items-center justify-between shadow-lg animate-fade-in">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0" />
            <span>{toast}</span>
          </div>
          <button onClick={() => setToast(null)} className="text-emerald-500 hover:text-emerald-300">Dismiss</button>
        </div>
      )}

      {validationError && (
        <div className="p-4 rounded-2xl bg-rose-500/10 border border-rose-500/30 text-rose-400 text-xs font-semibold flex items-center justify-between shadow-lg animate-fade-in">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-5 h-5 text-rose-400 shrink-0" />
            <span>{validationError}</span>
          </div>
          <button onClick={() => setValidationError(null)} className="text-rose-400 hover:text-rose-200">Dismiss</button>
        </div>
      )}

      {/* Header Controls & Navigation Tabs */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-slate-900/80 border border-slate-800 rounded-3xl p-6 shadow-xl backdrop-blur-xl">
        <div>
          <h3 className="text-lg font-bold text-white tracking-tight flex items-center gap-2">
            <Clock className="w-5 h-5 text-violet-400" />
            Scheduling & Operating Hours Foundation
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">Manage weekly business hours, staff recurring schedules, and time-off exceptions.</p>
        </div>

        <div className="flex items-center gap-2 bg-slate-950/80 border border-slate-800 p-1.5 rounded-2xl text-xs font-medium">
          <button
            onClick={() => setActiveTab('business')}
            className={`px-4 py-2 rounded-xl transition-all ${
              activeTab === 'business'
                ? 'bg-gradient-to-r from-violet-600 to-indigo-600 text-white font-bold shadow-md'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Location Hours
          </button>
          <button
            onClick={() => setActiveTab('staff')}
            className={`px-4 py-2 rounded-xl transition-all ${
              activeTab === 'staff'
                ? 'bg-gradient-to-r from-violet-600 to-indigo-600 text-white font-bold shadow-md'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Staff Schedule
          </button>
          <button
            onClick={() => setActiveTab('timeoff')}
            className={`px-4 py-2 rounded-xl transition-all ${
              activeTab === 'timeoff'
                ? 'bg-gradient-to-r from-violet-600 to-indigo-600 text-white font-bold shadow-md'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Time Off & Leave
          </button>
        </div>
      </div>

      {/* Timezone Indicator */}
      <div className="bg-slate-950/60 border border-slate-800/80 rounded-2xl p-4 flex items-center justify-between text-xs text-slate-400">
        <div className="flex items-center gap-2 font-mono">
          <Globe className="w-4 h-4 text-violet-400" />
          <span>Active Operating Timezone: <strong className="text-white">{timezone}</strong></span>
        </div>
        <span className="text-[11px] text-slate-500">Daylight Saving Time (DST) Aware</span>
      </div>

      {/* Tab 1: Weekly Business Hours Editor */}
      {activeTab === 'business' && (
        <form onSubmit={handleSaveBusinessHours} className="bg-slate-900/80 border border-slate-800 rounded-3xl p-6 md:p-8 space-y-6 shadow-xl backdrop-blur-xl">
          <div className="pb-4 border-b border-slate-800">
            <h4 className="text-base font-bold text-white">Location Operating Hours</h4>
            <p className="text-xs text-slate-400 mt-0.5">Set regular opening and closing times for your primary business branch.</p>
          </div>

          <div className="space-y-4">
            {WEEKDAYS.map((day) => {
              const bh = businessHours[day.key] || { isOpen: false, openTime: '09:00', closeTime: '18:00' };
              return (
                <div key={day.key} className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 bg-slate-950/60 border border-slate-800 rounded-2xl">
                  <div className="flex items-center gap-3 w-40">
                    <input
                      type="checkbox"
                      checked={bh.isOpen}
                      onChange={(e) =>
                        setBusinessHours((prev) => ({
                          ...prev,
                          [day.key]: { ...bh, isOpen: e.target.checked },
                        }))
                      }
                      className="w-5 h-5 accent-violet-600 rounded cursor-pointer"
                    />
                    <span className={`text-xs font-semibold ${bh.isOpen ? 'text-white' : 'text-slate-500'}`}>
                      {day.label}
                    </span>
                  </div>

                  {bh.isOpen ? (
                    <div className="flex items-center gap-3 font-mono text-xs">
                      <input
                        type="time"
                        value={bh.openTime}
                        onChange={(e) =>
                          setBusinessHours((prev) => ({
                            ...prev,
                            [day.key]: { ...bh, openTime: e.target.value },
                          }))
                        }
                        className="bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-slate-200 focus:outline-none focus:border-violet-500"
                      />
                      <span className="text-slate-500">to</span>
                      <input
                        type="time"
                        value={bh.closeTime}
                        onChange={(e) =>
                          setBusinessHours((prev) => ({
                            ...prev,
                            [day.key]: { ...bh, closeTime: e.target.value },
                          }))
                        }
                        className="bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-slate-200 focus:outline-none focus:border-violet-500"
                      />
                    </div>
                  ) : (
                    <span className="text-xs font-semibold text-rose-400/80 bg-rose-500/10 border border-rose-500/20 px-3 py-1 rounded-full">
                      Closed
                    </span>
                  )}
                </div>
              );
            })}
          </div>

          <div className="flex justify-end pt-4 border-t border-slate-800">
            <button
              type="submit"
              className="bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white font-bold text-xs px-6 py-3 rounded-xl transition-all shadow-lg flex items-center gap-2"
            >
              <Save className="w-4 h-4" />
              Save Location Business Hours
            </button>
          </div>
        </form>
      )}

      {/* Tab 2: Staff Schedule & Break Editor */}
      {activeTab === 'staff' && (
        <form onSubmit={handleSaveStaffSchedule} className="bg-slate-900/80 border border-slate-800 rounded-3xl p-6 md:p-8 space-y-6 shadow-xl backdrop-blur-xl">
          <div className="pb-4 border-b border-slate-800 flex justify-between items-center">
            <div>
              <h4 className="text-base font-bold text-white">Staff Member Recurring Schedule</h4>
              <p className="text-xs text-slate-400 mt-0.5">Configure weekly shifts and break intervals for staff providers.</p>
            </div>
          </div>

          <div className="space-y-4">
            {WEEKDAYS.map((day) => {
              const ss = staffSchedules[day.key] || { isWorking: false, startTime: '09:00', endTime: '17:00', breakStart: '12:00', breakEnd: '13:00' };
              return (
                <div key={day.key} className="p-4 bg-slate-950/60 border border-slate-800 rounded-2xl space-y-3">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <input
                        type="checkbox"
                        checked={ss.isWorking}
                        onChange={(e) =>
                          setStaffSchedules((prev) => ({
                            ...prev,
                            [day.key]: { ...ss, isWorking: e.target.checked },
                          }))
                        }
                        className="w-5 h-5 accent-violet-600 rounded cursor-pointer"
                      />
                      <span className={`text-xs font-semibold ${ss.isWorking ? 'text-white' : 'text-slate-500'}`}>
                        {day.label} Shift
                      </span>
                    </div>

                    {!ss.isWorking && (
                      <span className="text-xs font-semibold text-slate-500 bg-slate-800 px-3 py-1 rounded-full">
                        Day Off
                      </span>
                    )}
                  </div>

                  {ss.isWorking && (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-2 border-t border-slate-800/60 text-xs font-mono">
                      <div>
                        <span className="text-[10px] text-slate-500 block mb-1">Working Hours</span>
                        <div className="flex items-center gap-2">
                          <input
                            type="time"
                            value={ss.startTime}
                            onChange={(e) =>
                              setStaffSchedules((prev) => ({
                                ...prev,
                                [day.key]: { ...ss, startTime: e.target.value },
                              }))
                            }
                            className="bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-slate-200 focus:outline-none focus:border-violet-500"
                          />
                          <span className="text-slate-500">to</span>
                          <input
                            type="time"
                            value={ss.endTime}
                            onChange={(e) =>
                              setStaffSchedules((prev) => ({
                                ...prev,
                                [day.key]: { ...ss, endTime: e.target.value },
                              }))
                            }
                            className="bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-slate-200 focus:outline-none focus:border-violet-500"
                          />
                        </div>
                      </div>

                      <div>
                        <span className="text-[10px] text-slate-500 block mb-1">Meal / Break Interval</span>
                        <div className="flex items-center gap-2">
                          <input
                            type="time"
                            value={ss.breakStart}
                            onChange={(e) =>
                              setStaffSchedules((prev) => ({
                                ...prev,
                                [day.key]: { ...ss, breakStart: e.target.value },
                              }))
                            }
                            className="bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-slate-200 focus:outline-none focus:border-violet-500"
                          />
                          <span className="text-slate-500">to</span>
                          <input
                            type="time"
                            value={ss.breakEnd}
                            onChange={(e) =>
                              setStaffSchedules((prev) => ({
                                ...prev,
                                [day.key]: { ...ss, breakEnd: e.target.value },
                              }))
                            }
                            className="bg-slate-900 border border-slate-800 rounded-xl px-3 py-2 text-slate-200 focus:outline-none focus:border-violet-500"
                          />
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              );
            })}
          </div>

          <div className="flex justify-end pt-4 border-t border-slate-800">
            <button
              type="submit"
              className="bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white font-bold text-xs px-6 py-3 rounded-xl transition-all shadow-lg flex items-center gap-2"
            >
              <Save className="w-4 h-4" />
              Save Staff Recurring Shift & Breaks
            </button>
          </div>
        </form>
      )}

      {/* Tab 3: Time Off & Vacation Exceptions Manager */}
      {activeTab === 'timeoff' && (
        <div className="space-y-6">
          {/* Create Time Off Form */}
          <form onSubmit={handleAddTimeOff} className="bg-slate-900/80 border border-slate-800 rounded-3xl p-6 md:p-8 space-y-4 shadow-xl backdrop-blur-xl">
            <h4 className="text-base font-bold text-white flex items-center gap-2">
              <Palmtree className="w-5 h-5 text-emerald-400" />
              Add Time Off / Vacation Exception
            </h4>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Start Date *</label>
                <input
                  type="date"
                  required
                  value={newStartDate}
                  onChange={(e) => setNewStartDate(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-2.5 text-slate-200 text-xs font-mono focus:outline-none focus:border-violet-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">End Date *</label>
                <input
                  type="date"
                  required
                  value={newEndDate}
                  onChange={(e) => setNewEndDate(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-2.5 text-slate-200 text-xs font-mono focus:outline-none focus:border-violet-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Reason / Note</label>
                <input
                  type="text"
                  placeholder="e.g. Annual Leave, Holiday"
                  value={newReason}
                  onChange={(e) => setNewReason(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-2.5 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                />
              </div>
            </div>

            <div className="flex justify-between items-center pt-2">
              <label className="flex items-center gap-2 text-xs text-slate-300 cursor-pointer">
                <input
                  type="checkbox"
                  checked={newIsAllDay}
                  onChange={(e) => setNewIsAllDay(e.target.checked)}
                  className="w-4 h-4 accent-violet-600 rounded cursor-pointer"
                />
                All Day Exception
              </label>

              <button
                type="submit"
                className="bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold px-5 py-2.5 rounded-xl shadow-lg flex items-center gap-2"
              >
                <Plus className="w-4 h-4" />
                Add Time Off Exception
              </button>
            </div>
          </form>

          {/* Active Exceptions List */}
          <div className="bg-slate-900/80 border border-slate-800 rounded-3xl p-6 space-y-4 shadow-xl backdrop-blur-xl">
            <h4 className="text-sm font-bold text-white">Scheduled Time-Off Exceptions</h4>

            <div className="space-y-2">
              {timeOffs.map((tOff) => (
                <div key={tOff.id} className="flex items-center justify-between p-4 bg-slate-950/60 border border-slate-800 rounded-2xl">
                  <div className="space-y-0.5">
                    <div className="text-xs font-bold text-white flex items-center gap-2">
                      <span>{tOff.reason}</span>
                      <span className="text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 px-2 py-0.5 rounded-full font-mono">
                        Approved
                      </span>
                    </div>
                    <div className="text-xs font-mono text-slate-400">
                      {tOff.startDate} to {tOff.endDate} ({tOff.isAllDay ? 'All Day' : 'Partial Hours'})
                    </div>
                  </div>

                  <button
                    onClick={() => handleDeleteTimeOff(tOff.id)}
                    className="p-2 text-rose-400 hover:bg-rose-500/10 rounded-xl transition-colors"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
