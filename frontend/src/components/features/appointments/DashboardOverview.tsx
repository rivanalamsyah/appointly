/**
 * DashboardOverview — Main dashboard React component.
 * Shows metrics cards, upcoming appointments, and activity feed.
 * Uses TanStack Query for data fetching.
 */
'use client';

import React from 'react';

// --- Metric Card Component ---------------------------------------------------

interface MetricCardProps {
  title: string;
  value: string | number;
  change?: string;
  changeType?: 'positive' | 'negative' | 'neutral';
  icon: React.ReactNode;
  gradient: string;
}

function MetricCard({ title, value, change, changeType = 'neutral', icon, gradient }: MetricCardProps) {
  const changeColors = {
    positive: 'text-emerald-400',
    negative: 'text-red-400',
    neutral: 'text-[var(--color-text-muted)]',
  };

  return (
    <div className="card group hover:border-[rgba(255,255,255,0.12)] transition-all duration-300 relative overflow-hidden">
      {/* Gradient accent */}
      <div className={`absolute inset-0 opacity-0 group-hover:opacity-100 transition-opacity duration-300 ${gradient}`} />

      <div className="relative z-10">
        <div className="flex items-center justify-between mb-4">
          <div className={`w-10 h-10 rounded-xl flex items-center justify-center ${gradient.replace('bg-gradient-to-br', 'bg-gradient-to-br')} opacity-80`}>
            {icon}
          </div>
          {change && (
            <span className={`text-xs font-medium ${changeColors[changeType]}`}>
              {change}
            </span>
          )}
        </div>
        <p className="text-[var(--color-text-muted)] text-sm font-medium">{title}</p>
        <p className="text-[var(--color-text-primary)] text-2xl font-bold mt-1 font-[var(--font-display)]">
          {value}
        </p>
      </div>
    </div>
  );
}

// --- Appointment Row Component -----------------------------------------------

interface AppointmentRowProps {
  customerName: string;
  service: string;
  staff: string;
  time: string;
  status: 'pending' | 'confirmed' | 'completed' | 'cancelled';
}

function AppointmentRow({ customerName, service, staff, time, status }: AppointmentRowProps) {
  const statusConfig = {
    pending:   { label: 'Pending',   className: 'badge-pending' },
    confirmed: { label: 'Confirmed', className: 'badge-confirmed' },
    completed: { label: 'Completed', className: 'badge-completed' },
    cancelled: { label: 'Cancelled', className: 'badge-cancelled' },
  };

  const s = statusConfig[status];

  return (
    <tr className="group">
      <td className="px-4 py-3">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-gradient-to-br from-indigo-500 to-violet-600 flex items-center justify-center text-white text-xs font-semibold flex-shrink-0">
            {customerName.charAt(0).toUpperCase()}
          </div>
          <div>
            <p className="text-sm font-medium text-[var(--color-text-primary)]">{customerName}</p>
            <p className="text-xs text-[var(--color-text-muted)]">{service}</p>
          </div>
        </div>
      </td>
      <td className="px-4 py-3">
        <p className="text-sm text-[var(--color-text-secondary)]">{staff}</p>
      </td>
      <td className="px-4 py-3">
        <p className="text-sm text-[var(--color-text-secondary)]">{time}</p>
      </td>
      <td className="px-4 py-3">
        <span className={`badge ${s.className}`}>{s.label}</span>
      </td>
      <td className="px-4 py-3 text-right">
        <button className="btn btn-ghost btn-sm opacity-0 group-hover:opacity-100 transition-opacity">
          View
        </button>
      </td>
    </tr>
  );
}

// --- Main Component ----------------------------------------------------------

