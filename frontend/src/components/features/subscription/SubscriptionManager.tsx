'use client';

import React, { useState, useEffect } from 'react';
import { 
  Sparkles, 
  Check, 
  X, 
  ShieldCheck, 
  Zap, 
  AlertCircle, 
  CreditCard, 
  ArrowRight, 
  Users, 
  MapPin, 
  Calendar, 
  FileText, 
  Clock, 
  CheckCircle2, 
  HelpCircle,
  Download,
  TrendingUp,
  RefreshCw
} from 'lucide-react';

interface Plan {
  id: string;
  name: string;
  code: string;
  description: string;
  price_monthly: number;
  price_yearly: number;
  max_staff: number;
  max_locations: number;
  monthly_booking_limit: number;
  features: string[];
}

interface Subscription {
  id: string;
  organization_id: string;
  plan_id: string;
  plan?: Plan;
  status: 'trialing' | 'active' | 'past_due' | 'cancelled' | 'expired';
  billing_period: 'monthly' | 'yearly';
  amount: number;
  currency: string;
  current_period_start: string;
  current_period_end: string;
  cancel_at_period_end: boolean;
  trial_ends_at?: string;
  canceled_at?: string;
}

interface UsageMetrics {
  current_staff_count: number;
  max_staff_allowed: number;
  current_location_count: number;
  max_locations_allowed: number;
  current_monthly_bookings: number;
  max_monthly_bookings: number;
  billing_cycle_month: string;
}

