'use client';

import React, { useState } from 'react';
import { Building2, Globe, Clock, DollarSign, Save, CheckCircle2, AlertCircle, Sparkles, Sliders } from 'lucide-react';
import type { Organization } from '@/types/api';

interface BusinessSettingsFormProps {
  initialOrg?: Partial<Organization>;
}

export default function BusinessSettingsForm({
  initialOrg = {
    id: 'org-1',
    name: 'Luxe Hair Salon & Spa',
    slug: 'luxe-salon',
    business_type: 'salon',
    email: 'contact@luxesalon.com',
    phone: '+1 (555) 234-5678',
    website: 'https://luxesalon.com',
    description: 'Premier luxury salon offering haircut, coloring, balayage, and spa treatments.',
    timezone: 'America/New_York',
    currency: 'USD',
    country: 'US',
    status: 'active',
  }
}: BusinessSettingsFormProps) {
  const [name, setName] = useState(initialOrg.name || '');
  const [slug, setSlug] = useState(initialOrg.slug || '');
  const [email, setEmail] = useState(initialOrg.email || '');
  const [phone, setPhone] = useState(initialOrg.phone || '');
  const [website, setWebsite] = useState(initialOrg.website || '');
  const [description, setDescription] = useState(initialOrg.description || '');
  const [timezone, setTimezone] = useState(initialOrg.timezone || 'UTC');
  const [currency, setCurrency] = useState(initialOrg.currency || 'USD');

  // Booking settings
  const [allowOnlineBooking, setAllowOnlineBooking] = useState(true);
  const [allowGuestBooking, setAllowGuestBooking] = useState(true);
  const [autoConfirmBookings, setAutoConfirmBookings] = useState(true);

  const [isSaving, setIsSaving] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    setToastMessage(null);

    // Simulate API call
    setTimeout(() => {
      setIsSaving(false);
      setToastMessage('Business profile and booking settings saved successfully!');
      setTimeout(() => setToastMessage(null), 4000);
    }, 800);
  };

  return (
    <form onSubmit={handleSave} className="space-y-8 max-w-4xl">
      {toastMessage && (
        <div className="p-4 rounded-2xl bg-emerald-500/10 border border-emerald-500/30 text-emerald-400 text-xs font-semibold flex items-center justify-between shadow-lg animate-fade-in">
          <div className="flex items-center gap-2">
            <CheckCircle2 className="w-5 h-5 text-emerald-400 shrink-0" />
            <span>{toastMessage}</span>
          </div>
          <button
            type="button"
            onClick={() => setToastMessage(null)}
            className="text-emerald-500 hover:text-emerald-300"
          >
            Dismiss
          </button>
        </div>
      )}

      {/* Profile Section */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-3xl p-6 md:p-8 space-y-6 shadow-xl backdrop-blur-xl">
        <div className="flex items-center justify-between pb-4 border-b border-slate-800">
          <div>
            <h3 className="text-lg font-bold text-white tracking-tight flex items-center gap-2">
              <Building2 className="w-5 h-5 text-violet-400" />
              Business Profile
            </h3>
            <p className="text-xs text-slate-400 mt-0.5">Manage your public brand identity and contact information.</p>
          </div>
          <span className="text-[11px] font-mono bg-violet-500/10 text-violet-400 border border-violet-500/20 px-3 py-1 rounded-full">
            Public Slug: /book/{slug}
          </span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-1.5">Business Name *</label>
            <input
              type="text"
              required
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-1.5">Public Booking Slug *</label>
            <div className="flex items-center">
              <span className="bg-slate-950 border border-r-0 border-slate-800 rounded-l-xl px-3 py-3 text-slate-500 text-xs font-mono">
                /book/
              </span>
              <input
                type="text"
                required
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-r-xl px-4 py-3 text-slate-200 text-xs font-mono focus:outline-none focus:border-violet-500 transition-colors"
              />
            </div>
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-1.5">Public Email Address *</label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-1.5">Phone Number</label>
            <input
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors"
            />
          </div>

          <div className="md:col-span-2">
            <label className="block text-xs font-semibold text-slate-300 mb-1.5">Website URL</label>
            <input
              type="url"
              value={website}
              onChange={(e) => setWebsite(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors"
            />
          </div>

          <div className="md:col-span-2">
            <label className="block text-xs font-semibold text-slate-300 mb-1.5">Business Description</label>
            <textarea
              rows={3}
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors resize-none"
            />
          </div>
        </div>
      </div>

      {/* Regional & Timezone Settings */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-3xl p-6 md:p-8 space-y-6 shadow-xl backdrop-blur-xl">
        <div className="pb-4 border-b border-slate-800">
          <h3 className="text-lg font-bold text-white tracking-tight flex items-center gap-2">
            <Globe className="w-5 h-5 text-indigo-400" />
            Regional & Localization Settings
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">Controls time slot calculations and currency formatting.</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-1.5">Primary IANA Timezone *</label>
            <select
              value={timezone}
              onChange={(e) => setTimezone(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors font-mono cursor-pointer"
            >
              <option value="America/New_York">America/New_York (Eastern Time)</option>
              <option value="Asia/Jakarta">Asia/Jakarta (WIB)</option>
              <option value="Europe/London">Europe/London (GMT)</option>
              <option value="UTC">UTC (Universal)</option>
            </select>
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-1.5">Currency (ISO 4217) *</label>
            <select
              value={currency}
              onChange={(e) => setCurrency(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors font-mono cursor-pointer"
            >
              <option value="USD">USD ($)</option>
              <option value="IDR">IDR (Rp)</option>
              <option value="EUR">EUR (€)</option>
              <option value="GBP">GBP (£)</option>
            </select>
          </div>
        </div>
      </div>

      {/* Booking Rules Toggle */}
      <div className="bg-slate-900/80 border border-slate-800 rounded-3xl p-6 md:p-8 space-y-6 shadow-xl backdrop-blur-xl">
        <div className="pb-4 border-b border-slate-800">
          <h3 className="text-lg font-bold text-white tracking-tight flex items-center gap-2">
            <Sliders className="w-5 h-5 text-cyan-400" />
            Booking Engine Preferences
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">Control customer online booking behavior.</p>
        </div>

        <div className="space-y-4">
          <div className="flex items-center justify-between p-4 bg-slate-950/60 border border-slate-800 rounded-2xl">
            <div>
              <h4 className="text-xs font-semibold text-white">Enable Public Online Booking</h4>
              <p className="text-[11px] text-slate-400">Allow customers to book appointments directly via public link</p>
            </div>
            <input
              type="checkbox"
              checked={allowOnlineBooking}
              onChange={(e) => setAllowOnlineBooking(e.target.checked)}
              className="w-5 h-5 accent-violet-600 rounded cursor-pointer"
            />
          </div>

          <div className="flex items-center justify-between p-4 bg-slate-950/60 border border-slate-800 rounded-2xl">
            <div>
              <h4 className="text-xs font-semibold text-white">Allow Guest Bookings</h4>
              <p className="text-[11px] text-slate-400">Customers can book without creating an account</p>
            </div>
            <input
              type="checkbox"
              checked={allowGuestBooking}
              onChange={(e) => setAllowGuestBooking(e.target.checked)}
              className="w-5 h-5 accent-violet-600 rounded cursor-pointer"
            />
          </div>

          <div className="flex items-center justify-between p-4 bg-slate-950/60 border border-slate-800 rounded-2xl">
            <div>
              <h4 className="text-xs font-semibold text-white">Auto-Confirm Bookings</h4>
              <p className="text-[11px] text-slate-400">Automatically set new bookings to Confirmed status</p>
            </div>
            <input
              type="checkbox"
              checked={autoConfirmBookings}
              onChange={(e) => setAutoConfirmBookings(e.target.checked)}
              className="w-5 h-5 accent-violet-600 rounded cursor-pointer"
            />
          </div>
        </div>
      </div>

      {/* Save Button */}
      <div className="flex justify-end pt-2">
        <button
          type="submit"
          disabled={isSaving}
          className="bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white text-xs font-bold px-8 py-3.5 rounded-xl transition-all shadow-lg shadow-violet-900/40 disabled:opacity-50 flex items-center gap-2"
        >
          {isSaving ? (
            <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
          ) : (
            <>
              <Save className="w-4 h-4" />
              Save Profile Settings
            </>
          )}
        </button>
      </div>
    </form>
  );
}