export default function DashboardOverview() {
  // TODO: Replace with TanStack Query when backend is ready
  const metrics = [
    {
      title: 'Today\'s Appointments',
      value: '12',
      change: '+3 from yesterday',
      changeType: 'positive' as const,
      gradient: 'from-indigo-500/20 to-violet-500/10',
      icon: (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <rect width="18" height="18" x="3" y="4" rx="2" ry="2"/>
          <line x1="16" x2="16" y1="2" y2="6"/>
          <line x1="8" x2="8" y1="2" y2="6"/>
          <line x1="3" x2="21" y1="10" y2="10"/>
          <path d="m9 16 2 2 4-4"/>
        </svg>
      ),
    },
    {
      title: 'Total Customers',
      value: '284',
      change: '+12 this month',
      changeType: 'positive' as const,
      gradient: 'from-emerald-500/20 to-teal-500/10',
      icon: (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/>
          <circle cx="9" cy="7" r="4"/>
          <path d="M22 21v-2a4 4 0 0 0-3-3.87"/>
          <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
        </svg>
      ),
    },
    {
      title: 'Revenue This Month',
      value: 'Rp 8.4M',
      change: '+18% vs last month',
      changeType: 'positive' as const,
      gradient: 'from-amber-500/20 to-orange-500/10',
      icon: (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <line x1="12" x2="12" y1="2" y2="22"/>
          <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>
        </svg>
      ),
    },
    {
      title: 'Completion Rate',
      value: '94.2%',
      change: '+2.1% this week',
      changeType: 'positive' as const,
      gradient: 'from-blue-500/20 to-cyan-500/10',
      icon: (
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
          <path d="m9 11 3 3L22 4"/>
        </svg>
      ),
    },
  ];

  const upcomingAppointments: AppointmentRowProps[] = [
    { customerName: 'Rina Susanti', service: 'Haircut & Styling', staff: 'Maya', time: '09:00 AM', status: 'confirmed' },
    { customerName: 'Budi Santoso', service: 'Color Treatment', staff: 'Alex', time: '10:30 AM', status: 'confirmed' },
    { customerName: 'Dewi Rahayu', service: 'Manicure & Pedicure', staff: 'Sarah', time: '11:00 AM', status: 'pending' },
    { customerName: 'Ahmad Fauzi', service: 'Beard Trim', staff: 'David', time: '01:00 PM', status: 'confirmed' },
    { customerName: 'Sari Wulandari', service: 'Hair Treatment', staff: 'Maya', time: '02:30 PM', status: 'pending' },
  ];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-[var(--color-text-primary)]">
            Good morning! 👋
          </h1>
          <p className="text-sm text-[var(--color-text-muted)] mt-1">
            Friday, October 2, 2026 — Here's what's happening today.
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button className="btn btn-secondary btn-sm">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" x2="12" y1="15" y2="3"/>
            </svg>
            Export
          </button>
          <a href="/dashboard/appointments/new" className="btn btn-primary btn-sm">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M5 12h14"/><path d="M12 5v14"/>
            </svg>
            New Appointment
          </a>
        </div>
      </div>

      {/* Metrics Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {metrics.map((metric) => (
          <MetricCard key={metric.title} {...metric} />
        ))}
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">

        {/* Upcoming Appointments Table */}
        <div className="lg:col-span-2 card">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">
              Today's Appointments
            </h2>
            <a
              href="/dashboard/appointments"
              className="text-xs text-[var(--color-primary-400)] hover:text-[var(--color-primary-300)] transition-colors"
            >
              View all →
            </a>
          </div>
          <div className="overflow-x-auto -mx-6">
            <table className="data-table">
              <thead>
                <tr>
                  <th className="pl-6">Customer</th>
                  <th>Staff</th>
                  <th>Time</th>
                  <th>Status</th>
                  <th className="pr-6"></th>
                </tr>
              </thead>
              <tbody>
                {upcomingAppointments.map((appt, i) => (
                  <AppointmentRow key={i} {...appt} />
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Quick Stats Sidebar */}
        <div className="space-y-4">

          {/* Staff availability */}
          <div className="card">
            <h2 className="text-sm font-semibold text-[var(--color-text-primary)] mb-3">
              Staff Today
            </h2>
            <div className="space-y-2.5">
              {[
                { name: 'Maya', appointments: 4, available: true },
                { name: 'Alex', appointments: 3, available: true },
                { name: 'Sarah', appointments: 2, available: true },
                { name: 'David', appointments: 3, available: false },
              ].map((s) => (
                <div key={s.name} className="flex items-center gap-3">
                  <div className="w-7 h-7 rounded-full bg-gradient-to-br from-indigo-400/30 to-violet-400/30 flex items-center justify-center text-[10px] font-semibold text-[var(--color-primary-400)] flex-shrink-0">
                    {s.name[0]}
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-xs font-medium text-[var(--color-text-primary)]">{s.name}</p>
                    <p className="text-[10px] text-[var(--color-text-muted)]">{s.appointments} appts</p>
                  </div>
                  <div className={`status-dot ${s.available ? 'online' : 'offline'}`} />
                </div>
              ))}
            </div>
          </div>

          {/* Quick Actions */}
          <div className="card">
            <h2 className="text-sm font-semibold text-[var(--color-text-primary)] mb-3">
              Quick Actions
            </h2>
            <div className="space-y-1.5">
              {[
                { label: 'New Appointment', href: '/dashboard/appointments/new', icon: '+' },
                { label: 'Add Customer', href: '/dashboard/customers/new', icon: '+' },
                { label: 'View Calendar', href: '/dashboard/calendar', icon: '→' },
                { label: 'Check Reports', href: '/dashboard/reports', icon: '→' },
              ].map((action) => (
                <a
                  key={action.label}
                  href={action.href}
                  className="flex items-center justify-between p-2.5 rounded-lg hover:bg-[rgba(255,255,255,0.04)] transition-colors group"
                >
                  <span className="text-sm text-[var(--color-text-secondary)] group-hover:text-[var(--color-text-primary)] transition-colors">
                    {action.label}
                  </span>
                  <span className="text-[var(--color-text-muted)] group-hover:text-[var(--color-primary-400)] transition-colors text-sm">
                    {action.icon}
                  </span>
                </a>
              ))}
            </div>
          </div>

        </div>
      </div>
    </div>
  );
}
