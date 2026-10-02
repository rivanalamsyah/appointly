'use client';

import React, { useState } from 'react';
import { Building2, Globe, Sparkles, ArrowRight, AlertCircle } from 'lucide-react';

export default function OnboardingForm() {
  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [businessType, setBusinessType] = useState('salon');
  const [email, setEmail] = useState('');
  const [timezone, setTimezone] = useState('UTC');
  const [currency, setCurrency] = useState('USD');
  const [isLoading, setIsLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState('');

  const handleNameChange = (val: string) => {
    setName(val);
    const autoSlug = val.toLowerCase().replace(/[^a-z0-9]/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '');
    setSlug(autoSlug);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setErrorMessage('');

    const token = localStorage.getItem('appointly_auth_token');
    if (!token) {
      window.location.href = '/auth/login';
      return;
    }

    try {
      const response = await fetch('/api/v1/organizations', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          name,
          slug,
          business_type: businessType,
          email,
          timezone,
          currency,
        }),
      });

      const data = await response.json();

      if (!response.ok || !data.data) {
        throw new Error(data.error?.message || 'Failed to create organization');
      }

      // Store selected tenant slug
      if (data.data.organization?.slug) {
        localStorage.setItem('appointly_tenant_slug', data.data.organization.slug);
      }

      // Redirect to main admin dashboard
      window.location.href = '/dashboard';
    } catch (err: any) {
      setErrorMessage(err.message || 'Failed to create organization');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="w-full max-w-lg mx-auto space-y-6">
      <div className="text-center space-y-2">
        <div className="w-12 h-12 rounded-2xl bg-gradient-to-tr from-violet-600 via-indigo-600 to-cyan-500 flex items-center justify-center text-white font-extrabold mx-auto shadow-lg shadow-violet-900/40">
          <Building2 className="w-6 h-6" />
        </div>
        <h2 className="text-2xl font-extrabold text-white tracking-tight">Set up your Organization</h2>
        <p className="text-xs text-slate-400">Configure your business workspace & tenant portal</p>
      </div>

      <div className="bg-slate-900/80 border border-slate-800 rounded-3xl p-6 md:p-8 shadow-2xl backdrop-blur-xl">
        {errorMessage && (
          <div className="mb-4 p-3 rounded-xl bg-rose-500/10 border border-rose-500/20 text-rose-400 text-xs flex items-center gap-2">
            <AlertCircle className="w-4 h-4 shrink-0" />
            <span>{errorMessage}</span>
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-1.5">Business Name *</label>
            <input
              type="text"
              required
              placeholder="e.g. Luxe Hair Salon & Spa"
              value={name}
              onChange={(e) => handleNameChange(e.target.value)}
              className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors"
            />
          </div>

          <div>
            <label className="block text-xs font-semibold text-slate-300 mb-1.5">Booking URL Slug *</label>
            <div className="flex items-center">
              <span className="bg-slate-950 border border-r-0 border-slate-800 rounded-l-xl px-3 py-3 text-slate-500 text-xs font-mono">
                appointly.app/book/
              </span>
              <input
                type="text"
                required
                placeholder="luxe-salon"
                value={slug}
                onChange={(e) => setSlug(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-r-xl px-4 py-3 text-slate-200 text-xs font-mono focus:outline-none focus:border-violet-500 transition-colors"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1.5">Industry / Vertical</label>
              <select
                value={businessType}
                onChange={(e) => setBusinessType(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors cursor-pointer"
              >
                <option value="salon">Salon / Barbershop</option>
                <option value="clinic">Healthcare / Clinic</option>
                <option value="consultant">Consulting</option>
                <option value="law_firm">Legal Services</option>
                <option value="tutor">Education / Tutoring</option>
                <option value="gym">Fitness / Gym</option>
                <option value="other">Other Service Business</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1.5">Business Email</label>
              <input
                type="email"
                required
                placeholder="info@business.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1.5">Timezone</label>
              <select
                value={timezone}
                onChange={(e) => setTimezone(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors cursor-pointer font-mono"
              >
                <option value="UTC">UTC</option>
                <option value="Asia/Jakarta">Asia/Jakarta (WIB)</option>
                <option value="America/New_York">America/New_York (EST)</option>
                <option value="Europe/London">Europe/London (GMT)</option>
              </select>
            </div>
            <div>
              <label className="block text-xs font-semibold text-slate-300 mb-1.5">Currency</label>
              <select
                value={currency}
                onChange={(e) => setCurrency(e.target.value)}
                className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 transition-colors cursor-pointer font-mono"
              >
                <option value="USD">USD ($)</option>
                <option value="IDR">IDR (Rp)</option>
                <option value="EUR">EUR (€)</option>
                <option value="GBP">GBP (£)</option>
              </select>
            </div>
          </div>

          <button
            type="submit"
            disabled={isLoading}
            className="w-full bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white font-bold py-3.5 rounded-xl text-xs transition-all shadow-lg shadow-violet-900/40 disabled:opacity-50 flex items-center justify-center gap-2 mt-4"
          >
            {isLoading ? (
              <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
            ) : (
              <>
                Create Workspace & Launch Dashboard <ArrowRight className="w-4 h-4" />
              </>
            )}
          </button>
        </form>
      </div>
    </div>
  );
}
