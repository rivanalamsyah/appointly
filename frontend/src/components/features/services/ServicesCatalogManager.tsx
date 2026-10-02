'use client';

import React, { useState } from 'react';
import { 
  Scissors, 
  Plus, 
  Clock, 
  DollarSign, 
  Tag, 
  Edit3, 
  Trash2, 
  Eye, 
  EyeOff, 
  Check, 
  Sparkles, 
  AlertCircle,
  FolderPlus
} from 'lucide-react';
import type { Service, ServiceCategory } from '@/types/api';

interface ServicesCatalogManagerProps {
  initialCategories?: Partial<ServiceCategory>[];
  initialServices?: Partial<Service>[];
}

export default function ServicesCatalogManager({
  initialCategories = [
    { id: 'cat-1', name: 'Haircuts & Styling', color: '#8b5cf6', is_active: true },
    { id: 'cat-2', name: 'Color & Highlights', color: '#ec4899', is_active: true },
    { id: 'cat-3', name: 'Beard & Grooming', color: '#3b82f6', is_active: true },
  ],
  initialServices = [
    {
      id: 'srv-1',
      category_id: 'cat-1',
      name: 'Signature Haircut & Styling',
      description: 'Precision cut tailored to your face shape, includes shampoo and styling.',
      duration_minutes: 45,
      buffer_before_minutes: 5,
      buffer_after_minutes: 10,
      price_cents: 6500,
      currency: 'USD',
      is_active: true,
      is_public: true,
    },
    {
      id: 'srv-2',
      category_id: 'cat-2',
      name: 'Full Balayage & Gloss Toner',
      description: 'Custom dimensional balayage highlights with gloss toner finish.',
      duration_minutes: 120,
      buffer_before_minutes: 10,
      buffer_after_minutes: 20,
      price_cents: 18000,
      currency: 'USD',
      is_active: true,
      is_public: true,
    },
    {
      id: 'srv-3',
      category_id: 'cat-3',
      name: 'Hot Towel Beard Beard Trim',
      description: 'Straight razor line-up with hot towel and beard oil treatment.',
      duration_minutes: 30,
      buffer_before_minutes: 0,
      buffer_after_minutes: 5,
      price_cents: 3500,
      currency: 'USD',
      is_active: false,
      is_public: true,
    }
  ]
}: ServicesCatalogManagerProps) {
  const [categories, setCategories] = useState(initialCategories);
  const [services, setServices] = useState(initialServices);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');

  // Service Modal State
  const [isServiceModalOpen, setIsServiceModalOpen] = useState(false);
  const [editingService, setEditingService] = useState<Partial<Service> | null>(null);

  // Form Fields
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [durationMinutes, setDurationMinutes] = useState<number>(30);
  const [bufferBeforeMinutes, setBufferBeforeMinutes] = useState<number>(0);
  const [bufferAfterMinutes, setBufferAfterMinutes] = useState<number>(10);
  const [priceCents, setPriceCents] = useState<number>(5000);
  const [isPublic, setIsPublic] = useState(true);

  // Category Modal State
  const [isCategoryModalOpen, setIsCategoryModalOpen] = useState(false);
  const [newCatName, setNewCatName] = useState('');

  const [toast, setToast] = useState<string | null>(null);

  const filteredServices = selectedCategory === 'all'
    ? services
    : services.filter((s) => s.category_id === selectedCategory);

  const formatPrice = (cents: number, currency: string = 'USD') => {
    return new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(cents / 100);
  };

  const openCreateServiceModal = () => {
    setEditingService(null);
    setName('');
    setDescription('');
    setCategoryId(categories[0]?.id || '');
    setDurationMinutes(30);
    setBufferBeforeMinutes(0);
    setBufferAfterMinutes(10);
    setPriceCents(5000);
    setIsPublic(true);
    setIsServiceModalOpen(true);
  };

  const openEditServiceModal = (svc: Partial<Service>) => {
    setEditingService(svc);
    setName(svc.name || '');
    setDescription(svc.description || '');
    setCategoryId(svc.category_id || '');
    setDurationMinutes(svc.duration_minutes || 30);
    setBufferBeforeMinutes(svc.buffer_before_minutes || 0);
    setBufferAfterMinutes(svc.buffer_after_minutes || 0);
    setPriceCents(svc.price_cents || 0);
    setIsPublic(svc.is_public ?? true);
    setIsServiceModalOpen(true);
  };

  const handleSaveService = (e: React.FormEvent) => {
    e.preventDefault();
    if (durationMinutes <= 0) return;

    if (editingService) {
      setServices((prev) =>
        prev.map((s) =>
          s.id === editingService.id
            ? {
                ...s,
                name,
                description,
                category_id: categoryId,
                duration_minutes: durationMinutes,
                buffer_before_minutes: bufferBeforeMinutes,
                buffer_after_minutes: bufferAfterMinutes,
                price_cents: priceCents,
                is_public: isPublic,
              }
            : s
        )
      );
      setToast('Service updated successfully');
    } else {
      const newSvc: Partial<Service> = {
        id: `srv-${Date.now()}`,
        name,
        description,
        category_id: categoryId,
        duration_minutes: durationMinutes,
        buffer_before_minutes: bufferBeforeMinutes,
        buffer_after_minutes: bufferAfterMinutes,
        price_cents: priceCents,
        currency: 'USD',
        is_active: true,
        is_public: isPublic,
      };
      setServices((prev) => [...prev, newSvc]);
      setToast('New service added to catalog');
    }
    setIsServiceModalOpen(false);
    setTimeout(() => setToast(null), 3000);
  };

  const handleToggleStatus = (id: string) => {
    setServices((prev) =>
      prev.map((s) => (s.id === id ? { ...s, is_active: !s.is_active } : s))
    );
    setToast('Service availability updated');
    setTimeout(() => setToast(null), 3000);
  };

  const handleCreateCategory = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newCatName.trim()) return;
    const newCat = {
      id: `cat-${Date.now()}`,
      name: newCatName,
      color: '#8b5cf6',
      is_active: true,
    };
    setCategories((prev) => [...prev, newCat]);
    setNewCatName('');
    setIsCategoryModalOpen(false);
    setToast('Service category created');
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

      {/* Header & Controls */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-slate-900/80 border border-slate-800 rounded-3xl p-6 shadow-xl backdrop-blur-xl">
        <div>
          <h3 className="text-lg font-bold text-white tracking-tight flex items-center gap-2">
            <Scissors className="w-5 h-5 text-violet-400" />
            Service Catalog Management
          </h3>
          <p className="text-xs text-slate-400 mt-0.5">Configure bookable offerings, duration buffers, and pricing tiers.</p>
        </div>

        <div className="flex items-center gap-3 flex-wrap">
          <button
            onClick={() => setIsCategoryModalOpen(true)}
            className="bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold px-4 py-2.5 rounded-xl border border-slate-700 transition-colors flex items-center gap-2"
          >
            <FolderPlus className="w-4 h-4 text-violet-400" />
            Add Category
          </button>
          <button
            onClick={openCreateServiceModal}
            className="bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white text-xs font-bold px-5 py-2.5 rounded-xl transition-all shadow-lg shadow-violet-900/30 flex items-center gap-2"
          >
            <Plus className="w-4 h-4" />
            Add Service
          </button>
        </div>
      </div>

      {/* Category Tabs Filter */}
      <div className="flex items-center gap-2 overflow-x-auto pb-1 scrollbar-none">
        <button
          onClick={() => setSelectedCategory('all')}
          className={`px-4 py-2 rounded-xl text-xs font-semibold transition-all whitespace-nowrap ${
            selectedCategory === 'all'
              ? 'bg-violet-600 text-white shadow-md shadow-violet-900/40'
              : 'bg-slate-900/80 border border-slate-800 text-slate-400 hover:text-slate-200'
          }`}
        >
          All Services ({services.length})
        </button>
        {categories.map((cat) => (
          <button
            key={cat.id}
            onClick={() => setSelectedCategory(cat.id!)}
            className={`px-4 py-2 rounded-xl text-xs font-semibold transition-all whitespace-nowrap flex items-center gap-2 ${
              selectedCategory === cat.id
                ? 'bg-violet-600 text-white shadow-md shadow-violet-900/40'
                : 'bg-slate-900/80 border border-slate-800 text-slate-400 hover:text-slate-200'
            }`}
          >
            <span className="w-2 h-2 rounded-full" style={{ backgroundColor: cat.color || '#8b5cf6' }} />
            {cat.name}
          </button>
        ))}
      </div>

      {/* Services Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredServices.map((svc) => (
          <div
            key={svc.id}
            className={`bg-slate-900/80 border rounded-3xl p-6 space-y-4 transition-all shadow-xl backdrop-blur-xl flex flex-col justify-between ${
              svc.is_active
                ? 'border-slate-800 hover:border-slate-700'
                : 'border-slate-800/40 opacity-60'
            }`}
          >
            <div className="space-y-3">
              <div className="flex items-start justify-between gap-3">
                <div>
                  <h4 className="font-bold text-white text-base">{svc.name}</h4>
                  <span className="text-[10px] font-mono text-violet-400 uppercase tracking-wider">
                    {categories.find((c) => c.id === svc.category_id)?.name || 'General'}
                  </span>
                </div>

                {/* Status Toggle Switch */}
                <button
                  type="button"
                  onClick={() => handleToggleStatus(svc.id!)}
                  className={`px-2.5 py-1 rounded-full text-[10px] font-bold transition-all border ${
                    svc.is_active
                      ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30'
                      : 'bg-slate-800 text-slate-400 border-slate-700'
                  }`}
                >
                  {svc.is_active ? 'Active' : 'Inactive'}
                </button>
              </div>

              <p className="text-xs text-slate-400 line-clamp-2 leading-relaxed">{svc.description}</p>

              {/* Timing Breakdown */}
              <div className="bg-slate-950/60 border border-slate-800/80 rounded-2xl p-3 grid grid-cols-3 text-center text-xs">
                <div>
                  <span className="text-[10px] text-slate-500 block">Service</span>
                  <span className="font-mono font-semibold text-slate-200">{svc.duration_minutes}m</span>
                </div>
                <div>
                  <span className="text-[10px] text-slate-500 block">Prep Buffer</span>
                  <span className="font-mono text-slate-400">{svc.buffer_before_minutes || 0}m</span>
                </div>
                <div>
                  <span className="text-[10px] text-slate-500 block">Clean Buffer</span>
                  <span className="font-mono text-slate-400">{svc.buffer_after_minutes || 0}m</span>
                </div>
              </div>
            </div>

            <div className="pt-3 border-t border-slate-800/80 flex items-center justify-between">
              <div className="text-lg font-bold text-emerald-400">
                {formatPrice(svc.price_cents || 0, svc.currency)}
              </div>

              <div className="flex items-center gap-2">
                <button
                  onClick={() => openEditServiceModal(svc)}
                  className="p-2 text-slate-400 hover:text-white hover:bg-slate-800 rounded-xl transition-colors"
                >
                  <Edit3 className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Create / Edit Service Modal */}
      {isServiceModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md animate-fade-in">
          <div className="bg-slate-900 border border-slate-800 rounded-3xl p-6 md:p-8 w-full max-w-lg shadow-2xl space-y-6 max-h-[90vh] overflow-y-auto custom-scrollbar">
            <div className="flex justify-between items-center pb-4 border-b border-slate-800">
              <h3 className="text-lg font-bold text-white tracking-tight">
                {editingService ? 'Edit Service Offering' : 'Add New Service Offering'}
              </h3>
              <button
                onClick={() => setIsServiceModalOpen(false)}
                className="text-slate-400 hover:text-white text-xs font-semibold px-2 py-1"
              >
                Close
              </button>
            </div>

            <form onSubmit={handleSaveService} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Service Name *</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Signature Haircut & Styling"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Category</label>
                <select
                  value={categoryId}
                  onChange={(e) => setCategoryId(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 cursor-pointer"
                >
                  {categories.map((c) => (
                    <option key={c.id} value={c.id}>{c.name}</option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Description</label>
                <textarea
                  rows={3}
                  placeholder="Service details shown on customer booking page..."
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500 resize-none"
                />
              </div>

              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">Duration (Min) *</label>
                  <input
                    type="number"
                    min={5}
                    required
                    value={durationMinutes}
                    onChange={(e) => setDurationMinutes(parseInt(e.target.value) || 0)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-3 text-slate-200 text-xs font-mono focus:outline-none focus:border-violet-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">Prep Buffer (Min)</label>
                  <input
                    type="number"
                    min={0}
                    value={bufferBeforeMinutes}
                    onChange={(e) => setBufferBeforeMinutes(parseInt(e.target.value) || 0)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-3 text-slate-200 text-xs font-mono focus:outline-none focus:border-violet-500"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">Clean Buffer (Min)</label>
                  <input
                    type="number"
                    min={0}
                    value={bufferAfterMinutes}
                    onChange={(e) => setBufferAfterMinutes(parseInt(e.target.value) || 0)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-3 py-3 text-slate-200 text-xs font-mono focus:outline-none focus:border-violet-500"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Price (Cents / Smallest Unit) *</label>
                <input
                  type="number"
                  min={0}
                  required
                  value={priceCents}
                  onChange={(e) => setPriceCents(parseInt(e.target.value) || 0)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs font-mono focus:outline-none focus:border-violet-500"
                />
                <span className="text-[10px] text-slate-500 mt-1 block">
                  Formatted preview: <strong className="text-emerald-400">{formatPrice(priceCents)}</strong>
                </span>
              </div>

              <div className="flex items-center justify-between p-4 bg-slate-950/60 border border-slate-800 rounded-2xl">
                <div>
                  <h4 className="text-xs font-semibold text-white">Public Customer Portal Visible</h4>
                  <p className="text-[11px] text-slate-400">Show this service on customer online booking page</p>
                </div>
                <input
                  type="checkbox"
                  checked={isPublic}
                  onChange={(e) => setIsPublic(e.target.checked)}
                  className="w-5 h-5 accent-violet-600 rounded cursor-pointer"
                />
              </div>

              <div className="flex justify-end gap-3 pt-4 border-t border-slate-800">
                <button
                  type="button"
                  onClick={() => setIsServiceModalOpen(false)}
                  className="px-4 py-2.5 text-xs font-semibold text-slate-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="bg-gradient-to-r from-violet-600 to-indigo-600 text-white text-xs font-bold px-6 py-2.5 rounded-xl shadow-lg"
                >
                  Save Service
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Category Modal */}
      {isCategoryModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md animate-fade-in">
          <div className="bg-slate-900 border border-slate-800 rounded-3xl p-6 w-full max-w-sm shadow-2xl space-y-4">
            <h3 className="text-base font-bold text-white">Add Service Category</h3>
            <form onSubmit={handleCreateCategory} className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Category Name *</label>
                <input
                  type="text"
                  required
                  placeholder="e.g. Massages & Spa"
                  value={newCatName}
                  onChange={(e) => setNewCatName(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-xs focus:outline-none focus:border-violet-500"
                />
              </div>
              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setIsCategoryModalOpen(false)}
                  className="px-4 py-2 text-xs font-semibold text-slate-400"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="bg-violet-600 text-white text-xs font-bold px-5 py-2 rounded-xl"
                >
                  Create Category
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
