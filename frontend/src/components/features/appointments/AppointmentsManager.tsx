import React, { useState } from 'react';
import {
  CalendarCheck,
  Plus,
  Search,
  Filter,
  Clock,
  User,
  Scissors,
  CheckCircle2,
  XCircle,
  AlertCircle,
  RefreshCw,
  MoreVertical,
  X,
  Sparkles,
  Calendar as CalendarIcon,
  DollarSign,
  AlertTriangle,
  FileText
} from 'lucide-react';
import type { Appointment, AppointmentStatus } from '@/types/api';

interface StatusHistoryItem {
  from_status: string;
  to_status: string;
  reason: string;
  changed_at: string;
}

interface AppointmentUI extends Appointment {
  history?: StatusHistoryItem[];
  staff_name?: string;
  service_name?: string;
}

const INITIAL_APPOINTMENTS: AppointmentUI[] = [
  {
    id: 'appt-101',
    organization_id: 'org-demo',
    location_id: 'loc-1',
    service_id: 'svc-1',
    staff_id: 'st-1',
    staff_name: 'John Stylist',
    service_name: 'Signature Haircut & Styling',
    guest_name: 'David Beckham',
    guest_email: 'david@example.com',
    guest_phone: '+1 555-019-9988',
    start_time: new Date(Date.now() + 2 * 3600000).toISOString(),
    end_time: new Date(Date.now() + 3 * 3600000).toISOString(),
    timezone: 'UTC',
    status: 'confirmed',
    price_cents: 4500,
    currency: 'USD',
    notes: 'Prefers classic pompadour style.',
    source: 'online',
    created_at: new Date(Date.now() - 86400000).toISOString(),
    updated_at: new Date().toISOString(),
    history: [
      { from_status: 'pending', to_status: 'confirmed', reason: 'Auto-confirmed by system policy', changed_at: new Date(Date.now() - 86400000).toISOString() }
    ]
  },
  {
    id: 'appt-102',
    organization_id: 'org-demo',
    location_id: 'loc-1',
    service_id: 'svc-2',
    staff_id: 'st-2',
    staff_name: 'Sarah Facial Specialist',
    service_name: 'HydraFacial MD Rejuvenation',
    guest_name: 'Sophia Loren',
    guest_email: 'sophia@example.com',
    guest_phone: '+1 555-084-2211',
    start_time: new Date(Date.now() + 24 * 3600000).toISOString(),
    end_time: new Date(Date.now() + 25 * 3600000).toISOString(),
    timezone: 'UTC',
    status: 'pending',
    price_cents: 12000,
    currency: 'USD',
    notes: 'Sensitive skin treatment.',
    source: 'online',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    history: []
  },
  {
    id: 'appt-103',
    organization_id: 'org-demo',
    location_id: 'loc-1',
    service_id: 'svc-1',
    staff_id: 'st-1',
    staff_name: 'John Stylist',
    service_name: 'Beard Trim & Grooming',
    guest_name: 'Robert De Niro',
    guest_email: 'robert@example.com',
    guest_phone: '+1 555-392-8811',
    start_time: new Date(Date.now() - 48 * 3600000).toISOString(),
    end_time: new Date(Date.now() - 47 * 3600000).toISOString(),
    timezone: 'UTC',
    status: 'completed',
    price_cents: 3000,
    currency: 'USD',
    source: 'manual',
    created_at: new Date(Date.now() - 90000000).toISOString(),
    updated_at: new Date(Date.now() - 47 * 3600000).toISOString(),
    history: [
      { from_status: 'confirmed', to_status: 'completed', reason: 'Service delivered', changed_at: new Date(Date.now() - 47 * 3600000).toISOString() }
    ]
  }
];

