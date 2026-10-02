'use client';

import React, { useState } from 'react';
import { MapPin, Plus, Phone, Mail, Globe, Check, Trash2, Edit2, ShieldAlert, Star, Building2 } from 'lucide-react';
import type { Location } from '@/types/api';

interface LocationsManagerProps {
  initialLocations?: Partial<Location>[];
}

export default function LocationsManager({
  initialLocations = [
    {
      id: 'loc-1',
      organization_id: 'org-1',
      name: 'Downtown Main Branch',
      phone: '+1 (555) 123-4567',
      email: 'downtown@luxesalon.com',
      address_line1: '123 Beauty Blvd, Suite 100',
      city: 'New York',
      state: 'NY',
      postal_code: '10001',
      country: 'US',
      timezone: 'America/New_York',
      is_active: true,
    },
    {
      id: 'loc-2',
      organization_id: 'org-1',
      name: 'Westside Boutique Salon',
      phone: '+1 (555) 987-6543',
      email: 'westside@luxesalon.com',
      address_line1: '456 Fashion Ave',
      city: 'New York',
      state: 'NY',
      postal_code: '10018',
      country: 'US',
      timezone: 'America/New_York',
      is_active: true,
    }
  ]
}: LocationsManagerProps) {
  const [locations, setLocations] = useState(initialLocations);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingLocation, setEditingLocation] = useState<Partial<Location> | null>(null);

  // Form State
  const [name, setName] = useState('');
  const [phone, setPhone] = useState('');
  const [email, setEmail] = useState('');
  const [addressLine1, setAddressLine1] = useState('');
  const [city, setCity] = useState('');
  const [timezone, setTimezone] = useState('America/New_York');

  // Delete Confirmation State
  const [deletingLocationId, setDeletingLocationId] = useState<string | null>(null);
  const [toast, setToast] = useState<string | null>(null);

  const openCreateModal = () => {
    setEditingLocation(null);
    setName('');
    setPhone('');
    setEmail('');
    setAddressLine1('');
    setCity('');
    setTimezone('America/New_York');
    setIsModalOpen(true);
  };

  const openEditModal = (loc: Partial<Location>) => {
    setEditingLocation(loc);
    setName(loc.name || '');
    setPhone(loc.phone || '');
    setEmail(loc.email || '');
    setAddressLine1(loc.address_line1 || '');
    setCity(loc.city || '');
    setTimezone(loc.timezone || 'America/New_York');
    setIsModalOpen(true);
  };

  const handleSaveLocation = (e: React.FormEvent) => {
    e.preventDefault();
    if (editingLocation) {
      // Update existing location
      setLocations((prev) =>
        prev.map((l) => (l.id === editingLocation.id ? { ...l, name, phone, email, address_line1: addressLine1, city, timezone } : l))
      );
      setToast('Location updated successfully');
    } else {
      // Create new location
      const newLoc: Partial<Location> = {
        id: `loc-${Date.now()}`,
        name,
        phone,
        email,
        address_line1: addressLine1,
        city,
        timezone,
        is_active: true,
      };
      setLocations((prev) => [...prev, newLoc]);
      setToast('New location created successfully');
    }
    setIsModalOpen(false);
    setTimeout(() => setToast(null), 3000);
  };

  const handleDeleteLocation = (id: string) => {
    setLocations((prev) => prev.filter((l) => l.id !== id));
    setDeletingLocationId(null);
    setToast('Location removed');
    setTimeout(() => setToast(null), 3000);
  };

  return (
    <div className="space-y-6 max-w-5xl">
      {toast && (
        <div className="p-4 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-semibold flex items-center justify-between shadow-lg animate-fade-in">
          <span>{toast}</span>
          <button onClick={() => setToast(null)} className="text-emerald-500 hover:text-emerald-300">Dismiss</button>
        </div>
      )}

      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-slate-900/80 border border-slate-800 rounded-3xl p-6 shadow-xl backdrop-blur-xl">
        <div>
          <h3 className="text-lg font-bold text-white tracking-tight flex items-center gap-2">
            <MapPin className="w-5 h-5 text-violet-400" />
            Multi-Branch Locations
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">Manage physical salon branches, clinic locations, and operating timezones.</p>
        </div>

        <button
          onClick={openCreateModal}
          className="bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white text-xs font-bold px-5 py-2.5 rounded-xl transition-all shadow-lg shadow-violet-900/30 flex items-center gap-2 self-start sm:self-auto"
        >
          <Plus className="w-4 h-4" />
          Add New Location
        </button>
      </div>

      {/* Locations Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {locations.map((loc, idx) => (
          <div
            key={loc.id}
            className="bg-slate-900/80 border border-slate-800 rounded-3xl p-6 space-y-4 hover:border-slate-700 transition-all shadow-xl backdrop-blur-xl relative group"
          >
            <div className="flex items-start justify-between gap-4">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <h4 className="font-bold text-white text-base">{loc.name}</h4>
                  {idx === 0 && (
                    <span className="inline-flex items-center gap-1 bg-amber-500/10 text-amber-400 border border-amber-500/30 text-[10px] font-semibold px-2.5 py-0.5 rounded-full">
                      <Star className="w-3 h-3 fill-amber-400" /> Primary Branch
                    </span>
                  )}
                </div>
                <p className="text-xs text-slate-400">{loc.address_line1}, {loc.city}</p>
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={() => openEditModal(loc)}
                  className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition-colors"
                >
                  <Edit2 className="w-4 h-4" />
                </button>
                {idx !== 0 && (
                  <button
                    onClick={() => setDeletingLocationId(loc.id!)}
                    className="p-2 text-rose-400 hover:text-rose-300 hover:bg-rose-500/10 rounded-xl transition-colors"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                )}
              </div>
            </div>

            <div className="pt-3 border-t border-slate-800/80 grid grid-cols-2 gap-3 text-xs text-slate-400">
              <div className="flex items-center gap-2">
                <Phone className="w-3.5 h-3.5 text-violet-400" />
                <span className="truncate">{loc.phone || 'No phone'}</span>
              </div>
              <div className="flex items-center gap-2 font-mono">
                <Globe className="w-3.5 h-3.5 text-indigo-400" />
                <span className="truncate">{loc.timezone}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Create / Edit Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md animate-fade-in">
          <div className="bg-slate-900 border border-slate-800 rounded-3xl p-6 md:p-8 w-full max-w-lg shadow-2xl space-y-6">
            <div className="flex justify-between items-center pb-4 border-b border-slate-800">
              <h3 className="text-lg font-bold text-white tracking-tight">
                {editingLocation ? 'Edit Branch Location' : 'Add New Branch Location'}
              </h3>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-slate-400 hover:text-white text-xs font-semibold px-2 py-1 rounded"
              >
                Close
              </button>
            </div>

            <form onSubmit={handleSaveLocation} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Branch Name *</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Westside Boutique Branch"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Street Address *</label>
                <input
                  type="text"
                  required
                  placeholder="123 Main Street, Suite 200"
                  value={addressLine1}
                  onChange={(e) => setAddressLine1(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">City *</label>
                  <input
                    type="text"
                    required
                    placeholder="New York"
                    value={city}
                    onChange={(e) => setCity(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">IANA Timezone *</label>
                  <select
                    value={timezone}
                    onChange={(e) => setTimezone(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-3 text-slate-200 text-xs font-mono focus:outline-none focus:border-violet-500 cursor-pointer"
                  >
                    <option value="America/New_York">America/New_York</option>
                    <option value="Asia/Jakarta">Asia/Jakarta</option>
                    <option value="Europe/London">Europe/London</option>
                  </select>
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">Phone</label>
                  <input
                    type="tel"
                    placeholder="+1 (555) 000-0000"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">Email</label>
                  <input
                    type="email"
                    placeholder="branch@luxesalon.com"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                  />
                </div>
              </div>

              <div className="flex justify-end gap-3 pt-4 border-t border-slate-800">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2.5 text-xs font-semibold text-slate-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="bg-gradient-to-r from-violet-600 to-indigo-600 text-white text-xs font-bold px-6 py-2.5 rounded-xl shadow-lg"
                >
                  Save Location
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Destructive Delete Dialog */}
      {deletingLocationId && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md animate-fade-in">
          <div className="bg-slate-900 border border-slate-800 rounded-3xl p-6 w-full max-w-sm shadow-2xl space-y-4 text-center">
            <div className="w-12 h-12 bg-rose-500/10 border border-rose-500/20 rounded-full flex items-center justify-center text-rose-400 mx-auto">
              <ShieldAlert className="w-6 h-6" />
            </div>
            <h4 className="text-base font-bold text-white">Delete Branch Location?</h4>
            <p className="text-xs text-slate-400">This action will remove the location from your workspace. This action cannot be undone.</p>
            <div className="flex items-center justify-center gap-3 pt-2">
              <button
                onClick={() => setDeletingLocationId(null)}
                className="px-4 py-2 text-xs font-semibold text-slate-400 hover:text-white"
              >
                Cancel
              </button>
              <button
                onClick={() => handleDeleteLocation(deletingLocationId)}
                className="bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold px-5 py-2 rounded-xl shadow-lg shadow-rose-900/30"
              >
                Delete Location
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
