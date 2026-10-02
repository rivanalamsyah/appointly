'use client';

import React, { useState, useEffect, useTransition } from 'react';
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
  AlertCircle,
  RefreshCw,
  List as ListIcon,
  Grid as GridIcon,
  Shield,
  Eye,
  DollarSign,
  Phone,
  Mail,
  X,
  RotateCcw
} from 'lucide-react';
import type { Appointment, Staff, Location, Service, AppointmentStatus } from '@/types/api';

interface CalendarViewProps {
  initialAppointments?: Appointment[];
  staffList?: Staff[];
  locations?: Location[];
  services?: Service[];
  timezone?: string;
}

export default function CalendarView({ 
  initialAppointments = [], 
  staffList = [], 
  locations = [],
  services = [],
  timezone = 'America/New_York'
}: CalendarViewProps) {
  // View states
  const [viewMode, setViewMode] = useState<'day' | 'week' | 'month' | 'list'>('week');
  const [selectedDate, setSelectedDate] = useState<Date>(new Date());
  
  // Filter states
  const [selectedStaff, setSelectedStaff] = useState<string>('all');
  const [selectedLocation, setSelectedLocation] = useState<string>('all');
  const [selectedService, setSelectedService] = useState<string>('all');
  const [selectedStatus, setSelectedStatus] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState<string>('');

  // Data & Loading
  const [appointments, setAppointments] = useState<Appointment[]>(initialAppointments);
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [isPending, startTransition] = useTransition();
  const [error, setError] = useState<string | null>(null);
  const [isAutoRefresh, setIsAutoRefresh] = useState<boolean>(true);

  // Selected appointment detail drawer state
  const [activeAppointment, setActiveAppointment] = useState<Appointment | null>(null);
  const [isDetailOpen, setIsDetailOpen] = useState<boolean>(false);

  // Create modal state for empty slot click
  const [isCreateModalOpen, setIsCreateModalOpen] = useState<boolean>(false);
  const [createSlotData, setCreateSlotData] = useState<{ date: string; time: string; staffId?: string } | null>(null);

  // Drag and drop reschedule state
  const [draggedAppt, setDraggedAppt] = useState<Appointment | null>(null);
  const [conflictWarning, setConflictWarning] = useState<string | null>(null);
  const [cancelReason, setCancelReason] = useState<string>('');
  const [showCancelDialog, setShowCancelDialog] = useState<boolean>(false);

  // Sample mock appointments if none provided
  useEffect(() => {
    if (appointments.length === 0) {
      const today = new Date();
      const isoToday = today.toISOString().split('T')[0];
      setAppointments([
        {
          id: 'appt-1',
          tenant_id: 't-1',
          customer_id: 'c-1',
          service_id: 'srv-1',
          staff_id: 'stf-1',
          location_id: 'loc-1',
          start_time: `${isoToday}T09:00:00Z`,
          end_time: `${isoToday}T09:45:00Z`,
          timezone: timezone,
          status: 'CONFIRMED',
          price_cents: 6500,
          currency: 'USD',
          guest_name: 'Sophia Williams',
          guest_email: 'sophia@example.com',
          guest_phone: '+1 (555) 345-6789',
          notes: 'Prefers quiet session with moisturizing toner.',
          source: 'online',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: 'appt-2',
          tenant_id: 't-1',
          customer_id: 'c-2',
          service_id: 'srv-2',
          staff_id: 'stf-2',
          location_id: 'loc-1',
          start_time: `${isoToday}T11:00:00Z`,
          end_time: `${isoToday}T11:30:00Z`,
          timezone: timezone,
          status: 'PENDING_DEPOSIT',
          price_cents: 3500,
          currency: 'USD',
          guest_name: 'Marcus Brody',
          guest_email: 'marcus@example.com',
          guest_phone: '+1 (555) 987-6543',
          notes: 'First time beard edge-up.',
          source: 'online',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        },
        {
          id: 'appt-3',
          tenant_id: 't-1',
          customer_id: 'c-3',
          service_id: 'srv-3',
          staff_id: 'stf-1',
          location_id: 'loc-1',
          start_time: `${isoToday}T14:00:00Z`,
          end_time: `${isoToday}T16:00:00Z`,
          timezone: timezone,
          status: 'COMPLETED',
          price_cents: 18000,
          currency: 'USD',
          guest_name: 'Emma Watson',
          guest_email: 'emma@example.com',
          guest_phone: '+1 (555) 222-3333',
          notes: 'Balayage highlights package.',
          source: 'manual',
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        }
      ]);
    }
  }, []);

  // Fetch appointments from API
  const fetchAppointments = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams();
      if (selectedStaff !== 'all') params.append('staff_id', selectedStaff);
      if (selectedLocation !== 'all') params.append('location_id', selectedLocation);
      if (selectedService !== 'all') params.append('service_id', selectedService);
      if (selectedStatus !== 'all') params.append('status', selectedStatus);

      const res = await fetch(`/api/v1/appointments?${params.toString()}`);
      if (res.ok) {
        const json = await res.json();
        if (json.data && Array.isArray(json.data)) {
          setAppointments(json.data);
        }
      }
    } catch (err: any) {
      // Keep existing state on API error
    } finally {
      setIsLoading(false);
    }
  };

  // Auto-refresh interval (every 30s)
  useEffect(() => {
    if (!isAutoRefresh) return;
    const timer = setInterval(() => {
      fetchAppointments();
    }, 30000);
    return () => clearInterval(timer);
  }, [isAutoRefresh, selectedStaff, selectedLocation, selectedService, selectedStatus]);

  // Date navigation calculations
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
  const timeSlots = Array.from({ length: 12 }, (_, i) => `${(i + 8).toString().padStart(2, '0')}:00`); // 08:00 - 19:00

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

  // Status badge styling helper
  const getStatusBadge = (status: string) => {
    switch (status.toUpperCase()) {
      case 'CONFIRMED':
        return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30';
      case 'PENDING_DEPOSIT':
      case 'PENDING':
        return 'bg-amber-500/10 text-amber-400 border-amber-500/30';
      case 'COMPLETED':
        return 'bg-blue-500/10 text-blue-400 border-blue-500/30';
      case 'RESCHEDULED':
        return 'bg-indigo-500/10 text-indigo-400 border-indigo-500/30';
      case 'CANCELLED':
        return 'bg-rose-500/10 text-rose-400 border-rose-500/30';
      case 'NO_SHOW':
        return 'bg-slate-500/10 text-slate-400 border-slate-500/30';
      default:
        return 'bg-slate-500/10 text-slate-400 border-slate-500/30';
    }
  };

  // Filtered appointments list
  const filteredAppointments = appointments.filter((appt) => {
    if (selectedStaff !== 'all' && appt.staff_id !== selectedStaff) return false;
    if (selectedLocation !== 'all' && appt.location_id !== selectedLocation) return false;
    if (selectedService !== 'all' && appt.service_id !== selectedService) return false;
    if (selectedStatus !== 'all' && appt.status.toUpperCase() !== selectedStatus.toUpperCase()) return false;
    if (searchQuery.trim() !== '') {
      const q = searchQuery.toLowerCase();
      const matchName = appt.guest_name?.toLowerCase().includes(q);
      const matchEmail = appt.guest_email?.toLowerCase().includes(q);
      const matchPhone = appt.guest_phone?.toLowerCase().includes(q);
      if (!matchName && !matchEmail && !matchPhone) return false;
    }
    return true;
  });

  // Handle slot click to open Create Appointment Modal
  const handleEmptySlotClick = (dateStr: string, timeStr: string) => {
    setCreateSlotData({
      date: dateStr,
      time: timeStr,
      staffId: selectedStaff !== 'all' ? selectedStaff : undefined,
    });
    setIsCreateModalOpen(true);
  };

  // Handle status update (Confirm, Complete, No-Show, Cancel)
  const handleUpdateStatus = async (apptId: string, newStatus: string) => {
    const previousState = [...appointments];
    // Optimistic UI update
    setAppointments((prev) =>
      prev.map((a) => (a.id === apptId ? { ...a, status: newStatus as AppointmentStatus } : a))
    );
    if (activeAppointment?.id === apptId) {
      setActiveAppointment((prev) => (prev ? { ...prev, status: newStatus as AppointmentStatus } : null));
    }

    try {
      const res = await fetch(`/api/v1/appointments/${apptId}/status`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ status: newStatus, reason: cancelReason }),
      });
      if (!res.ok) {
        throw new Error('Failed to update status');
      }
    } catch (err) {
      // Rollback on failure
      setAppointments(previousState);
      setError('Could not update status. Transaction rolled back.');
    } finally {
      setShowCancelDialog(false);
      setCancelReason('');
    }
  };

  // Handle Reschedule Drop / Change
  const handleReschedule = async (apptId: string, targetDateStr: string, targetTimeStr: string) => {
    const appt = appointments.find((a) => a.id === apptId);
    if (!appt) return;

    const newStartISO = `${targetDateStr}T${targetTimeStr}:00Z`;

    // Conflict Check
    const hasConflict = appointments.some(
      (a) =>
        a.id !== apptId &&
        a.staff_id === appt.staff_id &&
        a.status !== 'CANCELLED' &&
        a.start_time.startsWith(newStartISO.substring(0, 16))
    );

    if (hasConflict) {
      setConflictWarning(`Conflict detected: Staff member already has a booking at ${targetTimeStr}.`);
      return;
    }

    setConflictWarning(null);
    const previousState = [...appointments];

    // Optimistic UI update
    setAppointments((prev) =>
      prev.map((a) =>
        a.id === apptId
          ? {
              ...a,
              start_time: newStartISO,
              status: 'RESCHEDULED',
            }
          : a
      )
    );

    try {
      const res = await fetch(`/api/v1/appointments/${apptId}/reschedule`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ new_start_time: newStartISO }),
      });
      if (!res.ok) throw new Error('Reschedule failed');
    } catch (err) {
      // Rollback on failure
      setAppointments(previousState);
      setConflictWarning('Reschedule failed on server validation. Transaction rolled back.');
    }
  };

  return (
    <div className="space-y-6">
      {/* Calendar Header Controls */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-slate-900/80 backdrop-blur-md border border-slate-800 rounded-2xl p-4 shadow-xl">
        <div className="flex items-center gap-3">
          {/* Navigation Controls */}
          <div className="flex items-center gap-1 bg-slate-800/80 rounded-xl p-1 border border-slate-700/50">
            <button
              onClick={() => navigateDate('prev')}
              className="p-2 text-slate-400 hover:text-slate-100 hover:bg-slate-700/50 rounded-lg transition-colors"
              title="Previous Period"
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
              title="Next Period"
            >
              <ChevronRight className="w-5 h-5" />
            </button>
          </div>

          <div>
            <h2 className="text-xl font-bold text-slate-100 tracking-tight">
              {selectedDate.toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
            </h2>
            <div className="flex items-center gap-2 text-[11px] text-slate-400">
              <Clock className="w-3 h-3 text-violet-400" />
              <span>Timezone: <strong className="text-slate-300 font-mono">{timezone}</strong></span>
            </div>
          </div>
        </div>

        {/* View Switcher & Action Controls */}
        <div className="flex items-center gap-3 flex-wrap">
          {/* Refresh Toggle */}
          <button
            onClick={fetchAppointments}
            className="flex items-center gap-1.5 bg-slate-800/80 hover:bg-slate-700/80 text-slate-300 text-xs px-3 py-2 rounded-xl border border-slate-700/50 transition-colors"
            title="Refresh Schedule"
          >
            <RefreshCw className={`w-3.5 h-3.5 ${isLoading ? 'animate-spin text-violet-400' : ''}`} />
            <span className="hidden sm:inline">Refresh</span>
          </button>

          {/* View Mode Switcher */}
          <div className="flex bg-slate-800/80 p-1 rounded-xl border border-slate-700/50 text-xs font-medium">
            {(['day', 'week', 'month', 'list'] as const).map((mode) => (
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

          {/* Create Booking Button */}
          <button
            onClick={() => {
              setCreateSlotData(null);
              setIsCreateModalOpen(true);
            }}
            className="flex items-center gap-2 bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white text-xs font-bold px-4 py-2 rounded-xl transition-all shadow-lg shadow-violet-900/30 active:scale-95"
          >
            <Plus className="w-4 h-4" />
            New Appointment
          </button>
        </div>
      </div>

      {/* Filter Bar */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3 bg-slate-900/50 border border-slate-800 rounded-2xl p-4 shadow-lg backdrop-blur-md">
        {/* Search Input */}
        <div>
          <label className="block text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-1">Search</label>
          <input
            type="text"
            placeholder="Search customer name/email..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-1.5 text-xs text-slate-200 focus:outline-none focus:border-violet-500"
          />
        </div>

        {/* Staff Filter */}
        <div>
          <label className="block text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-1">Staff</label>
          <select
            value={selectedStaff}
            onChange={(e) => setSelectedStaff(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-1.5 text-xs text-slate-200 focus:outline-none focus:border-violet-500 cursor-pointer"
          >
            <option value="all">All Staff</option>
            {staffList.map((s) => (
              <option key={s.id} value={s.id}>{s.name}</option>
            ))}
          </select>
        </div>

        {/* Location Filter */}
        <div>
          <label className="block text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-1">Location</label>
          <select
            value={selectedLocation}
            onChange={(e) => setSelectedLocation(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-1.5 text-xs text-slate-200 focus:outline-none focus:border-violet-500 cursor-pointer"
          >
            <option value="all">All Locations</option>
            {locations.map((loc) => (
              <option key={loc.id} value={loc.id}>{loc.name}</option>
            ))}
          </select>
        </div>

        {/* Service Filter */}
        <div>
          <label className="block text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-1">Service</label>
          <select
            value={selectedService}
            onChange={(e) => setSelectedService(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-1.5 text-xs text-slate-200 focus:outline-none focus:border-violet-500 cursor-pointer"
          >
            <option value="all">All Services</option>
            {services.map((svc) => (
              <option key={svc.id} value={svc.id}>{svc.name}</option>
            ))}
          </select>
        </div>

        {/* Status Filter */}
        <div>
          <label className="block text-[10px] font-semibold text-slate-400 uppercase tracking-wider mb-1">Status</label>
          <select
            value={selectedStatus}
            onChange={(e) => setSelectedStatus(e.target.value)}
            className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-1.5 text-xs text-slate-200 focus:outline-none focus:border-violet-500 cursor-pointer"
          >
            <option value="all">All Statuses</option>
            <option value="PENDING">Pending</option>
            <option value="CONFIRMED">Confirmed</option>
            <option value="RESCHEDULED">Rescheduled</option>
            <option value="COMPLETED">Completed</option>
            <option value="CANCELLED">Cancelled</option>
            <option value="NO_SHOW">No Show</option>
          </select>
        </div>
      </div>

      {/* Conflict / Error Banner */}
      {conflictWarning && (
        <div className="bg-rose-500/10 border border-rose-500/30 rounded-2xl p-4 text-xs text-rose-300 flex items-center justify-between animate-fade-in">
          <div className="flex items-center gap-2">
            <AlertCircle className="w-4 h-4 text-rose-400 shrink-0" />
            <span>{conflictWarning}</span>
          </div>
          <button onClick={() => setConflictWarning(null)} className="text-rose-400 hover:text-rose-200">
            <X className="w-4 h-4" />
          </button>
        </div>
      )}

      {/* View Mode Container */}
      <div className="bg-slate-900/40 border border-slate-800/80 rounded-2xl overflow-hidden shadow-2xl backdrop-blur-md">
        
        {/* LIST VIEW */}
        {viewMode === 'list' && (
          <div className="p-4 space-y-4">
            <div className="flex items-center justify-between text-xs text-slate-400 px-2">
              <span>Showing {filteredAppointments.length} appointments</span>
            </div>

            <div className="divide-y divide-slate-800/60">
              {filteredAppointments.map((appt) => (
                <div
                  key={appt.id}
                  onClick={() => {
                    setActiveAppointment(appt);
                    setIsDetailOpen(true);
                  }}
                  className="p-4 hover:bg-slate-800/30 transition-colors rounded-xl cursor-pointer flex flex-col md:flex-row md:items-center justify-between gap-4"
                >
                  <div className="flex items-start gap-4">
                    <div className="w-10 h-10 rounded-xl bg-slate-800 border border-slate-700 flex items-center justify-center text-violet-400 font-bold">
                      {appt.guest_name?.charAt(0) || 'G'}
                    </div>
                    <div>
                      <div className="flex items-center gap-3">
                        <h4 className="font-semibold text-slate-100 text-sm">{appt.guest_name}</h4>
                        <span className={`text-[10px] font-mono font-bold px-2 py-0.5 rounded-full border ${getStatusBadge(appt.status)}`}>
                          {appt.status}
                        </span>
                      </div>
                      <p className="text-xs text-slate-400 mt-0.5">{appt.guest_email} • {appt.guest_phone}</p>
                      <div className="flex items-center gap-4 text-xs font-mono text-slate-400 pt-2">
                        <span className="flex items-center gap-1">
                          <Clock className="w-3.5 h-3.5 text-violet-400" />
                          {new Date(appt.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="text-right flex md:flex-col items-center md:items-end justify-between border-t md:border-t-0 border-slate-800/60 pt-2 md:pt-0">
                    <div className="text-sm font-bold text-emerald-400">
                      ${((appt.price_cents || 0) / 100).toFixed(2)}
                    </div>
                    <span className="text-[11px] text-slate-400 capitalize">{appt.source} booking</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* WEEK VIEW GRID */}
        {viewMode === 'week' && (
          <div>
            <div className="grid grid-cols-8 border-b border-slate-800 text-center bg-slate-950/60 font-medium text-xs text-slate-400 select-none">
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

            <div className="divide-y divide-slate-800/40 max-h-[650px] overflow-y-auto custom-scrollbar">
              {timeSlots.map((time, timeIdx) => (
                <div key={timeIdx} className="grid grid-cols-8 min-h-[72px]">
                  <div className="p-2 border-r border-slate-800/80 text-xs font-mono text-slate-500 text-right pr-3 pt-2 bg-slate-950/20 select-none">
                    {time}
                  </div>

                  {weekDays.map((day, dayIdx) => {
                    const dayStr = day.toISOString().split('T')[0];
                    const apptsInSlot = filteredAppointments.filter((a) => {
                      return a.start_time.startsWith(dayStr) && a.start_time.includes(`T${time}`);
                    });

                    return (
                      <div
                        key={dayIdx}
                        onClick={() => handleEmptySlotClick(dayStr, time)}
                        className="border-r border-slate-800/40 p-1 relative group hover:bg-slate-800/20 transition-colors cursor-pointer min-h-[72px]"
                      >
                        {apptsInSlot.map((appt) => (
                          <div
                            key={appt.id}
                            onClick={(e) => {
                              e.stopPropagation();
                              setActiveAppointment(appt);
                              setIsDetailOpen(true);
                            }}
                            className={`p-2 rounded-xl text-xs border shadow-md transition-all hover:scale-[1.02] cursor-pointer mb-1 ${getStatusBadge(appt.status)}`}
                          >
                            <div className="font-bold truncate text-[11px]">{appt.guest_name}</div>
                            <div className="text-[10px] opacity-80 font-mono">
                              {new Date(appt.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                            </div>
                          </div>
                        ))}

                        <div className="opacity-0 group-hover:opacity-100 transition-opacity absolute inset-1 border border-dashed border-violet-500/40 rounded-lg flex items-center justify-center bg-violet-500/5 pointer-events-none">
                          <Plus className="w-4 h-4 text-violet-400" />
                        </div>
                      </div>
                    );
                  })}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* DAY VIEW */}
        {viewMode === 'day' && (
          <div className="p-6 space-y-4">
            <h3 className="text-lg font-bold text-white">
              Schedule for {selectedDate.toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric', year: 'numeric' })}
            </h3>
            <div className="space-y-3">
              {timeSlots.map((time) => {
                const dayStr = selectedDate.toISOString().split('T')[0];
                const appts = filteredAppointments.filter((a) => a.start_time.startsWith(dayStr) && a.start_time.includes(`T${time}`));
                return (
                  <div key={time} className="flex gap-4 border-b border-slate-800/60 pb-3 items-center">
                    <div className="w-16 font-mono text-xs text-slate-400 font-semibold">{time}</div>
                    <div className="flex-1 min-h-[48px] border border-dashed border-slate-800 rounded-xl p-2 flex items-center gap-3">
                      {appts.length === 0 ? (
                        <button
                          onClick={() => handleEmptySlotClick(dayStr, time)}
                          className="text-xs text-slate-500 hover:text-violet-400 flex items-center gap-1"
                        >
                          <Plus className="w-3.5 h-3.5" /> Book slot at {time}
                        </button>
                      ) : (
                        appts.map((a) => (
                          <div
                            key={a.id}
                            onClick={() => {
                              setActiveAppointment(a);
                              setIsDetailOpen(true);
                            }}
                            className={`p-2.5 rounded-xl text-xs font-semibold border cursor-pointer ${getStatusBadge(a.status)}`}
                          >
                            {a.guest_name} — {a.status}
                          </div>
                        ))
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* MONTH VIEW */}
        {viewMode === 'month' && (
          <div className="p-6 text-center text-slate-400 text-sm space-y-4">
            <CalendarIcon className="w-10 h-10 text-violet-400 mx-auto" />
            <p className="font-semibold text-slate-200">Month Overview Grid</p>
            <div className="grid grid-cols-7 gap-2 text-xs">
              {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map((d) => (
                <div key={d} className="font-bold text-slate-500 uppercase pb-2">{d}</div>
              ))}
              {Array.from({ length: 31 }).map((_, i) => (
                <div key={i} className="bg-slate-950/60 border border-slate-800/80 rounded-xl h-20 p-2 text-left text-xs font-mono hover:border-violet-500/50 transition-colors">
                  <span className="text-slate-400">{i + 1}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* APPOINTMENT DETAIL DRAWER MODAL */}
      {isDetailOpen && activeAppointment && (
        <div className="fixed inset-0 z-50 bg-slate-950/80 backdrop-blur-md flex items-center justify-end p-4 sm:p-6 animate-fade-in">
          <div className="bg-slate-900 border border-slate-800 rounded-3xl w-full max-w-lg p-6 space-y-6 shadow-2xl overflow-y-auto max-h-[90vh]">
            <div className="flex items-center justify-between border-b border-slate-800 pb-4">
              <div>
                <span className={`text-[10px] font-mono font-bold px-3 py-1 rounded-full border uppercase tracking-wider ${getStatusBadge(activeAppointment.status)}`}>
                  {activeAppointment.status}
                </span>
                <h3 className="text-xl font-bold text-white mt-2">{activeAppointment.guest_name}</h3>
              </div>
              <button onClick={() => setIsDetailOpen(false)} className="text-slate-400 hover:text-white p-2 rounded-xl hover:bg-slate-800">
                <X className="w-5 h-5" />
              </button>
            </div>

            {/* Quick Details */}
            <div className="bg-slate-950/80 border border-slate-800/80 rounded-2xl p-4 space-y-3 text-xs">
              <div className="flex justify-between items-center pb-2 border-b border-slate-800">
                <span className="text-slate-400">Date & Time</span>
                <span className="font-mono text-slate-200 font-semibold">{new Date(activeAppointment.start_time).toLocaleString()}</span>
              </div>
              <div className="flex justify-between items-center pb-2 border-b border-slate-800">
                <span className="text-slate-400">Contact Email</span>
                <span className="text-slate-200">{activeAppointment.guest_email}</span>
              </div>
              <div className="flex justify-between items-center pb-2 border-b border-slate-800">
                <span className="text-slate-400">Phone</span>
                <span className="text-slate-200">{activeAppointment.guest_phone}</span>
              </div>
              <div className="flex justify-between items-center">
                <span className="text-slate-400">Price Snapshot</span>
                <span className="text-emerald-400 font-bold text-sm">${((activeAppointment.price_cents || 0) / 100).toFixed(2)}</span>
              </div>
            </div>

            {/* Notes */}
            {activeAppointment.notes && (
              <div className="bg-slate-950/40 border border-slate-800/60 rounded-xl p-3 text-xs text-slate-300">
                <span className="text-slate-500 block text-[10px] uppercase font-semibold mb-1">Customer Notes</span>
                <p>{activeAppointment.notes}</p>
              </div>
            )}

            {/* Quick Operational Action Buttons */}
            <div className="space-y-2 pt-2">
              <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider block">Operational Actions</span>
              <div className="grid grid-cols-2 gap-2">
                <button
                  onClick={() => handleUpdateStatus(activeAppointment.id, 'CONFIRMED')}
                  className="bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-bold py-2.5 rounded-xl transition-all shadow-md"
                >
                  Confirm Booking
                </button>
                <button
                  onClick={() => handleUpdateStatus(activeAppointment.id, 'COMPLETED')}
                  className="bg-blue-600 hover:bg-blue-500 text-white text-xs font-bold py-2.5 rounded-xl transition-all shadow-md"
                >
                  Mark Completed
                </button>
                <button
                  onClick={() => handleUpdateStatus(activeAppointment.id, 'NO_SHOW')}
                  className="bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-bold py-2.5 rounded-xl transition-all"
                >
                  Mark No-Show
                </button>
                <button
                  onClick={() => setShowCancelDialog(true)}
                  className="bg-rose-600/20 hover:bg-rose-600/40 text-rose-300 border border-rose-500/30 text-xs font-bold py-2.5 rounded-xl transition-all"
                >
                  Cancel Booking
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
