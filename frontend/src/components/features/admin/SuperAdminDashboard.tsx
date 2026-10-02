'use client';

import React, { useState, useEffect } from 'react';
import { 
  ShieldAlert, 
  Activity, 
  Building2, 
  Users, 
  CheckCircle2, 
  AlertTriangle, 
  Search, 
  Lock, 
  Unlock, 
  Sparkles, 
  RefreshCw, 
  Key,
  FileText,
  X,
  Eye,
  ShieldCheck,
  Server
} from 'lucide-react';

interface SystemHealth {
  platform_status: string;
  database_status: string;
  cache_status: string;
  active_tenants_count: number;
  suspended_tenants_count: number;
  total_bookings_month: number;
  failed_jobs_count: number;
  webhook_failures_count: number;
  uptime_seconds: number;
}

interface PlatformOrg {
  organization_id: string;
  name: string;
  slug: string;
  status: 'active' | 'suspended';
  plan_slug: string;
  staff_count: number;
  monthly_bookings: number;
  created_at: string;
}

export default function SuperAdminDashboard() {
  const [health, setHealth] = useState<SystemHealth | null>(null);
  const [organizations, setOrganizations] = useState<PlatformOrg[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [searchQuery, setSearchQuery] = useState<string>('');

  // Action Modals
  const [selectedOrgForSuspend, setSelectedOrgForSuspend] = useState<PlatformOrg | null>(null);
  const [suspendReason, setSuspendReason] = useState<string>('Violation of terms of service');
  const [isProcessingAction, setIsProcessingAction] = useState<boolean>(false);

  // Plan Assignment Modal
  const [selectedOrgForPlan, setSelectedOrgForPlan] = useState<PlatformOrg | null>(null);
  const [newPlanSlug, setNewPlanSlug] = useState<string>('professional');

  // Break-glass inspection modal
  const [isBreakGlassOpen, setIsBreakGlassOpen] = useState<boolean>(false);
  const [breakGlassOrgId, setBreakGlassOrgId] = useState<string>('');
  const [breakGlassCustId, setBreakGlassCustId] = useState<string>('');
  const [breakGlassReason, setBreakGlassReason] = useState<string>('Support Ticket #8492 - Customer data recovery');
  const [breakGlassData, setBreakGlassData] = useState<any | null>(null);
  const [breakGlassError, setBreakGlassError] = useState<string | null>(null);

  useEffect(() => {
    fetchAdminData();
  }, []);

  const fetchAdminData = async () => {
    setIsLoading(true);
    try {
      // 1. Health Summary
      const healthRes = await fetch('/api/v1/admin/health', {
        headers: { 'X-Super-Admin-Key': 'super-admin-secret-appointly' }
      });
      if (healthRes.ok) {
        const json = await healthRes.json();
        setHealth(json.data || null);
      }

      // 2. Organizations Directory
      const orgsRes = await fetch('/api/v1/admin/organizations', {
        headers: { 'X-Super-Admin-Key': 'super-admin-secret-appointly' }
      });
      if (orgsRes.ok) {
        const json = await orgsRes.json();
        if (json.data && Array.isArray(json.data.organizations)) {
          setOrganizations(json.data.organizations);
          return;
        }
      }
    } catch (err) {
      console.error('Failed to load super admin data:', err);
    } finally {
      setIsLoading(false);
    }

    // Default mock data if API unavailable
    setHealth({
      platform_status: 'healthy',
      database_status: 'healthy',
      cache_status: 'healthy',
      active_tenants_count: 12,
      suspended_tenants_count: 1,
      total_bookings_month: 1485,
      failed_jobs_count: 0,
      webhook_failures_count: 0,
      uptime_seconds: 345600
    });

    setOrganizations([
      {
        organization_id: 'org-1',
        name: 'Luxe Hair & Beauty Salon',
        slug: 'luxe-salon',
        status: 'active',
        plan_slug: 'growth',
        staff_count: 4,
        monthly_bookings: 84,
        created_at: '2026-08-01T10:00:00Z'
      },
      {
        organization_id: 'org-2',
        name: 'Apex Dental Practice',
        slug: 'apex-dental',
        status: 'active',
        plan_slug: 'professional',
        staff_count: 8,
        monthly_bookings: 210,
        created_at: '2026-08-15T14:30:00Z'
      },
      {
        organization_id: 'org-3',
        name: 'Spammy Test Tenant',
        slug: 'spammy-tenant',
        status: 'suspended',
        plan_slug: 'starter',
        staff_count: 1,
        monthly_bookings: 0,
        created_at: '2026-09-20T08:15:00Z'
      }
    ]);
  };

  const handleToggleSuspend = async () => {
    if (!selectedOrgForSuspend) return;
    setIsProcessingAction(true);

    const isSuspended = selectedOrgForSuspend.status === 'suspended';
    const endpoint = isSuspended
      ? `/api/v1/admin/organizations/${selectedOrgForSuspend.organization_id}/activate`
      : `/api/v1/admin/organizations/${selectedOrgForSuspend.organization_id}/suspend`;

    try {
      const res = await fetch(endpoint, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-Super-Admin-Key': 'super-admin-secret-appointly'
        },
        body: JSON.stringify({ reason: suspendReason })
      });

      if (res.ok) {
        setSelectedOrgForSuspend(null);
        fetchAdminData();
      }
    } catch (err) {
      console.error('Suspend/activate error:', err);
    } finally {
      setIsProcessingAction(false);
    }
  };

  const handleAssignPlan = async () => {
    if (!selectedOrgForPlan) return;
    setIsProcessingAction(true);

    try {
      const res = await fetch(`/api/v1/admin/organizations/${selectedOrgForPlan.organization_id}/assign-plan`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-Super-Admin-Key': 'super-admin-secret-appointly'
        },
        body: JSON.stringify({ plan_slug: newPlanSlug })
      });

      if (res.ok) {
        setSelectedOrgForPlan(null);
        fetchAdminData();
      }
    } catch (err) {
      console.error('Assign plan error:', err);
    } finally {
      setIsProcessingAction(false);
    }
  };

  const handleExecuteBreakGlass = async () => {
    setIsProcessingAction(true);
    setBreakGlassError(null);
    setBreakGlassData(null);

    try {
      const res = await fetch(`/api/v1/admin/organizations/${breakGlassOrgId}/customers/${breakGlassCustId}/breakglass`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-Super-Admin-Key': 'super-admin-secret-appointly'
        },
        body: JSON.stringify({ reason: breakGlassReason })
      });

      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.message || 'Break-glass access failed');
      }

      setBreakGlassData(json.data);
    } catch (err: any) {
      setBreakGlassError(err.message || 'Privileged inspection failed.');
    } finally {
      setIsProcessingAction(false);
    }
  };

  const filteredOrgs = organizations.filter(o => 
    o.name.toLowerCase().includes(searchQuery.toLowerCase()) || 
    o.slug.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <div className="space-y-8 max-w-7xl mx-auto">
      {/* Super Admin Distinction Banner */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-purple-950 via-slate-900 to-amber-950 border-2 border-amber-500/40 p-8 shadow-2xl">
        <div className="absolute top-0 right-0 bg-gradient-to-l from-amber-500 to-purple-600 text-slate-950 text-[10px] font-extrabold uppercase tracking-widest px-4 py-1.5 rounded-bl-xl shadow-lg">
          Platform Super Admin Control Plane
        </div>

        <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6 pt-2">
          <div>
            <div className="flex items-center gap-3">
              <ShieldAlert className="w-8 h-8 text-amber-400" />
              <h1 className="text-3xl font-black text-white tracking-tight">System & Platform Administration</h1>
            </div>
            <p className="text-amber-200/80 text-xs max-w-2xl mt-1.5 font-medium">
              Platform-wide operational monitoring, tenant organization suspension, SaaS plan overrides, and audited break-glass support operations. Isolated from standard organization tenancy.
            </p>
          </div>

          <div className="flex items-center gap-3 shrink-0">
            <button
              onClick={() => setIsBreakGlassOpen(true)}
              className="px-4 py-2.5 rounded-xl bg-amber-500/10 hover:bg-amber-500/20 text-amber-300 text-xs font-bold transition border border-amber-500/30 flex items-center gap-1.5"
            >
              <Key className="w-4 h-4" /> Break-Glass Support Access
            </button>
            <button
              onClick={fetchAdminData}
              className="px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold transition border border-slate-700 flex items-center gap-1.5"
            >
              <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} /> Refresh Metrics
            </button>
          </div>
        </div>
      </div>

      {/* Operational Monitoring Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-slate-900/80 backdrop-blur-xl border border-slate-800 rounded-2xl p-5 shadow-xl space-y-2">
          <div className="flex items-center justify-between text-xs text-slate-400">
            <span>Platform Status</span>
            <Server className="w-4 h-4 text-emerald-400" />
          </div>
          <div className="flex items-center gap-2">
            <span className="text-2xl font-bold text-white capitalize">{health?.platform_status || 'Healthy'}</span>
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse" />
          </div>
          <p className="text-[11px] text-slate-500">Database & Redis Cache Operational</p>
        </div>

        <div className="bg-slate-900/80 backdrop-blur-xl border border-slate-800 rounded-2xl p-5 shadow-xl space-y-2">
          <div className="flex items-center justify-between text-xs text-slate-400">
            <span>Active Tenant Orgs</span>
            <Building2 className="w-4 h-4 text-indigo-400" />
          </div>
          <span className="text-2xl font-bold text-white">{health?.active_tenants_count || 0}</span>
          <p className="text-[11px] text-slate-500">{health?.suspended_tenants_count || 0} suspended account(s)</p>
        </div>

        <div className="bg-slate-900/80 backdrop-blur-xl border border-slate-800 rounded-2xl p-5 shadow-xl space-y-2">
          <div className="flex items-center justify-between text-xs text-slate-400">
            <span>Total Bookings (Month)</span>
            <Activity className="w-4 h-4 text-purple-400" />
          </div>
          <span className="text-2xl font-bold text-white">{health?.total_bookings_month || 0}</span>
          <p className="text-[11px] text-slate-500">Across all active platform tenants</p>
        </div>

        <div className="bg-slate-900/80 backdrop-blur-xl border border-slate-800 rounded-2xl p-5 shadow-xl space-y-2">
          <div className="flex items-center justify-between text-xs text-slate-400">
            <span>Failed Jobs / Webhooks</span>
            <AlertTriangle className="w-4 h-4 text-amber-400" />
          </div>
          <span className="text-2xl font-bold text-white">{health?.failed_jobs_count || 0}</span>
          <p className="text-[11px] text-emerald-400 font-medium">0 delivery failures detected</p>
        </div>
      </div>

      {/* Platform Organization Management Directory */}
      <div className="bg-slate-900/80 backdrop-blur-xl border border-slate-800 rounded-2xl p-6 shadow-xl space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h2 className="text-xl font-bold text-white flex items-center gap-2">
              <Building2 className="w-5 h-5 text-indigo-400" /> Platform Tenant Directory
            </h2>
            <p className="text-xs text-slate-400">Manage tenant organization status and SaaS plan overrides.</p>
          </div>

          <div className="relative w-full sm:w-64">
            <Search className="w-4 h-4 text-slate-500 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              placeholder="Search tenant name or slug..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full bg-slate-800/80 border border-slate-700 rounded-xl pl-9 pr-3 py-1.5 text-xs text-white placeholder:text-slate-500 focus:outline-none focus:border-indigo-500"
            />
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-slate-800/50 text-slate-400 font-semibold border-b border-slate-800">
              <tr>
                <th className="p-3">Organization Name</th>
                <th className="p-3">Tenant Slug</th>
                <th className="p-3">SaaS Tier</th>
                <th className="p-3">Staff Seats</th>
                <th className="p-3">Monthly Bookings</th>
                <th className="p-3">Account Status</th>
                <th className="p-3 text-right">Super Admin Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/50 text-slate-300">
              {filteredOrgs.map((org) => (
                <tr key={org.organization_id} className="hover:bg-slate-800/30">
                  <td className="p-3 font-semibold text-white">{org.name}</td>
                  <td className="p-3 font-mono text-indigo-400">/{org.slug}</td>
                  <td className="p-3">
                    <span className="px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-purple-500/10 text-purple-400 border border-purple-500/20">
                      {org.plan_slug}
                    </span>
                  </td>
                  <td className="p-3 font-mono">{org.staff_count} staff</td>
                  <td className="p-3 font-mono">{org.monthly_bookings} bookings</td>
                  <td className="p-3">
                    {org.status === 'active' ? (
                      <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                        Active
                      </span>
                    ) : (
                      <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-rose-500/10 text-rose-400 border border-rose-500/20">
                        Suspended
                      </span>
                    )}
                  </td>
                  <td className="p-3 text-right space-x-2">
                    <button
                      onClick={() => {
                        setSelectedOrgForPlan(org);
                        setNewPlanSlug(org.plan_slug);
                      }}
                      className="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 font-medium text-[11px] border border-slate-700"
                    >
                      Assign Plan
                    </button>
                    <button
                      onClick={() => setSelectedOrgForSuspend(org)}
                      className={`px-2.5 py-1 rounded-lg font-medium text-[11px] transition ${
                        org.status === 'active'
                          ? 'bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 border border-rose-500/20'
                          : 'bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-400 border border-emerald-500/20'
                      }`}
                    >
                      {org.status === 'active' ? 'Suspend' : 'Activate'}
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Suspend / Activate Modal */}
      {selectedOrgForSuspend && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 backdrop-blur-sm p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-md w-full p-6 shadow-2xl space-y-5 relative">
            <div>
              <h3 className="text-lg font-bold text-white">
                {selectedOrgForSuspend.status === 'active' ? 'Suspend Tenant Organization?' : 'Activate Tenant Organization?'}
              </h3>
              <p className="text-xs text-slate-400 mt-1">
                Target Organization: <strong className="text-white">{selectedOrgForSuspend.name}</strong> ({selectedOrgForSuspend.slug})
              </p>
            </div>

            {selectedOrgForSuspend.status === 'active' && (
              <div className="space-y-2">
                <label className="text-xs font-medium text-slate-300 block">Reason for suspension:</label>
                <textarea
                  rows={3}
                  value={suspendReason}
                  onChange={(e) => setSuspendReason(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
                />
              </div>
            )}

            <div className="flex justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => setSelectedOrgForSuspend(null)}
                className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={isProcessingAction}
                onClick={handleToggleSuspend}
                className={`px-4 py-2 rounded-xl text-xs font-bold transition ${
                  selectedOrgForSuspend.status === 'active'
                    ? 'bg-rose-600 hover:bg-rose-700 text-white'
                    : 'bg-emerald-600 hover:bg-emerald-700 text-white'
                }`}
              >
                {isProcessingAction ? 'Processing...' : selectedOrgForSuspend.status === 'active' ? 'Confirm Suspension' : 'Confirm Activation'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Plan Assignment Modal */}
      {selectedOrgForPlan && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 backdrop-blur-sm p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-md w-full p-6 shadow-2xl space-y-5 relative">
            <div>
              <h3 className="text-lg font-bold text-white">Super Admin Plan Override</h3>
              <p className="text-xs text-slate-400 mt-1">
                Assign SaaS Tier for <strong className="text-white">{selectedOrgForPlan.name}</strong>
              </p>
            </div>

            <div className="space-y-2">
              <label className="text-xs font-medium text-slate-300 block">Select SaaS Plan:</label>
              <select
                value={newPlanSlug}
                onChange={(e) => setNewPlanSlug(e.target.value)}
                className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              >
                <option value="starter">Free Starter</option>
                <option value="growth">Growth ($29/mo)</option>
                <option value="professional">Professional ($79/mo)</option>
                <option value="enterprise">Enterprise ($199/mo)</option>
              </select>
            </div>

            <div className="flex justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => setSelectedOrgForPlan(null)}
                className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={isProcessingAction}
                onClick={handleAssignPlan}
                className="px-4 py-2 rounded-xl bg-purple-600 hover:bg-purple-700 text-white text-xs font-bold transition"
              >
                {isProcessingAction ? 'Processing...' : 'Apply Plan Override'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Break-Glass Customer Inspection Modal */}
      {isBreakGlassOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 backdrop-blur-sm p-4">
          <div className="bg-slate-900 border border-amber-500/40 rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-5 relative">
            <button
              onClick={() => setIsBreakGlassOpen(false)}
              className="absolute top-4 right-4 text-slate-400 hover:text-white p-1"
            >
              <X className="w-5 h-5" />
            </button>

            <div>
              <span className="text-[10px] font-bold uppercase tracking-wider text-amber-400 bg-amber-500/10 px-2.5 py-0.5 rounded-full border border-amber-500/20">
                Audited Break-Glass Protocol
              </span>
              <h3 className="text-lg font-bold text-white mt-1">Privileged Customer Support Inspection</h3>
              <p className="text-xs text-slate-400 mt-1">
                Super Admins do not possess implicit access to customer private data. Executing this request writes an immutable audit log entry.
              </p>
            </div>

            <div className="space-y-3">
              <div className="space-y-1">
                <label className="text-xs font-medium text-slate-300 block">Organization ID:</label>
                <input
                  type="text"
                  placeholder="org-uuid-v4"
                  value={breakGlassOrgId}
                  onChange={(e) => setBreakGlassOrgId(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-amber-500"
                />
              </div>
              <div className="space-y-1">
                <label className="text-xs font-medium text-slate-300 block">Customer ID:</label>
                <input
                  type="text"
                  placeholder="customer-uuid-v4"
                  value={breakGlassCustId}
                  onChange={(e) => setBreakGlassCustId(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-amber-500"
                />
              </div>
              <div className="space-y-1">
                <label className="text-xs font-medium text-slate-300 block">Explicit Support Justification Reason:</label>
                <textarea
                  rows={2}
                  value={breakGlassReason}
                  onChange={(e) => setBreakGlassReason(e.target.value)}
                  className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-amber-500"
                />
              </div>
            </div>

            {breakGlassError && (
              <div className="p-3 bg-rose-500/10 border border-rose-500/20 rounded-xl text-rose-400 text-xs">
                {breakGlassError}
              </div>
            )}

            {breakGlassData && (
              <div className="p-4 bg-slate-800/60 rounded-xl border border-slate-700 space-y-2 text-xs text-slate-300">
                <div className="font-bold text-amber-400 mb-1">Inspected Customer Record:</div>
                <div>Name: <span className="text-white font-semibold">{breakGlassData.first_name} {breakGlassData.last_name}</span></div>
                <div>Email: <span className="text-white font-semibold">{breakGlassData.email}</span></div>
                <div>Phone: <span className="text-white font-semibold">{breakGlassData.phone}</span></div>
              </div>
            )}

            <div className="flex justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => setIsBreakGlassOpen(false)}
                className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium"
              >
                Close
              </button>
              <button
                type="button"
                disabled={isProcessingAction}
                onClick={handleExecuteBreakGlass}
                className="px-5 py-2 rounded-xl bg-amber-600 hover:bg-amber-700 text-slate-950 font-extrabold text-xs transition"
              >
                {isProcessingAction ? 'Auditing...' : 'Execute Break-Glass Inspection'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