export default function AppointmentsManager() {
  const [appointments, setAppointments] = useState<AppointmentUI[]>(INITIAL_APPOINTMENTS);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedStatus, setSelectedStatus] = useState<string>('ALL');

  // Selected appointment for detail drawer
  const [selectedAppt, setSelectedAppt] = useState<AppointmentUI | null>(null);

  // Modals state
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [isRescheduleOpen, setIsRescheduleOpen] = useState(false);
  const [isCancelOpen, setIsCancelOpen] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  // Form inputs
  const [formGuestName, setFormGuestName] = useState('');
  const [formGuestEmail, setFormGuestEmail] = useState('');
  const [formGuestPhone, setFormGuestPhone] = useState('');
  const [formServiceName, setFormServiceName] = useState('Signature Haircut & Styling');
  const [formStaffName, setFormStaffName] = useState('John Stylist');
  const [formDate, setFormDate] = useState('');
  const [formTime, setFormTime] = useState('10:00');
  const [formNotes, setFormNotes] = useState('');

  // Reschedule & Cancel inputs
  const [rescheduleDate, setRescheduleDate] = useState('');
  const [rescheduleTime, setRescheduleTime] = useState('14:00');
  const [rescheduleReason, setRescheduleReason] = useState('');
  const [cancelReason, setCancelReason] = useState('');

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3000);
  };

  const handleCreateAppointment = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formGuestName.trim()) return;

    const startTime = new Date(`${formDate || new Date().toISOString().split('T')[0]}T${formTime}:00Z`);
    const endTime = new Date(startTime.getTime() + 60 * 60000);

    const newAppt: AppointmentUI = {
      id: `appt-${Date.now()}`,
      organization_id: 'org-demo',
      location_id: 'loc-1',
      service_id: 'svc-1',
      staff_id: 'st-1',
      staff_name: formStaffName,
      service_name: formServiceName,
      guest_name: formGuestName.trim(),
      guest_email: formGuestEmail.trim(),
      guest_phone: formGuestPhone.trim(),
      start_time: startTime.toISOString(),
      end_time: endTime.toISOString(),
      timezone: 'UTC',
      status: 'confirmed',
      price_cents: 4500,
      currency: 'USD',
      notes: formNotes.trim(),
      source: 'manual',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      history: [
        { from_status: '', to_status: 'confirmed', reason: 'Admin manual booking', changed_at: new Date().toISOString() }
      ]
    };

    setAppointments(prev => [newAppt, ...prev]);
    setIsCreateOpen(false);
    showToast(`Appointment created for "${formGuestName.trim()}".`);
  };

  const handleUpdateStatus = (apptId: string, newStatus: AppointmentStatus, reason: string) => {
    setAppointments(prev =>
      prev.map(a => {
        if (a.id === apptId) {
          const oldStatus = a.status;
          const updated = {
            ...a,
            status: newStatus,
            updated_at: new Date().toISOString(),
            history: [
              ...(a.history || []),
              { from_status: oldStatus, to_status: newStatus, reason: reason || 'Status updated by staff', changed_at: new Date().toISOString() }
            ]
          };
          if (selectedAppt?.id === apptId) setSelectedAppt(updated);
          return updated;
        }
        return a;
      })
    );
    showToast(`Appointment status changed to ${newStatus}.`);
  };

  const handleReschedule = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedAppt || !rescheduleDate) return;

    const newStart = new Date(`${rescheduleDate}T${rescheduleTime}:00Z`);
    const duration = new Date(selectedAppt.end_time).getTime() - new Date(selectedAppt.start_time).getTime();
    const newEnd = new Date(newStart.getTime() + duration);

    handleUpdateStatus(selectedAppt.id, 'rescheduled', rescheduleReason || 'Customer requested time change');

    setAppointments(prev =>
      prev.map(a =>
        a.id === selectedAppt.id
          ? {
              ...a,
              start_time: newStart.toISOString(),
              end_time: newEnd.toISOString()
            }
          : a
      )
    );

    setIsRescheduleOpen(false);
    showToast('Appointment rescheduled successfully.');
  };

  const handleCancel = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedAppt) return;
    handleUpdateStatus(selectedAppt.id, 'cancelled', cancelReason || 'Cancelled by staff');
    setIsCancelOpen(false);
  };

  const filteredAppointments = appointments.filter(a => {
    if (selectedStatus !== 'ALL' && a.status !== selectedStatus) return false;
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      const matchGuest = a.guest_name?.toLowerCase().includes(q) || false;
      const matchService = a.service_name?.toLowerCase().includes(q) || false;
      const matchStaff = a.staff_name?.toLowerCase().includes(q) || false;
      if (!matchGuest && !matchService && !matchStaff) return false;
    }
    return true;
  });

  const getStatusBadge = (status: AppointmentStatus) => {
    switch (status) {
      case 'confirmed':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <CheckCircle2 className="w-3 h-3" /> Confirmed
          </span>
        );
      case 'pending':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <Clock className="w-3 h-3" /> Pending
          </span>
        );
      case 'rescheduled':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
            <RefreshCw className="w-3 h-3" /> Rescheduled
          </span>
        );
      case 'completed':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-cyan-500/10 text-cyan-400 border border-cyan-500/20">
            <CheckCircle2 className="w-3 h-3" /> Completed
          </span>
        );
      case 'cancelled':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-red-500/10 text-red-400 border border-red-500/20">
            <XCircle className="w-3 h-3" /> Cancelled
          </span>
        );
      case 'no_show':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-slate-500/10 text-slate-400 border border-slate-500/20">
            <AlertCircle className="w-3 h-3" /> No Show
          </span>
        );
    }
  };

  const formatMoney = (cents: number) => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(cents / 100);
  };

  return (
    <div className="space-y-6">
      {/* Toast */}
      {toastMessage && (
        <div className="fixed top-5 right-5 z-50 flex items-center gap-3 px-4 py-3 rounded-lg bg-[var(--color-surface-raised)] border border-indigo-500/40 text-white shadow-xl animate-fade-in">
          <Sparkles className="w-5 h-5 text-indigo-400" />
          <span className="text-sm font-medium">{toastMessage}</span>
        </div>
      )}

      {/* Header Banner */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-6 rounded-2xl bg-gradient-to-r from-slate-900 via-indigo-950/40 to-slate-900 border border-[var(--color-surface-border)] shadow-xl relative overflow-hidden">
        <div className="relative z-10">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400">
              <CalendarCheck className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-white font-[var(--font-display)]">Appointments Transactional Workflow</h1>
              <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                Core booking transaction management, status transitions, rescheduling, and cancellation controls.
              </p>
            </div>
          </div>
        </div>
        <div className="relative z-10">
          <button
            onClick={() => setIsCreateOpen(true)}
            className="flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 shadow-lg shadow-indigo-600/30 transition-all duration-150"
          >
            <Plus className="w-4 h-4" /> Create Booking
          </button>
        </div>
      </div>

      {/* Filter Bar */}
      <div className="p-4 rounded-xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] flex flex-wrap items-center gap-4">
        {/* Search */}
        <div className="relative flex-1 min-w-[240px]">
          <Search className="w-4 h-4 text-[var(--color-text-muted)] absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            placeholder="Search by customer, service, or staff..."
            className="w-full pl-10 pr-4 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white placeholder-[var(--color-text-muted)] focus:outline-none focus:border-indigo-500"
          />
        </div>

        {/* Status Filter */}
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-[var(--color-text-muted)]" />
          <select
            value={selectedStatus}
            onChange={e => setSelectedStatus(e.target.value)}
            className="px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
          >
            <option value="ALL">All Statuses</option>
            <option value="confirmed">Confirmed</option>
            <option value="pending">Pending</option>
            <option value="rescheduled">Rescheduled</option>
            <option value="completed">Completed</option>
            <option value="cancelled">Cancelled</option>
            <option value="no_show">No Show</option>
          </select>
        </div>
      </div>

      {/* Appointments List & Detail Column */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* List Column */}
        <div className={`space-y-3 ${selectedAppt ? 'lg:col-span-1' : 'lg:col-span-3'}`}>
          {filteredAppointments.length === 0 ? (
            <div className="p-12 text-center rounded-2xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)]">
              <CalendarIcon className="w-12 h-12 text-[var(--color-text-muted)] mx-auto mb-3 opacity-50" />
              <h3 className="text-sm font-semibold text-white">No appointments found</h3>
              <p className="text-xs text-[var(--color-text-muted)] mt-1">Try adjusting search query or status filter.</p>
            </div>
          ) : (
            filteredAppointments.map(a => {
              const isSelected = selectedAppt?.id === a.id;
              const startStr = new Date(a.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
              const dateStr = new Date(a.start_time).toLocaleDateString([], { month: 'short', day: 'numeric', year: 'numeric' });

              return (
                <div
                  key={a.id}
                  onClick={() => setSelectedAppt(a)}
                  className={`p-4 rounded-xl border transition-all cursor-pointer ${
                    isSelected
                      ? 'bg-indigo-500/10 border-indigo-500/40 shadow-lg'
                      : 'bg-[var(--color-surface-raised)] border-[var(--color-surface-border)] hover:border-slate-700'
                  }`}
                >
                  <div className="flex items-start justify-between gap-3 mb-2">
                    <div>
                      <h3 className="text-sm font-bold text-white font-[var(--font-display)]">{a.guest_name}</h3>
                      <p className="text-xs text-indigo-400 font-medium">{a.service_name}</p>
                    </div>
                    {getStatusBadge(a.status)}
                  </div>

                  <div className="grid grid-cols-2 gap-2 text-[11px] text-[var(--color-text-muted)] pt-3 border-t border-[var(--color-surface-border)]">
                    <div className="flex items-center gap-1.5">
                      <CalendarIcon className="w-3.5 h-3.5 text-indigo-400 flex-shrink-0" />
                      <span>{dateStr}</span>
                    </div>
                    <div className="flex items-center gap-1.5">
                      <Clock className="w-3.5 h-3.5 text-indigo-400 flex-shrink-0" />
                      <span>{startStr}</span>
                    </div>
                    <div className="flex items-center gap-1.5">
                      <User className="w-3.5 h-3.5 text-indigo-400 flex-shrink-0" />
                      <span className="truncate">{a.staff_name}</span>
                    </div>
                    <div className="flex items-center gap-1.5 text-emerald-400 font-medium">
                      <DollarSign className="w-3.5 h-3.5 flex-shrink-0" />
                      <span>{formatMoney(a.price_cents)}</span>
                    </div>
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Appointment Detail & Status History Drawer */}
        {selectedAppt && (
          <div className="lg:col-span-2 p-6 rounded-2xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] shadow-xl space-y-6 animate-fade-in">
            <div className="flex items-start justify-between pb-4 border-b border-[var(--color-surface-border)]">
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-lg font-bold text-white font-[var(--font-display)]">{selectedAppt.guest_name}</h2>
                  {getStatusBadge(selectedAppt.status)}
                </div>
                <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                  ID: <code className="text-indigo-300 font-mono text-[10px]">{selectedAppt.id}</code> • Source: {selectedAppt.source}
                </p>
              </div>
              <button onClick={() => setSelectedAppt(null)} className="p-1 rounded-lg text-[var(--color-text-muted)] hover:text-white">
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Timing & Financial Summary */}
            <div className="grid grid-cols-3 gap-3">
              <div className="p-3.5 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
                <p className="text-[10px] uppercase font-semibold text-[var(--color-text-muted)] mb-1">Date & Time</p>
                <p className="text-xs text-white font-medium">
                  {new Date(selectedAppt.start_time).toLocaleDateString()} at{' '}
                  {new Date(selectedAppt.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </p>
              </div>
              <div className="p-3.5 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
                <p className="text-[10px] uppercase font-semibold text-[var(--color-text-muted)] mb-1">Assigned Staff</p>
                <p className="text-xs text-white font-medium">{selectedAppt.staff_name}</p>
              </div>
              <div className="p-3.5 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
                <p className="text-[10px] uppercase font-semibold text-[var(--color-text-muted)] mb-1">Price Snapshot</p>
                <p className="text-xs text-emerald-400 font-bold">{formatMoney(selectedAppt.price_cents)}</p>
              </div>
            </div>

            {/* Action Toolbar */}
            <div className="p-4 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] space-y-2">
              <h4 className="text-xs font-bold text-white">Status Transition Actions</h4>
              <div className="flex flex-wrap items-center gap-2 pt-1">
                {selectedAppt.status === 'pending' && (
                  <button
                    onClick={() => handleUpdateStatus(selectedAppt.id, 'confirmed', 'Confirmed by admin')}
                    className="px-3 py-1.5 rounded-lg text-xs font-semibold text-white bg-emerald-600 hover:bg-emerald-500 shadow-md shadow-emerald-600/30"
                  >
                    Confirm Booking
                  </button>
                )}
                {selectedAppt.status !== 'completed' && selectedAppt.status !== 'cancelled' && (
                  <>
                    <button
                      onClick={() => handleUpdateStatus(selectedAppt.id, 'completed', 'Service completed')}
                      className="px-3 py-1.5 rounded-lg text-xs font-semibold text-white bg-cyan-600 hover:bg-cyan-500 shadow-md shadow-cyan-600/30"
                    >
                      Complete Service
                    </button>
                    <button
                      onClick={() => {
                        setRescheduleDate(selectedAppt.start_time.split('T')[0]);
                        setIsRescheduleOpen(true);
                      }}
                      className="px-3 py-1.5 rounded-lg text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 shadow-md shadow-indigo-600/30"
                    >
                      Reschedule
                    </button>
                    <button
                      onClick={() => handleUpdateStatus(selectedAppt.id, 'no_show', 'Customer did not show up')}
                      className="px-3 py-1.5 rounded-lg text-xs font-semibold text-slate-300 bg-slate-800 hover:bg-slate-700 border border-slate-700"
                    >
                      Mark No-Show
                    </button>
                    <button
                      onClick={() => setIsCancelOpen(true)}
                      className="px-3 py-1.5 rounded-lg text-xs font-semibold text-red-400 bg-red-500/10 hover:bg-red-500/20 border border-red-500/30"
                    >
                      Cancel Booking
                    </button>
                  </>
                )}
              </div>
            </div>

            {/* Status History Audit Trail */}
            <div className="space-y-3">
              <h4 className="text-xs font-bold text-white flex items-center gap-2">
                <FileText className="w-4 h-4 text-indigo-400" /> Status History Audit Trail
              </h4>
              <div className="space-y-2">
                {(!selectedAppt.history || selectedAppt.history.length === 0) ? (
                  <p className="text-xs text-[var(--color-text-muted)] italic">No status changes recorded yet.</p>
                ) : (
                  selectedAppt.history.map((h, idx) => (
                    <div key={idx} className="p-3 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-xs space-y-1">
                      <div className="flex items-center justify-between text-[11px]">
                        <span className="font-semibold text-indigo-300">
                          {h.from_status ? `${h.from_status} → ${h.to_status}` : `Initial: ${h.to_status}`}
                        </span>
                        <span className="text-[var(--color-text-muted)]">{new Date(h.changed_at).toLocaleString()}</span>
                      </div>
                      <p className="text-[var(--color-text-muted)]">{h.reason}</p>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Create Booking Modal */}
      {isCreateOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
          <div className="w-full max-w-md rounded-2xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] shadow-2xl p-6">
            <div className="flex items-center justify-between pb-4 mb-4 border-b border-[var(--color-surface-border)]">
              <h2 className="text-base font-bold text-white font-[var(--font-display)]">Create Manual Booking</h2>
              <button onClick={() => setIsCreateOpen(false)} className="p-1 rounded-lg text-[var(--color-text-muted)] hover:text-white">
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleCreateAppointment} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Customer Name *</label>
                <input
                  type="text"
                  required
                  value={formGuestName}
                  onChange={e => setFormGuestName(e.target.value)}
                  placeholder="Full Name"
                  className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Email</label>
                  <input
                    type="email"
                    value={formGuestEmail}
                    onChange={e => setFormGuestEmail(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Phone</label>
                  <input
                    type="text"
                    value={formGuestPhone}
                    onChange={e => setFormGuestPhone(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Date</label>
                  <input
                    type="date"
                    required
                    value={formDate}
                    onChange={e => setFormDate(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Time</label>
                  <input
                    type="time"
                    required
                    value={formTime}
                    onChange={e => setFormTime(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              <div className="flex items-center justify-end gap-3 pt-4 border-t border-[var(--color-surface-border)]">
                <button type="button" onClick={() => setIsCreateOpen(false)} className="px-4 py-2 text-xs text-[var(--color-text-muted)] hover:text-white">
                  Cancel
                </button>
                <button type="submit" className="px-4 py-2 text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg shadow-md shadow-indigo-600/30">
                  Save Appointment
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Reschedule Modal */}
      {isRescheduleOpen && selectedAppt && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
          <div className="w-full max-w-md rounded-2xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] shadow-2xl p-6">
            <h2 className="text-base font-bold text-white font-[var(--font-display)] mb-4">Reschedule Appointment</h2>
            <form onSubmit={handleReschedule} className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">New Date</label>
                  <input
                    type="date"
                    required
                    value={rescheduleDate}
                    onChange={e => setRescheduleDate(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">New Time</label>
                  <input
                    type="time"
                    required
                    value={rescheduleTime}
                    onChange={e => setRescheduleTime(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Reschedule Reason</label>
                <input
                  type="text"
                  value={rescheduleReason}
                  onChange={e => setRescheduleReason(e.target.value)}
                  placeholder="Optional reason..."
                  className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                />
              </div>
              <div className="flex justify-end gap-3 pt-4 border-t border-[var(--color-surface-border)]">
                <button type="button" onClick={() => setIsRescheduleOpen(false)} className="px-4 py-2 text-xs text-[var(--color-text-muted)] hover:text-white">
                  Cancel
                </button>
                <button type="submit" className="px-4 py-2 text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg">
                  Confirm Reschedule
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Cancel Modal */}
      {isCancelOpen && selectedAppt && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
          <div className="w-full max-w-md rounded-2xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] shadow-2xl p-6">
            <h2 className="text-base font-bold text-white font-[var(--font-display)] mb-2">Cancel Appointment</h2>
            <p className="text-xs text-[var(--color-text-muted)] mb-4">
              Are you sure you want to cancel the booking for <strong className="text-white">{selectedAppt.guest_name}</strong>?
            </p>
            <form onSubmit={handleCancel} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Cancellation Reason</label>
                <input
                  type="text"
                  required
                  value={cancelReason}
                  onChange={e => setCancelReason(e.target.value)}
                  placeholder="e.g. Customer request or emergency"
                  className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                />
              </div>
              <div className="flex justify-end gap-3 pt-4 border-t border-[var(--color-surface-border)]">
                <button type="button" onClick={() => setIsCancelOpen(false)} className="px-4 py-2 text-xs text-[var(--color-text-muted)] hover:text-white">
                  Close
                </button>
                <button type="submit" className="px-4 py-2 text-xs font-semibold text-white bg-red-600 hover:bg-red-500 rounded-lg">
                  Confirm Cancellation
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