export default function SubscriptionManager() {
  const [plans, setPlans] = useState<Plan[]>([]);
  const [subscription, setSubscription] = useState<Subscription | null>(null);
  const [usage, setUsage] = useState<UsageMetrics | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [billingCycle, setBillingCycle] = useState<'monthly' | 'yearly'>('monthly');

  // Checkout modal
  const [selectedPlanForCheckout, setSelectedPlanForCheckout] = useState<Plan | null>(null);
  const [isCheckoutModalOpen, setIsCheckoutModalOpen] = useState<boolean>(false);
  const [paymentProvider, setPaymentProvider] = useState<string>('stripe');
  const [isProcessingCheckout, setIsProcessingCheckout] = useState<boolean>(false);
  const [checkoutSuccess, setCheckoutSuccess] = useState<string | null>(null);
  const [checkoutError, setCheckoutError] = useState<string | null>(null);

  // Cancellation modal
  const [isCancelModalOpen, setIsCancelModalOpen] = useState<boolean>(false);
  const [cancelReason, setCancelReason] = useState<string>('too_expensive');
  const [isProcessingCancel, setIsProcessingCancel] = useState<boolean>(false);

  useEffect(() => {
    fetchSubscriptionData();
  }, []);

  const fetchSubscriptionData = async () => {
    setIsLoading(true);
    try {
      // 1. Fetch Plans
      const plansRes = await fetch('/api/v1/plans');
      if (plansRes.ok) {
        const json = await plansRes.json();
        if (json.data && Array.isArray(json.data.plans)) {
          setPlans(json.data.plans);
        }
      }

      // 2. Fetch Current Subscription & Usage
      const subRes = await fetch('/api/v1/subscriptions/current');
      if (subRes.ok) {
        const json = await subRes.json();
        if (json.data) {
          setSubscription(json.data.subscription || null);
          setUsage(json.data.usage || null);
          if (json.data.subscription?.plan) {
            setBillingCycle(json.data.subscription.billing_period || 'monthly');
          }
        }
      }
    } catch (err) {
      console.error('Failed to load subscription info:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const handleOpenCheckout = (plan: Plan) => {
    setSelectedPlanForCheckout(plan);
    setCheckoutError(null);
    setCheckoutSuccess(null);
    setIsCheckoutModalOpen(true);
  };

  const handleExecuteCheckout = async () => {
    if (!selectedPlanForCheckout) return;
    setIsProcessingCheckout(true);
    setCheckoutError(null);

    try {
      const res = await fetch('/api/v1/subscriptions/checkout', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          plan_id: selectedPlanForCheckout.id,
          billing_period: billingCycle,
          provider: paymentProvider,
        })
      });

      const json = await res.json();
      if (!res.ok) {
        throw new Error(json.message || 'Checkout failed');
      }

      setCheckoutSuccess(`Successfully subscribed to ${selectedPlanForCheckout.name} plan!`);
      setTimeout(() => {
        setIsCheckoutModalOpen(false);
        fetchSubscriptionData();
      }, 1500);
    } catch (err: any) {
      setCheckoutError(err.message || 'An error occurred during checkout.');
    } finally {
      setIsProcessingCheckout(false);
    }
  };

  const handleCancelSubscription = async () => {
    setIsProcessingCancel(true);
    try {
      const res = await fetch('/api/v1/subscriptions/cancel', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ reason: cancelReason })
      });

      if (res.ok) {
        setIsCancelModalOpen(false);
        fetchSubscriptionData();
      }
    } catch (err) {
      console.error('Cancellation error:', err);
    } finally {
      setIsProcessingCancel(false);
    }
  };

  // UI state calculation helpers
  const currentPlanCode = subscription?.plan?.code || 'starter';

  const getStatusBadge = (status?: string) => {
    switch (status) {
      case 'active':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"><CheckCircle2 className="w-3 h-3 mr-1" /> Active</span>;
      case 'trialing':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-amber-500/10 text-amber-400 border border-amber-500/20"><Zap className="w-3 h-3 mr-1" /> Free Trial</span>;
      case 'past_due':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/20"><AlertCircle className="w-3 h-3 mr-1" /> Past Due</span>;
      case 'cancelled':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-500/10 text-slate-400 border border-slate-500/20"><X className="w-3 h-3 mr-1" /> Cancelled</span>;
      case 'expired':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-800 text-slate-400 border border-slate-700"><Clock className="w-3 h-3 mr-1" /> Expired</span>;
      default:
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20"><Sparkles className="w-3 h-3 mr-1" /> Free Starter</span>;
    }
  };

  const calculateUsagePercent = (current: number, max: number) => {
    if (max <= 0) return 0; // Unlimited or 0
    return Math.min(Math.round((current / max) * 100), 100);
  };

  return (
    <div className="space-y-8 max-w-7xl mx-auto">
      {/* Header Banner */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 border border-slate-800 p-8 shadow-2xl">
        <div className="absolute -right-12 -top-12 w-64 h-64 bg-indigo-500/10 rounded-full blur-3xl pointer-events-none" />
        <div className="absolute right-32 -bottom-12 w-64 h-64 bg-purple-500/10 rounded-full blur-3xl pointer-events-none" />

        <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div>
            <div className="flex items-center gap-3 mb-2">
              <h1 className="text-3xl font-extrabold text-white tracking-tight">SaaS Subscription & Plans</h1>
              {getStatusBadge(subscription?.status)}
            </div>
            <p className="text-slate-400 text-sm max-w-2xl">
              Manage your Appointly tenant subscription level, feature access limits, seat quotas, and SaaS billing invoices. Decoupled from client appointment transactions.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={fetchSubscriptionData}
              className="px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-sm font-medium transition flex items-center gap-2 border border-slate-700"
            >
              <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} /> Refresh Status
            </button>
            {subscription && subscription.status !== 'cancelled' && currentPlanCode !== 'starter' && (
              <button
                onClick={() => setIsCancelModalOpen(true)}
                className="px-4 py-2.5 rounded-xl bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 text-sm font-medium transition border border-rose-500/20"
              >
                Cancel Subscription
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Current Active Plan Overview & Usage Gauges */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Active Plan Card */}
        <div className="lg:col-span-1 bg-slate-900/80 backdrop-blur-xl border border-slate-800 rounded-2xl p-6 shadow-xl flex flex-col justify-between">
          <div>
            <div className="flex items-center justify-between mb-4">
              <span className="text-xs uppercase font-bold tracking-wider text-indigo-400 bg-indigo-500/10 px-3 py-1 rounded-full border border-indigo-500/20">
                Current Tier
              </span>
              <ShieldCheck className="w-5 h-5 text-indigo-400" />
            </div>

            <h2 className="text-2xl font-bold text-white mb-1">
              {subscription?.plan?.name || 'Free Starter Plan'}
            </h2>
            <p className="text-slate-400 text-xs mb-6">
              {subscription?.plan?.description || 'Essential appointment booking for small business owners and freelancers.'}
            </p>

            <div className="space-y-4 py-4 border-t border-b border-slate-800/80 mb-6">
              <div className="flex justify-between items-center text-sm">
                <span className="text-slate-400">Monthly Price:</span>
                <span className="text-white font-semibold">
                  {subscription?.plan ? `$${subscription.plan.price_monthly}/mo` : '$0/mo'}
                </span>
              </div>
              <div className="flex justify-between items-center text-sm">
                <span className="text-slate-400">Billing Cycle:</span>
                <span className="text-white font-semibold capitalize">{subscription?.billing_period || 'Monthly'}</span>
              </div>
              <div className="flex justify-between items-center text-sm">
                <span className="text-slate-400">Current Renewal Period:</span>
                <span className="text-slate-300 font-medium">
                  {subscription?.current_period_end 
                    ? new Date(subscription.current_period_end).toLocaleDateString() 
                    : 'Unlimited'}
                </span>
              </div>
              {subscription?.cancel_at_period_end && (
                <div className="p-3 bg-amber-500/10 border border-amber-500/20 rounded-xl text-xs text-amber-400">
                  <AlertCircle className="w-4 h-4 inline mr-1" />
                  Scheduled for cancellation at period end.
                </div>
              )}
            </div>
          </div>

          <div className="text-xs text-slate-500 flex items-center gap-1.5">
            <CreditCard className="w-4 h-4 text-slate-400" />
            <span>SaaS Provider Payment Method: Card ending in ****4242</span>
          </div>
        </div>

        {/* Real-time Usage Metrics & Limit Enforcement */}
        <div className="lg:col-span-2 bg-slate-900/80 backdrop-blur-xl border border-slate-800 rounded-2xl p-6 shadow-xl space-y-6">
          <div className="flex items-center justify-between border-b border-slate-800/80 pb-4">
            <div>
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <TrendingUp className="w-5 h-5 text-indigo-400" /> Resource Quota & Usage Limits
              </h3>
              <p className="text-xs text-slate-400">Real-time enforcement at application layer</p>
            </div>
            <span className="text-xs font-mono bg-slate-800 text-slate-300 px-3 py-1 rounded-lg border border-slate-700">
              Cycle: {usage?.billing_cycle_month || 'Current Month'}
            </span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Staff Seats Usage */}
            <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-800 space-y-3">
              <div className="flex items-center justify-between text-xs text-slate-400">
                <span className="flex items-center gap-1.5 font-medium text-slate-300">
                  <Users className="w-4 h-4 text-indigo-400" /> Staff Quota
                </span>
                <span className="font-mono text-white">
                  {usage?.current_staff_count || 0} / {usage?.max_staff_allowed || 2}
                </span>
              </div>
              <div className="w-full bg-slate-700/50 h-2 rounded-full overflow-hidden">
                <div 
                  className={`h-full transition-all duration-500 rounded-full ${
                    calculateUsagePercent(usage?.current_staff_count || 0, usage?.max_staff_allowed || 2) >= 100 
                      ? 'bg-rose-500' 
                      : calculateUsagePercent(usage?.current_staff_count || 0, usage?.max_staff_allowed || 2) > 75 
                      ? 'bg-amber-400' 
                      : 'bg-indigo-500'
                  }`} 
                  style={{ width: `${calculateUsagePercent(usage?.current_staff_count || 0, usage?.max_staff_allowed || 2)}%` }} 
                />
              </div>
              <p className="text-[11px] text-slate-400">
                {calculateUsagePercent(usage?.current_staff_count || 0, usage?.max_staff_allowed || 2) >= 100 ? (
                  <span className="text-rose-400 font-medium">Limit reached. Upgrade for more staff seats.</span>
                ) : (
                  `${(usage?.max_staff_allowed || 2) - (usage?.current_staff_count || 0)} staff seat(s) available.`
                )}
              </p>
            </div>

            {/* Locations Usage */}
            <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-800 space-y-3">
              <div className="flex items-center justify-between text-xs text-slate-400">
                <span className="flex items-center gap-1.5 font-medium text-slate-300">
                  <MapPin className="w-4 h-4 text-emerald-400" /> Locations Quota
                </span>
                <span className="font-mono text-white">
                  {usage?.current_location_count || 0} / {usage?.max_locations_allowed || 1}
                </span>
              </div>
              <div className="w-full bg-slate-700/50 h-2 rounded-full overflow-hidden">
                <div 
                  className="h-full transition-all duration-500 rounded-full bg-emerald-500" 
                  style={{ width: `${calculateUsagePercent(usage?.current_location_count || 0, usage?.max_locations_allowed || 1)}%` }} 
                />
              </div>
              <p className="text-[11px] text-slate-400">
                Max locations permitted by plan limits.
              </p>
            </div>

            {/* Monthly Bookings Usage */}
            <div className="p-4 rounded-xl bg-slate-800/50 border border-slate-800 space-y-3">
              <div className="flex items-center justify-between text-xs text-slate-400">
                <span className="flex items-center gap-1.5 font-medium text-slate-300">
                  <Calendar className="w-4 h-4 text-purple-400" /> Monthly Bookings
                </span>
                <span className="font-mono text-white">
                  {usage?.current_monthly_bookings || 0} / {usage?.max_monthly_bookings || 100}
                </span>
              </div>
              <div className="w-full bg-slate-700/50 h-2 rounded-full overflow-hidden">
                <div 
                  className="h-full transition-all duration-500 rounded-full bg-purple-500" 
                  style={{ width: `${calculateUsagePercent(usage?.current_monthly_bookings || 0, usage?.max_monthly_bookings || 100)}%` }} 
                />
              </div>
              <p className="text-[11px] text-slate-400">
                Resets every 1st of the month.
              </p>
            </div>
          </div>

          {/* Active Features Checklist */}
          <div className="pt-2">
            <span className="text-xs font-semibold text-slate-400 uppercase tracking-wider block mb-3">
              Features Enabled in Your Active Tier:
            </span>
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
              {subscription?.plan?.features?.map((feat, idx) => (
                <div key={idx} className="flex items-center gap-2 text-xs text-slate-300 bg-slate-800/30 px-3 py-2 rounded-lg border border-slate-800">
                  <Check className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                  <span className="capitalize">{feat.replace(/_/g, ' ')}</span>
                </div>
              )) || (
                <>
                  <div className="flex items-center gap-2 text-xs text-slate-300 bg-slate-800/30 px-3 py-2 rounded-lg border border-slate-800">
                    <Check className="w-3.5 h-3.5 text-emerald-400 shrink-0" /> Up to 2 Staff Seats
                  </div>
                  <div className="flex items-center gap-2 text-xs text-slate-300 bg-slate-800/30 px-3 py-2 rounded-lg border border-slate-800">
                    <Check className="w-3.5 h-3.5 text-emerald-400 shrink-0" /> 1 Business Location
                  </div>
                  <div className="flex items-center gap-2 text-xs text-slate-300 bg-slate-800/30 px-3 py-2 rounded-lg border border-slate-800">
                    <Check className="w-3.5 h-3.5 text-emerald-400 shrink-0" /> 100 Monthly Bookings
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Plan Selection & Comparison Table */}
      <div className="space-y-6 pt-6">
        <div className="text-center space-y-2 max-w-2xl mx-auto">
          <h2 className="text-2xl font-bold text-white">Choose the Right SaaS Plan for Your Business</h2>
          <p className="text-slate-400 text-sm">
            Scale seamlessly as your appointment volume and staff count grow. Upgrade or downgrade anytime.
          </p>

          {/* Billing Cycle Toggle */}
          <div className="pt-4 flex items-center justify-center gap-4">
            <span className={`text-xs font-semibold ${billingCycle === 'monthly' ? 'text-white' : 'text-slate-400'}`}>
              Monthly Billing
            </span>
            <button
              onClick={() => setBillingCycle(prev => prev === 'monthly' ? 'yearly' : 'monthly')}
              className="w-12 h-6 rounded-full bg-indigo-600/40 p-1 border border-indigo-500/30 transition duration-300 flex items-center"
            >
              <div className={`w-4 h-4 rounded-full bg-indigo-400 transition-transform duration-300 ${billingCycle === 'yearly' ? 'translate-x-6' : ''}`} />
            </button>
            <span className={`text-xs font-semibold flex items-center gap-1 ${billingCycle === 'yearly' ? 'text-white' : 'text-slate-400'}`}>
              Yearly Billing <span className="bg-emerald-500/10 text-emerald-400 text-[10px] px-2 py-0.5 rounded-full border border-emerald-500/20 font-bold">Save 20%</span>
            </span>
          </div>
        </div>

        {/* Plans Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {plans.map((plan) => {
            const isCurrent = currentPlanCode === plan.code;
            const price = billingCycle === 'yearly' ? Math.round((plan.price_yearly / 12)) : plan.price_monthly;

            return (
              <div 
                key={plan.id}
                className={`relative rounded-2xl bg-slate-900/90 border ${
                  isCurrent 
                    ? 'border-indigo-500 ring-2 ring-indigo-500/20 shadow-indigo-500/10' 
                    : plan.code === 'professional'
                    ? 'border-purple-500/50'
                    : 'border-slate-800 hover:border-slate-700'
                } p-6 shadow-xl flex flex-col justify-between transition-all duration-300`}
              >
                {plan.code === 'professional' && (
                  <div className="absolute -top-3 left-1/2 -translate-x-1/2 bg-gradient-to-r from-indigo-500 to-purple-500 text-white text-[10px] font-bold uppercase tracking-widest px-3 py-1 rounded-full shadow-lg">
                    Most Popular
                  </div>
                )}

                <div>
                  <div className="flex justify-between items-start mb-4">
                    <div>
                      <h3 className="text-xl font-bold text-white">{plan.name}</h3>
                      <p className="text-slate-400 text-xs mt-1 min-h-[36px]">{plan.description}</p>
                    </div>
                  </div>

                  <div className="my-6">
                    <span className="text-4xl font-extrabold text-white">${price}</span>
                    <span className="text-slate-400 text-xs font-medium">/month</span>
                    {billingCycle === 'yearly' && plan.price_yearly > 0 && (
                      <p className="text-[11px] text-slate-500 mt-1">Billed annually (${plan.price_yearly}/yr)</p>
                    )}
                  </div>

                  <div className="space-y-3 border-t border-slate-800/80 pt-4 mb-6 text-xs text-slate-300">
                    <div className="flex items-center gap-2">
                      <Users className="w-4 h-4 text-indigo-400 shrink-0" />
                      <span>{plan.max_staff >= 999 ? 'Unlimited Staff' : `Up to ${plan.max_staff} Staff Seats`}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <MapPin className="w-4 h-4 text-emerald-400 shrink-0" />
                      <span>{plan.max_locations >= 999 ? 'Unlimited Locations' : `Up to ${plan.max_locations} Location(s)`}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <Calendar className="w-4 h-4 text-purple-400 shrink-0" />
                      <span>{plan.monthly_booking_limit >= 99999 ? 'Unlimited Monthly Bookings' : `${plan.monthly_booking_limit} Monthly Bookings`}</span>
                    </div>

                    <div className="pt-2 border-t border-slate-800/50 space-y-2">
                      {plan.features?.map((feat, idx) => (
                        <div key={idx} className="flex items-center gap-2 text-slate-400">
                          <Check className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                          <span className="capitalize">{feat.replace(/_/g, ' ')}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>

                <div>
                  {isCurrent ? (
                    <button 
                      disabled 
                      className="w-full py-2.5 rounded-xl bg-slate-800 text-slate-400 text-xs font-bold border border-slate-700 cursor-default"
                    >
                      Current Plan
                    </button>
                  ) : (
                    <button
                      onClick={() => handleOpenCheckout(plan)}
                      className={`w-full py-2.5 rounded-xl text-xs font-bold transition flex items-center justify-center gap-1.5 shadow-lg ${
                        plan.code === 'professional' || plan.code === 'enterprise'
                          ? 'bg-gradient-to-r from-indigo-500 to-purple-600 hover:from-indigo-600 hover:to-purple-700 text-white'
                          : 'bg-indigo-600 hover:bg-indigo-700 text-white'
                      }`}
                    >
                      Select {plan.name} <ArrowRight className="w-3.5 h-3.5" />
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* SaaS Billing History & Invoices */}
      <div className="bg-slate-900/80 backdrop-blur-xl border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-white flex items-center gap-2">
            <FileText className="w-5 h-5 text-indigo-400" /> SaaS Subscription Billing History
          </h3>
          <span className="text-xs text-slate-400">Decoupled from Customer Appointment Payments</span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead className="bg-slate-800/50 text-slate-400 font-semibold border-b border-slate-800">
              <tr>
                <th className="p-3">Invoice ID</th>
                <th className="p-3">Date</th>
                <th className="p-3">Plan / Description</th>
                <th className="p-3">Amount</th>
                <th className="p-3">Provider</th>
                <th className="p-3">Status</th>
                <th className="p-3 text-right">Invoice PDF</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/50 text-slate-300">
              <tr>
                <td className="p-3 font-mono">INV-SAAS-2026-001</td>
                <td className="p-3">Oct 01, 2026</td>
                <td className="p-3 font-medium text-white">Appointly Growth Plan (Monthly)</td>
                <td className="p-3 font-semibold text-white">$29.00</td>
                <td className="p-3 capitalize">Stripe SaaS</td>
                <td className="p-3">
                  <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    Paid
                  </span>
                </td>
                <td className="p-3 text-right">
                  <button className="text-indigo-400 hover:text-indigo-300 font-medium inline-flex items-center gap-1">
                    <Download className="w-3.5 h-3.5" /> Download
                  </button>
                </td>
              </tr>
              <tr>
                <td className="p-3 font-mono text-slate-500">INV-SAAS-2026-000</td>
                <td className="p-3 text-slate-500">Sep 01, 2026</td>
                <td className="p-3 font-medium text-slate-400">Free Starter Trial</td>
                <td className="p-3 font-semibold text-slate-400">$0.00</td>
                <td className="p-3 capitalize text-slate-500">System</td>
                <td className="p-3">
                  <span className="px-2 py-0.5 rounded-full text-[10px] font-bold bg-slate-800 text-slate-400 border border-slate-700">
                    Completed
                  </span>
                </td>
                <td className="p-3 text-right">
                  <span className="text-slate-600 text-[11px]">N/A</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      {/* Checkout Modal */}
      {isCheckoutModalOpen && selectedPlanForCheckout && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 backdrop-blur-sm p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-md w-full p-6 shadow-2xl space-y-6 relative animate-in fade-in zoom-in-95">
            <button
              onClick={() => setIsCheckoutModalOpen(false)}
              className="absolute top-4 right-4 text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800"
            >
              <X className="w-5 h-5" />
            </button>

            <div>
              <span className="text-xs font-semibold text-indigo-400 uppercase tracking-wider">SaaS Checkout</span>
              <h3 className="text-xl font-bold text-white mt-1">Upgrade to {selectedPlanForCheckout.name}</h3>
              <p className="text-xs text-slate-400 mt-1">
                You are subscribing to Appointly SaaS tier for your organization.
              </p>
            </div>

            {checkoutSuccess ? (
              <div className="p-4 bg-emerald-500/10 border border-emerald-500/20 rounded-xl text-emerald-400 text-xs font-medium text-center space-y-2">
                <CheckCircle2 className="w-8 h-8 text-emerald-400 mx-auto" />
                <p>{checkoutSuccess}</p>
              </div>
            ) : (
              <>
                <div className="bg-slate-800/50 p-4 rounded-xl border border-slate-800 space-y-2">
                  <div className="flex justify-between text-xs">
                    <span className="text-slate-400">Selected Plan:</span>
                    <span className="text-white font-semibold">{selectedPlanForCheckout.name}</span>
                  </div>
                  <div className="flex justify-between text-xs">
                    <span className="text-slate-400">Billing Interval:</span>
                    <span className="text-white font-semibold capitalize">{billingCycle}</span>
                  </div>
                  <div className="flex justify-between text-xs border-t border-slate-700/50 pt-2 font-bold">
                    <span className="text-slate-300">Total Due Today:</span>
                    <span className="text-indigo-400">
                      ${billingCycle === 'yearly' ? selectedPlanForCheckout.price_yearly : selectedPlanForCheckout.price_monthly}
                    </span>
                  </div>
                </div>

                <div className="space-y-2">
                  <label className="text-xs font-medium text-slate-300 block">SaaS Billing Provider:</label>
                  <div className="grid grid-cols-2 gap-2">
                    <button
                      type="button"
                      onClick={() => setPaymentProvider('stripe')}
                      className={`p-3 rounded-xl border text-xs font-semibold flex items-center justify-center gap-2 transition ${
                        paymentProvider === 'stripe'
                          ? 'border-indigo-500 bg-indigo-500/10 text-white'
                          : 'border-slate-800 bg-slate-800/30 text-slate-400 hover:bg-slate-800'
                      }`}
                    >
                      <CreditCard className="w-4 h-4" /> Credit Card (Stripe)
                    </button>
                    <button
                      type="button"
                      onClick={() => setPaymentProvider('midtrans')}
                      className={`p-3 rounded-xl border text-xs font-semibold flex items-center justify-center gap-2 transition ${
                        paymentProvider === 'midtrans'
                          ? 'border-indigo-500 bg-indigo-500/10 text-white'
                          : 'border-slate-800 bg-slate-800/30 text-slate-400 hover:bg-slate-800'
                      }`}
                    >
                      <Zap className="w-4 h-4" /> Midtrans Gateway
                    </button>
                  </div>
                </div>

                {checkoutError && (
                  <div className="p-3 bg-rose-500/10 border border-rose-500/20 rounded-xl text-rose-400 text-xs">
                    {checkoutError}
                  </div>
                )}

                <div className="flex justify-end gap-3 pt-2">
                  <button
                    type="button"
                    onClick={() => setIsCheckoutModalOpen(false)}
                    className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium"
                  >
                    Cancel
                  </button>
                  <button
                    type="button"
                    disabled={isProcessingCheckout}
                    onClick={handleExecuteCheckout}
                    className="px-5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold transition flex items-center gap-2"
                  >
                    {isProcessingCheckout ? (
                      <>
                        <RefreshCw className="w-4 h-4 animate-spin" /> Processing...
                      </>
                    ) : (
                      <>
                        Confirm & Pay ${billingCycle === 'yearly' ? selectedPlanForCheckout.price_yearly : selectedPlanForCheckout.price_monthly}
                      </>
                    )}
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}

      {/* Cancellation Modal */}
      {isCancelModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 backdrop-blur-sm p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-md w-full p-6 shadow-2xl space-y-5 relative">
            <div>
              <h3 className="text-lg font-bold text-white">Cancel Subscription?</h3>
              <p className="text-xs text-slate-400 mt-1">
                Your subscription will remain active until the end of your current billing period. Afterwards, your account will revert to the Free Starter tier limits.
              </p>
            </div>

            <div className="space-y-2">
              <label className="text-xs font-medium text-slate-300 block">Reason for cancellation:</label>
              <select
                value={cancelReason}
                onChange={(e) => setCancelReason(e.target.value)}
                className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
              >
                <option value="too_expensive">Too expensive / Budget constraint</option>
                <option value="missing_features">Missing required features</option>
                <option value="switching_service">Switching to another platform</option>
                <option value="temporary">Temporary pause</option>
                <option value="other">Other reason</option>
              </select>
            </div>

            <div className="flex justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => setIsCancelModalOpen(false)}
                className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium"
              >
                Keep Subscription
              </button>
              <button
                type="button"
                disabled={isProcessingCancel}
                onClick={handleCancelSubscription}
                className="px-4 py-2 rounded-xl bg-rose-600 hover:bg-rose-700 text-white text-xs font-bold transition"
              >
                {isProcessingCancel ? 'Processing...' : 'Confirm Cancellation'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
