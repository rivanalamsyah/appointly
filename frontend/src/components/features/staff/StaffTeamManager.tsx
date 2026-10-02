'use client';

import React, { useState } from 'react';
import { User, Plus, Mail, Phone, Shield, Scissors, Edit2, Trash2, CheckCircle2, Sparkles, UserPlus, Sliders } from 'lucide-react';
import type { Staff, Service } from '@/types/api';

interface StaffTeamManagerProps {
  initialStaff?: Partial<Staff>[];
  availableServices?: Partial<Service>[];
}

export default function StaffTeamManager({
  initialStaff = [
    {
      id: 'stf-1',
      organization_id: 'org-1',
      user_id: 'u-1',
      first_name: 'Alexander',
      last_name: 'Wright',
      email: 'alex@luxesalon.com',
      phone: '+1 (555) 234-5678',
      title: 'Master Hair Stylist',
      bio: 'Over 10 years of precision cutting experience in London and New York.',
      is_active: true,
      accepts_online: true,
    },
    {
      id: 'stf-2',
      organization_id: 'org-1',
      user_id: 'u-2',
      first_name: 'Elena',
      last_name: 'Rostova',
      email: 'elena@luxesalon.com',
      phone: '+1 (555) 876-5432',
      title: 'Color & Balayage Specialist',
      bio: 'Certified master colorist specializing in custom balayage highlights.',
      is_active: true,
      accepts_online: true,
    }
  ],
  availableServices = [
    { id: 'srv-1', name: 'Signature Haircut & Styling' },
    { id: 'srv-2', name: 'Full Balayage & Gloss Toner' },
    { id: 'srv-3', name: 'Hot Towel Beard Trim' },
  ]
}: StaffTeamManagerProps) {
  const [staffList, setStaffList] = useState(initialStaff);
  const [isStaffModalOpen, setIsStaffModalOpen] = useState(false);
  const [isServiceAssignmentOpen, setIsServiceAssignmentOpen] = useState(false);
  const [selectedStaff, setSelectedStaff] = useState<Partial<Staff> | null>(null);

  // Form Fields
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [title, setTitle] = useState('');
  const [bio, setBio] = useState('');
  const [acceptsOnline, setAcceptsOnline] = useState(true);

  // Service Assignment Checkbox State
  const [assignedServiceIds, setAssignedServiceIds] = useState<string[]>(['srv-1', 'srv-2']);
  const [toast, setToast] = useState<string | null>(null);

  const openCreateStaffModal = () => {
    setSelectedStaff(null);
    setFirstName('');
    setLastName('');
    setEmail('');
    setPhone('');
    setTitle('');
    setBio('');
    setAcceptsOnline(true);
    setIsStaffModalOpen(true);
  };

  const openEditStaffModal = (st: Partial<Staff>) => {
    setSelectedStaff(st);
    setFirstName(st.first_name || '');
    setLastName(st.last_name || '');
    setEmail(st.email || '');
    setPhone(st.phone || '');
    setTitle(st.title || '');
    setBio(st.bio || '');
    setAcceptsOnline(st.accepts_online ?? true);
    setIsStaffModalOpen(true);
  };

  const openServiceAssignModal = (st: Partial<Staff>) => {
    setSelectedStaff(st);
    setIsServiceAssignmentOpen(true);
  };

  const handleSaveStaff = (e: React.FormEvent) => {
    e.preventDefault();
    if (!firstName.trim() || !lastName.trim()) return;

    if (selectedStaff) {
      setStaffList((prev) =>
        prev.map((s) =>
          s.id === selectedStaff.id
            ? { ...s, first_name: firstName, last_name: lastName, email, phone, title, bio, accepts_online: acceptsOnline }
            : s
        )
      );
      setToast('Staff profile updated');
    } else {
      const newSt: Partial<Staff> = {
        id: `stf-${Date.now()}`,
        first_name: firstName,
        last_name: lastName,
        email,
        phone,
        title,
        bio,
        is_active: true,
        accepts_online: acceptsOnline,
      };
      setStaffList((prev) => [...prev, newSt]);
      setToast('New team member added');
    }
    setIsStaffModalOpen(false);
    setTimeout(() => setToast(null), 3000);
  };

  const handleToggleActive = (id: string) => {
    setStaffList((prev) =>
      prev.map((s) => (s.id === id ? { ...s, is_active: !s.is_active } : s))
    );
    setToast('Staff status updated');
    setTimeout(() => setToast(null), 3000);
  };

  const handleToggleServiceAssignment = (svcId: string) => {
    setAssignedServiceIds((prev) =>
      prev.includes(svcId) ? prev.filter((id) => id !== svcId) : [...prev, svcId]
    );
  };

  const handleSaveServiceAssignments = (e: React.FormEvent) => {
    e.preventDefault();
    setIsServiceAssignmentOpen(false);
    setToast('Service assignments saved for team member');
    setTimeout(() => setToast(null), 3000);
  };

  return (
    <div className="space-y-6 max-w-6xl">
      {toast && (
        <div className="p-4 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-semibold flex items-center justify-between shadow-lg animate-fade-in">
          <span>{toast}</span>
          <button onClick={() => setToast(null)} className="text-emerald-500 hover:text-emerald-300">Dismiss</button>
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-slate-900/80 border border-slate-800 rounded-3xl p-6 shadow-xl backdrop-blur-xl">
        <div>
          <h3 className="text-lg font-bold text-white tracking-tight flex items-center gap-2">
            <User className="w-5 h-5 text-violet-400" />
            Team & Staff Provider Management
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">Manage service providers, RBAC team roles, and service assignments.</p>
        </div>

        <button
          onClick={openCreateStaffModal}
          className="bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white text-xs font-bold px-5 py-2.5 rounded-xl transition-all shadow-lg shadow-violet-900/30 flex items-center gap-2 self-start md:self-auto"
        >
          <UserPlus className="w-4 h-4" />
          Add Staff Member
        </button>
      </div>

      {/* Staff Members Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {staffList.map((st) => (
          <div
            key={st.id}
            className={`bg-slate-900/80 border rounded-3xl p-6 space-y-4 transition-all shadow-xl backdrop-blur-xl flex flex-col justify-between ${
              st.is_active ? 'border-slate-800 hover:border-slate-700' : 'border-slate-800/40 opacity-60'
            }`}
          >
            <div className="space-y-4">
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-violet-600 to-indigo-600 flex items-center justify-center text-white font-bold text-lg shadow-md shrink-0">
                  {st.first_name?.charAt(0)}{st.last_name?.charAt(0)}
                </div>
                <div className="space-y-0.5">
                  <h4 className="font-bold text-white text-base leading-tight">
                    {st.first_name} {st.last_name}
                  </h4>
                  <p className="text-xs text-violet-400 font-medium">{st.title || 'Service Specialist'}</p>
                </div>
              </div>

              <p className="text-xs text-slate-400 leading-relaxed line-clamp-2">{st.bio}</p>

              {/* Assigned Services Pills */}
              <div className="space-y-1.5">
                <span className="text-[10px] text-slate-500 font-semibold uppercase tracking-wider">Assigned Services</span>
                <div className="flex flex-wrap gap-1.5">
                  {availableServices.slice(0, 2).map((svc) => (
                    <span key={svc.id} className="text-[10px] bg-violet-500/10 text-violet-300 border border-violet-500/20 px-2 py-0.5 rounded-full">
                      {svc.name}
                    </span>
                  ))}
                </div>
              </div>
            </div>

            <div className="pt-3 border-t border-slate-800/80 flex items-center justify-between">
              <button
                type="button"
                onClick={() => handleToggleActive(st.id!)}
                className={`px-3 py-1 rounded-full text-[10px] font-bold transition-all border ${
                  st.is_active
                    ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                    : 'bg-slate-800 text-slate-400 border-slate-700'
                }`}
              >
                {st.is_active ? 'Active' : 'Deactivated'}
              </button>

              <div className="flex items-center gap-2">
                <button
                  onClick={() => openServiceAssignModal(st)}
                  className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition-colors text-xs flex items-center gap-1"
                  title="Assign Services"
                >
                  <Scissors className="w-4 h-4 text-violet-400" />
                </button>
                <button
                  onClick={() => openEditStaffModal(st)}
                  className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition-colors"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Create / Edit Staff Modal */}
      {isStaffModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md animate-fade-in">
          <div className="bg-slate-900 border border-slate-800 rounded-3xl p-6 md:p-8 w-full max-w-lg shadow-2xl space-y-6">
            <div className="flex justify-between items-center pb-4 border-b border-slate-800">
              <h3 className="text-lg font-bold text-white tracking-tight">
                {selectedStaff ? 'Edit Staff Profile' : 'Add Staff Member'}
              </h3>
              <button onClick={() => setIsStaffModalOpen(false)} className="text-slate-400 hover:text-white text-xs">
                Close
              </button>
            </div>

            <form onSubmit={handleSaveStaff} className="space-y-4">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">First Name *</label>
                  <input
                    type="text"
                    required
                    placeholder="Alexander"
                    value={firstName}
                    onChange={(e) => setFirstName(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">Last Name *</label>
                  <input
                    type="text"
                    required
                    placeholder="Wright"
                    value={lastName}
                    onChange={(e) => setLastName(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Professional Title</label>
                <input
                  type="text"
                  placeholder="e.g. Senior Hair Stylist / Master Barber"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">Email Address</label>
                  <input
                    type="email"
                    placeholder="alex@business.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">Phone Number</label>
                  <input
                    type="tel"
                    placeholder="+1 (555) 000-0000"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Bio / Profile Description</label>
                <textarea
                  rows={3}
                  placeholder="Brief biography shown on customer booking page..."
                  value={bio}
                  onChange={(e) => setBio(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 resize-none"
                />
              </div>

              <div className="flex items-center justify-between p-4 bg-slate-950/60 border border-slate-800 rounded-2xl">
                <div>
                  <h4 className="text-xs font-semibold text-white">Accept Online Bookings</h4>
                  <p className="text-[11px] text-slate-400">Allow customers to choose this staff member online</p>
                </div>
                <input
                  type="checkbox"
                  checked={acceptsOnline}
                  onChange={(e) => setAcceptsOnline(e.target.checked)}
                  className="w-5 h-5 accent-violet-600 rounded cursor-pointer"
                />
              </div>

              <div className="flex justify-end gap-3 pt-4 border-t border-slate-800">
                <button
                  type="button"
                  onClick={() => setIsStaffModalOpen(false)}
                  className="px-4 py-2.5 text-xs font-semibold text-slate-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="bg-gradient-to-r from-violet-600 to-indigo-600 text-white text-xs font-bold px-6 py-2.5 rounded-xl shadow-lg"
                >
                  Save Staff Profile
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Many-to-Many Service Assignment Modal */}
      {isServiceAssignmentOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md animate-fade-in">
          <div className="bg-slate-900 border border-slate-800 rounded-3xl p-6 w-full max-w-md shadow-2xl space-y-4">
            <h3 className="text-base font-bold text-white flex items-center gap-2">
              <Scissors className="w-5 h-5 text-violet-400" />
              Assign Services to {selectedStaff?.first_name}
            </h3>
            <p className="text-xs text-slate-400">Select which catalog services this staff member is qualified to deliver.</p>

            <form onSubmit={handleSaveServiceAssignments} className="space-y-3 pt-2">
              {availableServices.map((svc) => (
                <label
                  key={svc.id}
                  className="flex items-center justify-between p-3.5 bg-slate-950/60 border border-slate-800 rounded-2xl cursor-pointer hover:border-slate-700 transition-colors"
                >
                  <span className="text-xs font-semibold text-white">{svc.name}</span>
                  <input
                    type="checkbox"
                    checked={assignedServiceIds.includes(svc.id!)}
                    onChange={() => handleToggleServiceAssignment(svc.id!)}
                    className="w-5 h-5 accent-violet-600 rounded cursor-pointer"
                  />
                </label>
              ))}

              <div className="flex justify-end gap-3 pt-4 border-t border-slate-800">
                <button
                  type="button"
                  onClick={() => setIsServiceAssignmentOpen(false)}
                  className="px-4 py-2 text-xs font-semibold text-slate-400"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="bg-violet-600 text-white text-xs font-bold px-5 py-2 rounded-xl shadow-lg"
                >
                  Save Service Assignments
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
