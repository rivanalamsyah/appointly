import React, { useState } from 'react';
import {
  Users,
  Plus,
  Search,
  Filter,
  Mail,
  Phone,
  Calendar,
  DollarSign,
  UserCheck,
  UserX,
  ShieldAlert,
  Clock,
  MessageSquare,
  History,
  Link,
  Edit2,
  Trash2,
  X,
  Sparkles,
  ChevronRight,
  Tag,
  AlertTriangle
} from 'lucide-react';
import type { Customer, CustomerStatus, CustomerSource } from '@/types/api';

// Extended type with notes timeline for UI
interface CustomerUI extends Customer {
  notes_list?: { id: string; author: string; content: string; created_at: string }[];
}

const INITIAL_CUSTOMERS: CustomerUI[] = [
  {
    id: 'cust-1',
    organization_id: 'org-demo',
    first_name: 'Jessica',
    last_name: 'Alba',
    email: 'jessica.alba@example.com',
    phone: '+1 555-019-2831',
    notes: 'Prefers organic skincare products. Regular facial client.',
    status: 'ACTIVE',
    source: 'ONLINE_BOOKING',
    total_appointments: 8,
    completed_appointments: 8,
    no_show_count: 0,
    total_spent_cents: 120000, // $1,200.00
    created_at: new Date(Date.now() - 60 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
    notes_list: [
      {
        id: 'n-1',
        author: 'Sarah (Front Desk)',
        content: 'Client requested allergic reaction patch test before chemical peel.',
        created_at: new Date(Date.now() - 5 * 86400000).toISOString()
      },
      {
        id: 'n-2',
        author: 'Dr. Michael',
        content: 'Completed 60-min HydraFacial treatment. Skin condition excellent.',
        created_at: new Date(Date.now() - 15 * 86400000).toISOString()
      }
    ]
  },
  {
    id: 'cust-2',
    organization_id: 'org-demo',
    first_name: 'Marcus',
    last_name: 'Vance',
    email: 'marcus.vance@example.com',
    phone: '+1 555-084-9923',
    notes: 'Barber VIP membership. Always books Saturday mornings.',
    status: 'ACTIVE',
    source: 'WALK_IN',
    total_appointments: 14,
    completed_appointments: 14,
    no_show_count: 0,
    total_spent_cents: 98000, // $980.00
    created_at: new Date(Date.now() - 120 * 86400000).toISOString(),
    updated_at: new Date().toISOString(),
    notes_list: [
      {
        id: 'n-3',
        author: 'Alex (Barber)',
        content: 'Prefers skin fade on sides and textured top cut.',
        created_at: new Date(Date.now() - 10 * 86400000).toISOString()
      }
    ]
  },
  {
    id: 'cust-3',
    organization_id: 'org-demo',
    first_name: 'Elena',
    last_name: 'Rostova',
    email: 'elena.rostova@example.com',
    phone: '+1 555-492-1048',
    notes: 'Late cancellation warning issued on Sep 20.',
    status: 'INACTIVE',
    source: 'MANUAL',
    total_appointments: 3,
    completed_appointments: 2,
    no_show_count: 1,
    total_spent_cents: 24000,
    created_at: new Date(Date.now() - 30 * 86400000).toISOString(),
    updated_at: new Date().toISOString()
  }
];

export default function CustomersManager() {
  const [customers, setCustomers] = useState<CustomerUI[]>(INITIAL_CUSTOMERS);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedStatus, setSelectedStatus] = useState<string>('ALL');
  const [selectedSource, setSelectedSource] = useState<string>('ALL');
  const [sortBy, setSortBy] = useState<'created_at' | 'name' | 'appointments'>('created_at');

  // Selected customer for detail drawer
  const [selectedCustomer, setSelectedCustomer] = useState<CustomerUI | null>(null);
  const [detailTab, setDetailTab] = useState<'profile' | 'notes' | 'history'>('profile');

  // Form modal
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingCustomer, setEditingCustomer] = useState<CustomerUI | null>(null);
  const [formFirstName, setFormFirstName] = useState('');
  const [formLastName, setFormLastName] = useState('');
  const [formEmail, setFormEmail] = useState('');
  const [formPhone, setFormPhone] = useState('');
  const [formNotes, setFormNotes] = useState('');
  const [formStatus, setFormStatus] = useState<CustomerStatus>('ACTIVE');
  const [formSource, setFormSource] = useState<CustomerSource>('MANUAL');
  const [formError, setFormError] = useState<string | null>(null);

  // New Note input in detail drawer
  const [newNoteText, setNewNoteText] = useState('');

  // Toast
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3000);
  };

  const handleOpenAdd = () => {
    setEditingCustomer(null);
    setFormFirstName('');
    setFormLastName('');
    setFormEmail('');
    setFormPhone('');
    setFormNotes('');
    setFormStatus('ACTIVE');
    setFormSource('MANUAL');
    setFormError(null);
    setIsFormOpen(true);
  };

  const handleOpenEdit = (c: CustomerUI) => {
    setEditingCustomer(c);
    setFormFirstName(c.first_name);
    setFormLastName(c.last_name);
    setFormEmail(c.email || '');
    setFormPhone(c.phone || '');
    setFormNotes(c.notes || '');
    setFormStatus(c.status || 'ACTIVE');
    setFormSource(c.source || 'MANUAL');
    setFormError(null);
    setIsFormOpen(true);
  };

  const handleSaveCustomer = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formFirstName.trim()) {
      setFormError('First name is required.');
      return;
    }

    if (editingCustomer) {
      setCustomers(prev =>
        prev.map(c =>
          c.id === editingCustomer.id
            ? {
                ...c,
                first_name: formFirstName.trim(),
                last_name: formLastName.trim(),
                email: formEmail.trim(),
                phone: formPhone.trim(),
                notes: formNotes.trim(),
                status: formStatus,
                source: formSource,
                updated_at: new Date().toISOString()
              }
            : c
        )
      );
      showToast(`Customer "${formFirstName} ${formLastName}" updated.`);
    } else {
      const newC: CustomerUI = {
        id: `cust-${Date.now()}`,
        organization_id: 'org-demo',
        first_name: formFirstName.trim(),
        last_name: formLastName.trim(),
        email: formEmail.trim(),
        phone: formPhone.trim(),
        notes: formNotes.trim(),
        status: formStatus,
        source: formSource,
        total_appointments: 0,
        completed_appointments: 0,
        no_show_count: 0,
        total_spent_cents: 0,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        notes_list: []
      };
      setCustomers(prev => [newC, ...prev]);
      showToast(`Customer "${formFirstName} ${formLastName}" created.`);
    }
    setIsFormOpen(false);
  };

  const handleDeleteCustomer = (id: string, name: string) => {
    if (confirm(`Are you sure you want to delete customer "${name}"?`)) {
      setCustomers(prev => prev.filter(c => c.id !== id));
      if (selectedCustomer?.id === id) setSelectedCustomer(null);
      showToast(`Customer "${name}" removed.`);
    }
  };

  const handleAddNoteToTimeline = () => {
    if (!selectedCustomer || !newNoteText.trim()) return;

    const newNoteObj = {
      id: `n-${Date.now()}`,
      author: 'Current Staff',
      content: newNoteText.trim(),
      created_at: new Date().toISOString()
    };

    setCustomers(prev =>
      prev.map(c =>
        c.id === selectedCustomer.id
          ? {
              ...c,
              notes_list: [newNoteObj, ...(c.notes_list || [])]
            }
          : c
      )
    );

    setSelectedCustomer(prev => (prev ? { ...prev, notes_list: [newNoteObj, ...(prev.notes_list || [])] } : null));
    setNewNoteText('');
    showToast('Activity note added to customer timeline.');
  };

  const filteredCustomers = customers
    .filter(c => {
      if (selectedStatus !== 'ALL' && c.status !== selectedStatus) return false;
      if (selectedSource !== 'ALL' && c.source !== selectedSource) return false;
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const fullName = `${c.first_name} ${c.last_name}`.toLowerCase();
        const emailMatch = c.email?.toLowerCase().includes(q) || false;
        const phoneMatch = c.phone?.includes(q) || false;
        if (!fullName.includes(q) && !emailMatch && !phoneMatch) return false;
      }
      return true;
    })
    .sort((a, b) => {
      if (sortBy === 'name') return a.first_name.localeCompare(b.first_name);
      if (sortBy === 'appointments') return b.total_appointments - a.total_appointments;
      return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
    });

  const getStatusBadge = (status: CustomerStatus) => {
    switch (status) {
      case 'ACTIVE':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <UserCheck className="w-3 h-3" /> Active
          </span>
        );
      case 'INACTIVE':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-slate-500/10 text-slate-400 border border-slate-500/20">
            <UserX className="w-3 h-3" /> Inactive
          </span>
        );
      case 'BLOCKED':
        return (
          <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[11px] font-medium bg-red-500/10 text-red-400 border border-red-500/20">
            <ShieldAlert className="w-3 h-3" /> Blocked
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

      {/* Banner */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-6 rounded-2xl bg-gradient-to-r from-slate-900 via-indigo-950/40 to-slate-900 border border-[var(--color-surface-border)] shadow-xl relative overflow-hidden">
        <div className="relative z-10">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400">
              <Users className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-white font-[var(--font-display)]">Customer CRM</h1>
              <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                Manage tenant customer profiles, history, timeline notes, and contact channels.
              </p>
            </div>
          </div>
        </div>
        <div className="relative z-10">
          <button
            onClick={handleOpenAdd}
            className="flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 shadow-lg shadow-indigo-600/30 transition-all duration-150"
          >
            <Plus className="w-4 h-4" /> Add Customer
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
            placeholder="Search by customer name, email, or phone..."
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
            <option value="ACTIVE">Active</option>
            <option value="INACTIVE">Inactive</option>
            <option value="BLOCKED">Blocked</option>
          </select>
        </div>

        {/* Source Filter */}
        <select
          value={selectedSource}
          onChange={e => setSelectedSource(e.target.value)}
          className="px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
        >
          <option value="ALL">All Sources</option>
          <option value="ONLINE_BOOKING">Online Booking</option>
          <option value="WALK_IN">Walk-In</option>
          <option value="MANUAL">Manual Entry</option>
        </select>

        {/* Sort By */}
        <select
          value={sortBy}
          onChange={e => setSortBy(e.target.value as any)}
          className="px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
        >
          <option value="created_at">Sort: Newest First</option>
          <option value="name">Sort: Name (A-Z)</option>
          <option value="appointments">Sort: Most Appointments</option>
        </select>
      </div>

      {/* Main Content Area */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Customer List Column */}
        <div className={`space-y-3 ${selectedCustomer ? 'lg:col-span-1' : 'lg:col-span-3'}`}>
          {filteredCustomers.length === 0 ? (
            <div className="p-12 text-center rounded-2xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)]">
              <Users className="w-12 h-12 text-[var(--color-text-muted)] mx-auto mb-3 opacity-50" />
              <h3 className="text-sm font-semibold text-white">No customers found</h3>
              <p className="text-xs text-[var(--color-text-muted)] mt-1">
                Try refining your search query or reset filter settings.
              </p>
            </div>
          ) : (
            filteredCustomers.map(c => {
              const isSelected = selectedCustomer?.id === c.id;
              return (
                <div
                  key={c.id}
                  onClick={() => setSelectedCustomer(c)}
                  className={`p-4 rounded-xl border transition-all cursor-pointer ${
                    isSelected
                      ? 'bg-indigo-500/10 border-indigo-500/40 shadow-lg'
                      : 'bg-[var(--color-surface-raised)] border-[var(--color-surface-border)] hover:border-slate-700'
                  }`}
                >
                  <div className="flex items-start justify-between gap-3 mb-2">
                    <div className="flex items-center gap-3">
                      <div className="w-9 h-9 rounded-full bg-gradient-to-br from-indigo-500 to-violet-600 flex items-center justify-center text-white text-xs font-bold shadow-md">
                        {c.first_name[0]}
                        {c.last_name[0] || ''}
                      </div>
                      <div>
                        <h3 className="text-sm font-bold text-white font-[var(--font-display)]">
                          {c.first_name} {c.last_name}
                        </h3>
                        <p className="text-[11px] text-[var(--color-text-muted)]">
                          Source: <span className="text-indigo-400 font-medium">{c.source}</span>
                        </p>
                      </div>
                    </div>
                    {getStatusBadge(c.status || 'ACTIVE')}
                  </div>

                  <div className="grid grid-cols-2 gap-2 text-[11px] text-[var(--color-text-muted)] pt-3 border-t border-[var(--color-surface-border)]">
                    <div className="flex items-center gap-1.5 truncate">
                      <Mail className="w-3.5 h-3.5 text-indigo-400 flex-shrink-0" />
                      <span className="truncate">{c.email || 'No email'}</span>
                    </div>
                    <div className="flex items-center gap-1.5 truncate">
                      <Phone className="w-3.5 h-3.5 text-indigo-400 flex-shrink-0" />
                      <span className="truncate">{c.phone || 'No phone'}</span>
                    </div>
                    <div className="flex items-center gap-1.5">
                      <Calendar className="w-3.5 h-3.5 text-indigo-400 flex-shrink-0" />
                      <span>{c.total_appointments} appts</span>
                    </div>
                    <div className="flex items-center gap-1.5 font-medium text-emerald-400">
                      <DollarSign className="w-3.5 h-3.5 flex-shrink-0" />
                      <span>{formatMoney(c.total_spent_cents)}</span>
                    </div>
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Customer Detail Drawer Column */}
        {selectedCustomer && (
          <div className="lg:col-span-2 p-6 rounded-2xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] shadow-xl space-y-6 animate-fade-in">
            {/* Drawer Header */}
            <div className="flex items-start justify-between pb-4 border-b border-[var(--color-surface-border)]">
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 rounded-full bg-gradient-to-br from-indigo-500 to-violet-600 flex items-center justify-center text-white text-base font-bold shadow-lg">
                  {selectedCustomer.first_name[0]}
                  {selectedCustomer.last_name[0] || ''}
                </div>
                <div>
                  <div className="flex items-center gap-2">
                    <h2 className="text-lg font-bold text-white font-[var(--font-display)]">
                      {selectedCustomer.first_name} {selectedCustomer.last_name}
                    </h2>
                    {getStatusBadge(selectedCustomer.status || 'ACTIVE')}
                  </div>
                  <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                    Customer ID: <code className="text-indigo-300 font-mono text-[10px]">{selectedCustomer.id}</code>
                  </p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleOpenEdit(selectedCustomer)}
                  className="p-2 rounded-lg text-[var(--color-text-muted)] hover:text-white hover:bg-[var(--color-surface-hover)]"
                  title="Edit Customer"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => handleDeleteCustomer(selectedCustomer.id, `${selectedCustomer.first_name} ${selectedCustomer.last_name}`)}
                  className="p-2 rounded-lg text-[var(--color-text-muted)] hover:text-red-400 hover:bg-red-500/10"
                  title="Delete Customer"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setSelectedCustomer(null)}
                  className="p-2 rounded-lg text-[var(--color-text-muted)] hover:text-white"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* Tabs */}
            <div className="flex border-b border-[var(--color-surface-border)]">
              <button
                onClick={() => setDetailTab('profile')}
                className={`px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors flex items-center gap-2 ${
                  detailTab === 'profile'
                    ? 'border-indigo-500 text-indigo-400'
                    : 'border-transparent text-[var(--color-text-muted)] hover:text-white'
                }`}
              >
                <Users className="w-3.5 h-3.5" /> Profile Info
              </button>
              <button
                onClick={() => setDetailTab('notes')}
                className={`px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors flex items-center gap-2 ${
                  detailTab === 'notes'
                    ? 'border-indigo-500 text-indigo-400'
                    : 'border-transparent text-[var(--color-text-muted)] hover:text-white'
                }`}
              >
                <MessageSquare className="w-3.5 h-3.5" /> Notes & Activity ({selectedCustomer.notes_list?.length || 0})
              </button>
              <button
                onClick={() => setDetailTab('history')}
                className={`px-4 py-2.5 text-xs font-semibold border-b-2 transition-colors flex items-center gap-2 ${
                  detailTab === 'history'
                    ? 'border-indigo-500 text-indigo-400'
                    : 'border-transparent text-[var(--color-text-muted)] hover:text-white'
                }`}
              >
                <History className="w-3.5 h-3.5" /> Appointment History ({selectedCustomer.total_appointments})
              </button>
            </div>

            {/* Tab 1: Profile */}
            {detailTab === 'profile' && (
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div className="p-3.5 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
                    <p className="text-[10px] uppercase tracking-wider text-[var(--color-text-muted)] font-semibold mb-1">Email Address</p>
                    <p className="text-xs text-white font-medium flex items-center gap-1.5">
                      <Mail className="w-3.5 h-3.5 text-indigo-400" />
                      {selectedCustomer.email || 'Not provided'}
                    </p>
                  </div>
                  <div className="p-3.5 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)]">
                    <p className="text-[10px] uppercase tracking-wider text-[var(--color-text-muted)] font-semibold mb-1">Phone Number</p>
                    <p className="text-xs text-white font-medium flex items-center gap-1.5">
                      <Phone className="w-3.5 h-3.5 text-indigo-400" />
                      {selectedCustomer.phone || 'Not provided'}
                    </p>
                  </div>
                </div>

                <div className="p-4 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] space-y-2">
                  <h4 className="text-xs font-bold text-white">Identity Linking & Guest Status</h4>
                  <div className="flex items-center justify-between text-xs">
                    <span className="text-[var(--color-text-muted)]">Portal Application Account:</span>
                    {selectedCustomer.user_id ? (
                      <span className="px-2 py-0.5 rounded text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-medium">
                        Linked (User ID: {selectedCustomer.user_id.slice(0, 8)}...)
                      </span>
                    ) : (
                      <span className="px-2 py-0.5 rounded text-[10px] bg-slate-500/10 text-slate-400 border border-slate-500/20 font-medium">
                        Unlinked Guest Entity
                      </span>
                    )}
                  </div>
                  <p className="text-[11px] text-[var(--color-text-muted)]">
                    Customers remain lightweight CRM records until they explicitly register an account or accept identity linking.
                  </p>
                </div>

                {selectedCustomer.notes && (
                  <div className="p-4 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] space-y-1">
                    <h4 className="text-xs font-bold text-white">General CRM Summary Notes</h4>
                    <p className="text-xs text-[var(--color-text-muted)] italic">
                      "{selectedCustomer.notes}"
                    </p>
                  </div>
                )}
              </div>
            )}

            {/* Tab 2: Notes & Activity Timeline */}
            {detailTab === 'notes' && (
              <div className="space-y-4">
                {/* New Note Form */}
                <div className="p-3 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] space-y-2">
                  <textarea
                    rows={2}
                    value={newNoteText}
                    onChange={e => setNewNoteText(e.target.value)}
                    placeholder="Add a new internal activity note for this customer..."
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] text-white placeholder-[var(--color-text-muted)] focus:outline-none focus:border-indigo-500 resize-none"
                  />
                  <div className="flex justify-end">
                    <button
                      onClick={handleAddNoteToTimeline}
                      disabled={!newNoteText.trim()}
                      className="px-3 py-1.5 rounded-lg text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50"
                    >
                      Add Note
                    </button>
                  </div>
                </div>

                {/* Timeline */}
                <div className="space-y-3 pt-2">
                  {(!selectedCustomer.notes_list || selectedCustomer.notes_list.length === 0) ? (
                    <p className="text-xs text-[var(--color-text-muted)] text-center py-4">No activity notes recorded yet.</p>
                  ) : (
                    selectedCustomer.notes_list.map(n => (
                      <div key={n.id} className="p-3 rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] space-y-1">
                        <div className="flex items-center justify-between text-[11px]">
                          <span className="font-bold text-indigo-400">{n.author}</span>
                          <span className="text-[var(--color-text-muted)] flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            {new Date(n.created_at).toLocaleDateString()}
                          </span>
                        </div>
                        <p className="text-xs text-white leading-relaxed">{n.content}</p>
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}

            {/* Tab 3: History Placeholder */}
            {detailTab === 'history' && (
              <div className="p-6 text-center rounded-xl bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] space-y-3">
                <History className="w-10 h-10 text-indigo-400 mx-auto opacity-75" />
                <h4 className="text-sm font-bold text-white">Appointment Domain Integration</h4>
                <p className="text-xs text-[var(--color-text-muted)] max-w-sm mx-auto">
                  Historical and upcoming appointments for {selectedCustomer.first_name} will be automatically aggregated here once the Booking Engine phase is active.
                </p>
                <div className="pt-2">
                  <span className="inline-flex items-center gap-1 px-3 py-1 rounded-full text-xs font-medium bg-indigo-500/10 text-indigo-300 border border-indigo-500/20">
                    Total Spent: {formatMoney(selectedCustomer.total_spent_cents)}
                  </span>
                </div>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Add / Edit Customer Modal */}
      {isFormOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
          <div className="w-full max-w-md rounded-2xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] shadow-2xl p-6">
            <div className="flex items-center justify-between pb-4 mb-4 border-b border-[var(--color-surface-border)]">
              <h2 className="text-base font-bold text-white font-[var(--font-display)]">
                {editingCustomer ? 'Edit Customer Profile' : 'Add New Customer'}
              </h2>
              <button onClick={() => setIsFormOpen(false)} className="p-1 rounded-lg text-[var(--color-text-muted)] hover:text-white">
                <X className="w-4 h-4" />
              </button>
            </div>

            {formError && (
              <div className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-center gap-2">
                <AlertTriangle className="w-4 h-4 flex-shrink-0" />
                <span>{formError}</span>
              </div>
            )}

            <form onSubmit={handleSaveCustomer} className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">First Name *</label>
                  <input
                    type="text"
                    required
                    value={formFirstName}
                    onChange={e => setFormFirstName(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Last Name</label>
                  <input
                    type="text"
                    value={formLastName}
                    onChange={e => setFormLastName(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Email Address</label>
                <input
                  type="email"
                  value={formEmail}
                  onChange={e => setFormEmail(e.target.value)}
                  placeholder="e.g. customer@example.com"
                  className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div>
                <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Phone Number</label>
                <input
                  type="text"
                  value={formPhone}
                  onChange={e => setFormPhone(e.target.value)}
                  placeholder="e.g. +1 555-019-2831"
                  className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Status</label>
                  <select
                    value={formStatus}
                    onChange={e => setFormStatus(e.target.value as CustomerStatus)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  >
                    <option value="ACTIVE">Active</option>
                    <option value="INACTIVE">Inactive</option>
                    <option value="BLOCKED">Blocked</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Acquisition Source</label>
                  <select
                    value={formSource}
                    onChange={e => setFormSource(e.target.value as CustomerSource)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  >
                    <option value="MANUAL">Manual Entry</option>
                    <option value="ONLINE_BOOKING">Online Booking</option>
                    <option value="WALK_IN">Walk-In</option>
                    <option value="IMPORT">Import</option>
                    <option value="OTHER">Other</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">Notes</label>
                <textarea
                  rows={3}
                  value={formNotes}
                  onChange={e => setFormNotes(e.target.value)}
                  placeholder="Internal preferences, allergies, or notes..."
                  className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500 resize-none"
                />
              </div>

              <div className="flex items-center justify-end gap-3 pt-4 border-t border-[var(--color-surface-border)]">
                <button
                  type="button"
                  onClick={() => setIsFormOpen(false)}
                  className="px-4 py-2 rounded-lg text-xs text-[var(--color-text-muted)] hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 shadow-md shadow-indigo-600/30"
                >
                  Save Customer
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
