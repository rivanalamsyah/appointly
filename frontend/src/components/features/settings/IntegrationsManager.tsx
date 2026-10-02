'use client';

import React, { useState, useEffect } from 'react';
import { 
  Calendar, 
  MessageSquare, 
  Mail, 
  Webhook, 
  CheckCircle2, 
  XCircle, 
  AlertCircle, 
  RefreshCw, 
  ExternalLink, 
  ShieldCheck, 
  Settings, 
  Link2, 
  Unlink,
  X,
  Copy,
  Check
} from 'lucide-react';

interface Integration {
  id: string;
  organization_id: string;
  provider: 'google_calendar' | 'outlook_calendar' | 'whatsapp_twilio' | 'email_resend' | 'outbound_webhook';
  category: 'calendar' | 'messaging' | 'email' | 'webhook';
  status: 'connected' | 'disconnected' | 'error' | 'syncing';
  account_email?: string;
  webhook_url?: string;
  webhook_secret?: string;
  last_synced_at?: string;
  error_message?: string;
}

export default function IntegrationsManager() {
  const [integrations, setIntegrations] = useState<Integration[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [activeModal, setActiveModal] = useState<Integration | null>(null);

  // Form states for modals
  const [accountEmail, setAccountEmail] = useState<string>('');
  const [webhookUrl, setWebhookUrl] = useState<string>('');
  const [webhookSecret, setWebhookSecret] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [copiedSecret, setCopiedSecret] = useState<boolean>(false);

  useEffect(() => {
    fetchIntegrations();
  }, []);

  const fetchIntegrations = async () => {
    setIsLoading(true);
    try {
      const res = await fetch('/api/v1/integrations');
      if (res.ok) {
        const json = await res.json();
        if (json.data && Array.isArray(json.data.integrations)) {
          setIntegrations(json.data.integrations);
          return;
        }
      }
    } catch (err) {
      console.error('Failed to load integrations:', err);
    } finally {
      setIsLoading(false);
    }

    // Default mock initial list if none returned
    setIntegrations([
      {
        id: 'int-1',
        organization_id: 'org-1',
        provider: 'google_calendar',
        category: 'calendar',
        status: 'connected',
        account_email: 'owner@salon.com',
        last_synced_at: new Date().toISOString()
      },
      {
        id: 'int-2',
        organization_id: 'org-1',
        provider: 'whatsapp_twilio',
        category: 'messaging',
        status: 'disconnected'
      },
      {
        id: 'int-3',
        organization_id: 'org-1',
        provider: 'email_resend',
        category: 'email',
        status: 'connected',
        account_email: 'notifications@appointly.app'
      },
      {
        id: 'int-4',
        organization_id: 'org-1',
        provider: 'outbound_webhook',
        category: 'webhook',
        status: 'disconnected'
      }
    ]);
  };

  const handleOpenConnectModal = (item: Integration) => {
    setActiveModal(item);
    setAccountEmail(item.account_email || '');
    setWebhookUrl(item.webhook_url || '');
    setWebhookSecret(item.webhook_secret || 'whsec_' + Math.random().toString(36).substring(2, 12));
  };

  const handleConnect = async () => {
    if (!activeModal) return;
    setIsSubmitting(true);

    try {
      const res = await fetch(`/api/v1/integrations/${activeModal.provider}/connect`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          account_email: accountEmail,
          webhook_url: webhookUrl,
          webhook_secret: webhookSecret,
        })
      });

      if (res.ok) {
        setActiveModal(null);
        fetchIntegrations();
      }
    } catch (err) {
      console.error('Connect error:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDisconnect = async (provider: string) => {
    setIsSubmitting(true);
    try {
      const res = await fetch(`/api/v1/integrations/${provider}/disconnect`, {
        method: 'POST'
      });

      if (res.ok) {
        setActiveModal(null);
        fetchIntegrations();
      }
    } catch (err) {
      console.error('Disconnect error:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleReconnect = async (provider: string) => {
    setIsLoading(true);
    try {
      await fetch(`/api/v1/integrations/${provider}/reconnect`, {
        method: 'POST'
      });
      fetchIntegrations();
    } catch (err) {
      console.error('Reconnect error:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'connected':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"><CheckCircle2 className="w-3 h-3 mr-1" /> Connected</span>;
      case 'syncing':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20"><RefreshCw className="w-3 h-3 mr-1 animate-spin" /> Syncing</span>;
      case 'error':
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/20"><AlertCircle className="w-3 h-3 mr-1" /> Auth Error</span>;
      default:
        return <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-semibold bg-slate-800 text-slate-400 border border-slate-700"><XCircle className="w-3 h-3 mr-1" /> Disconnected</span>;
    }
  };

  const getProviderInfo = (provider: string) => {
    switch (provider) {
      case 'google_calendar':
        return {
          title: 'Google Calendar',
          description: 'Synchronize tenant appointments bi-directionally with Google Calendar.',
          icon: <Calendar className="w-6 h-6 text-indigo-400" />
        };
      case 'whatsapp_twilio':
        return {
          title: 'WhatsApp Business (Twilio)',
          description: 'Automate WhatsApp booking confirmations and automated appointment reminders.',
          icon: <MessageSquare className="w-6 h-6 text-emerald-400" />
        };
      case 'email_resend':
        return {
          title: 'Email Notifications (Resend / SendGrid)',
          description: 'Send branded HTML booking receipts, cancellation notices, and staff invites.',
          icon: <Mail className="w-6 h-6 text-purple-400" />
        };
      case 'outbound_webhook':
        return {
          title: 'Outbound Webhooks',
          description: 'Deliver real-time HMAC signed JSON events to custom tenant HTTP endpoints.',
          icon: <Webhook className="w-6 h-6 text-amber-400" />
        };
      default:
        return {
          title: provider,
          description: 'External third-party integration connector.',
          icon: <Link2 className="w-6 h-6 text-slate-400" />
        };
    }
  };

  return (
    <div className="space-y-8 max-w-7xl mx-auto">
      {/* Header Banner */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-slate-900 via-indigo-950 to-slate-900 border border-slate-800 p-8 shadow-2xl">
        <div className="absolute -right-12 -top-12 w-64 h-64 bg-indigo-500/10 rounded-full blur-3xl pointer-events-none" />

        <div className="relative z-10 flex flex-col md:flex-row md:items-center justify-between gap-6">
          <div>
            <h1 className="text-3xl font-extrabold text-white tracking-tight flex items-center gap-3">
              External Integrations & Connectors
            </h1>
            <p className="text-slate-400 text-sm max-w-2xl mt-1">
              Connect external calendars, messaging gateways, email providers, and custom outbound webhooks. External integrations are strictly isolated and never alter core appointment rules directly.
            </p>
          </div>

          <button
            onClick={fetchIntegrations}
            className="px-4 py-2.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-sm font-medium transition flex items-center gap-2 border border-slate-700 shrink-0"
          >
            <RefreshCw className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`} /> Refresh Integrations
          </button>
        </div>
      </div>

      {/* Integrations Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {integrations.map((item) => {
          const info = getProviderInfo(item.provider);
          const isConnected = item.status === 'connected';

          return (
            <div 
              key={item.provider}
              className="bg-slate-900/80 backdrop-blur-xl border border-slate-800 rounded-2xl p-6 shadow-xl flex flex-col justify-between hover:border-slate-700 transition duration-300"
            >
              <div>
                <div className="flex items-center justify-between mb-4">
                  <div className="p-3 bg-slate-800/80 rounded-xl border border-slate-700/50">
                    {info.icon}
                  </div>
                  {getStatusBadge(item.status)}
                </div>

                <h3 className="text-xl font-bold text-white">{info.title}</h3>
                <p className="text-slate-400 text-xs mt-1.5 min-h-[36px]">{info.description}</p>

                <div className="my-5 p-3.5 bg-slate-800/40 rounded-xl border border-slate-800 text-xs space-y-1.5">
                  <div className="flex justify-between text-slate-400">
                    <span>Account Context:</span>
                    <span className="text-white font-medium">{item.account_email || 'Not configured'}</span>
                  </div>
                  {item.last_synced_at && (
                    <div className="flex justify-between text-slate-400">
                      <span>Last Synchronized:</span>
                      <span className="text-slate-300">{new Date(item.last_synced_at).toLocaleString()}</span>
                    </div>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-3 pt-2">
                {isConnected ? (
                  <>
                    <button
                      onClick={() => handleReconnect(item.provider)}
                      className="flex-1 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-semibold transition border border-slate-700 flex items-center justify-center gap-1.5"
                    >
                      <RefreshCw className="w-3.5 h-3.5" /> Reconnect / Sync
                    </button>
                    <button
                      onClick={() => handleDisconnect(item.provider)}
                      className="px-4 py-2 rounded-xl bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 text-xs font-semibold transition border border-rose-500/20"
                    >
                      Disconnect
                    </button>
                  </>
                ) : (
                  <button
                    onClick={() => handleOpenConnectModal(item)}
                    className="w-full py-2.5 rounded-xl bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold transition flex items-center justify-center gap-1.5 shadow-lg shadow-indigo-600/20"
                  >
                    <Link2 className="w-4 h-4" /> Connect {info.title}
                  </button>
                )}
              </div>
            </div>
          );
        })}
      </div>

      {/* Modal for Connection / Webhook Config */}
      {activeModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 backdrop-blur-sm p-4">
          <div className="bg-slate-900 border border-slate-800 rounded-2xl max-w-md w-full p-6 shadow-2xl space-y-6 relative animate-in fade-in zoom-in-95">
            <button
              onClick={() => setActiveModal(null)}
              className="absolute top-4 right-4 text-slate-400 hover:text-white p-1 rounded-lg hover:bg-slate-800"
            >
              <X className="w-5 h-5" />
            </button>

            <div>
              <span className="text-xs font-semibold text-indigo-400 uppercase tracking-wider">Integration OAuth & Credentials</span>
              <h3 className="text-xl font-bold text-white mt-1">Configure {getProviderInfo(activeModal.provider).title}</h3>
              <p className="text-xs text-slate-400 mt-1">
                Establish encrypted credential mapping for your organization.
              </p>
            </div>

            <div className="space-y-4">
              {activeModal.provider === 'outbound_webhook' ? (
                <>
                  <div className="space-y-1">
                    <label className="text-xs font-medium text-slate-300 block">Target Webhook Endpoint URL:</label>
                    <input
                      type="url"
                      placeholder="https://api.yourdomain.com/webhooks/appointly"
                      value={webhookUrl}
                      onChange={(e) => setWebhookUrl(e.target.value)}
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="text-xs font-medium text-slate-300 block">HMAC Signature Secret Key:</label>
                    <div className="relative">
                      <input
                        type="text"
                        readOnly
                        value={webhookSecret}
                        className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3.5 py-2 text-xs font-mono text-indigo-300 pr-10 focus:outline-none"
                      />
                      <button
                        type="button"
                        onClick={() => {
                          navigator.clipboard.writeText(webhookSecret);
                          setCopiedSecret(true);
                          setTimeout(() => setCopiedSecret(false), 2000);
                        }}
                        className="absolute right-2 top-1/2 -translate-y-1/2 text-slate-400 hover:text-white p-1"
                      >
                        {copiedSecret ? <Check className="w-4 h-4 text-emerald-400" /> : <Copy className="w-4 h-4" />}
                      </button>
                    </div>
                  </div>
                </>
              ) : (
                <div className="space-y-1">
                  <label className="text-xs font-medium text-slate-300 block">Account Email / Account Identifier:</label>
                  <input
                    type="email"
                    placeholder="user@organization.com"
                    value={accountEmail}
                    onChange={(e) => setAccountEmail(e.target.value)}
                    className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3.5 py-2 text-xs text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
              )}
            </div>

            <div className="flex justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => setActiveModal(null)}
                className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs font-medium"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={isSubmitting}
                onClick={handleConnect}
                className="px-5 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-700 text-white text-xs font-bold transition flex items-center gap-2"
              >
                {isSubmitting ? 'Authenticating...' : 'Authorize & Connect'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
