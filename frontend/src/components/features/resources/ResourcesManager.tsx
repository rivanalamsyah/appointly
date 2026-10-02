import React, { useState } from 'react';
import {
  Box,
  Plus,
  Search,
  Filter,
  Edit2,
  Trash2,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  MapPin,
  Users,
  Settings,
  X,
  Layers,
  Sparkles
} from 'lucide-react';
import type { Resource, ResourceType, ResourceStatus } from '@/types/api';

// Sample pre-seeded resources for initial view
const INITIAL_RESOURCES: Resource[] = [
  {
    id: 'res-1',
    organization_id: 'org-demo',
    location_id: 'loc-1',
    name: 'Treatment Room 1 (Facial)',
    type: 'ROOM',
    capacity: 1,
    status: 'ACTIVE',
    description: 'Equipped with electric treatment bed, magnifying lamp, and steam machine.',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  },
  {
    id: 'res-2',
    organization_id: 'org-demo',
    location_id: 'loc-1',
    name: 'HydraFacial MD Machine #A',
    type: 'EQUIPMENT',
    capacity: 1,
    status: 'ACTIVE',
    description: 'Medical grade hydradermabrasion system.',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  },
  {
    id: 'res-3',
    organization_id: 'org-demo',
    location_id: 'loc-2',
    name: 'VIP Hair Styling Station #3',
    type: 'CHAIR',
    capacity: 1,
    status: 'ACTIVE',
    description: 'Takara Belmont luxury hydraulic styling chair with ring light mirror.',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  },
  {
    id: 'res-4',
    organization_id: 'org-demo',
    location_id: 'loc-1',
    name: 'IPL Laser Therapy Suite',
    type: 'ROOM',
    capacity: 1,
    status: 'MAINTENANCE',
    description: 'Scheduled calibration & lamp replacement until Oct 10.',
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  }
];

const SAMPLE_LOCATIONS = [
  { id: 'loc-1', name: 'Central Flagship Salon (Downtown)' },
  { id: 'loc-2', name: 'Northside Wellness Center' }
];

const SAMPLE_SERVICES = [
  { id: 'svc-1', name: 'Signature HydraFacial MD (60 min)' },
  { id: 'svc-2', name: 'IPL Skin Rejuvenation (45 min)' },
  { id: 'svc-3', name: 'VIP Hair Styling & Wash (90 min)' }
];

export default function ResourcesManager() {
  const [resources, setResources] = useState<Resource[]>(INITIAL_RESOURCES);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedLocation, setSelectedLocation] = useState<string>('ALL');
  const [selectedType, setSelectedType] = useState<string>('ALL');
  const [selectedStatus, setSelectedStatus] = useState<string>('ALL');

  // Modals state
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingResource, setEditingResource] = useState<Resource | null>(null);
  const [isAssignmentOpen, setIsAssignmentOpen] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  // Form inputs
  const [formName, setFormName] = useState('');
  const [formType, setFormType] = useState<ResourceType>('ROOM');
  const [formCapacity, setFormCapacity] = useState<number>(1);
  const [formLocationId, setFormLocationId] = useState<string>('loc-1');
  const [formStatus, setFormStatus] = useState<ResourceStatus>('ACTIVE');
  const [formDescription, setFormDescription] = useState('');
  const [formError, setFormError] = useState<string | null>(null);

  // Service Assignment state
  const [selectedServiceId, setSelectedServiceId] = useState<string>('svc-1');
  const [serviceResourceMap, setServiceResourceMap] = useState<Record<string, string[]>>({
    'svc-1': ['res-1', 'res-2'],
    'svc-2': ['res-4'],
    'svc-3': ['res-3']
  });

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3000);
  };

  const handleOpenAdd = () => {
    setEditingResource(null);
    setFormName('');
    setFormType('ROOM');
    setFormCapacity(1);
    setFormLocationId('loc-1');
    setFormStatus('ACTIVE');
    setFormDescription('');
    setFormError(null);
    setIsFormOpen(true);
  };

  const handleOpenEdit = (res: Resource) => {
    setEditingResource(res);
    setFormName(res.name);
    setFormType(res.type);
    setFormCapacity(res.capacity);
    setFormLocationId(res.location_id || 'loc-1');
    setFormStatus(res.status);
    setFormDescription(res.description || '');
    setFormError(null);
    setIsFormOpen(true);
  };

  const handleSaveResource = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formName.trim()) {
      setFormError('Resource name is required.');
      return;
    }
    if (formCapacity < 1) {
      setFormError('Capacity must be at least 1.');
      return;
    }

    if (editingResource) {
      // Update
      setResources(prev =>
        prev.map(r =>
          r.id === editingResource.id
            ? {
                ...r,
                name: formName.trim(),
                type: formType,
                capacity: formCapacity,
                location_id: formLocationId,
                status: formStatus,
                description: formDescription.trim(),
                updated_at: new Date().toISOString()
              }
            : r
        )
      );
      showToast(`Resource "${formName.trim()}" updated successfully.`);
    } else {
      // Create
      const newRes: Resource = {
        id: `res-${Date.now()}`,
        organization_id: 'org-demo',
        location_id: formLocationId,
        name: formName.trim(),
        type: formType,
        capacity: formCapacity,
        status: formStatus,
        description: formDescription.trim(),
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      };
      setResources(prev => [newRes, ...prev]);
      showToast(`Resource "${formName.trim()}" created successfully.`);
    }
    setIsFormOpen(false);
  };

  const handleDeleteResource = (id: string, name: string) => {
    if (confirm(`Are you sure you want to delete resource "${name}"?`)) {
      setResources(prev => prev.filter(r => r.id !== id));
      showToast(`Resource "${name}" deleted.`);
    }
  };

  const handleToggleStatus = (res: Resource, newStatus: ResourceStatus) => {
    setResources(prev =>
      prev.map(r => (r.id === res.id ? { ...r, status: newStatus, updated_at: new Date().toISOString() } : r))
    );
    showToast(`Status of "${res.name}" changed to ${newStatus}.`);
  };

  // Toggle assigned resource for current service in modal
  const handleToggleServiceResource = (resId: string) => {
    setServiceResourceMap(prev => {
      const current = prev[selectedServiceId] || [];
      const exists = current.includes(resId);
      const updated = exists ? current.filter(id => id !== resId) : [...current, resId];
      return { ...prev, [selectedServiceId]: updated };
    });
  };

  // Filtered resources
  const filteredResources = resources.filter(res => {
    if (selectedLocation !== 'ALL' && res.location_id !== selectedLocation) return false;
    if (selectedType !== 'ALL' && res.type !== selectedType) return false;
    if (selectedStatus !== 'ALL' && res.status !== selectedStatus) return false;
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      const matchName = res.name.toLowerCase().includes(q);
      const matchDesc = res.description?.toLowerCase().includes(q) || false;
      if (!matchName && !matchDesc) return false;
    }
    return true;
  });

  const getTypeBadgeColor = (type: ResourceType) => {
    switch (type) {
      case 'ROOM':
        return 'bg-indigo-500/10 text-indigo-400 border-indigo-500/30';
      case 'EQUIPMENT':
        return 'bg-purple-500/10 text-purple-400 border-purple-500/30';
      case 'CHAIR':
        return 'bg-cyan-500/10 text-cyan-400 border-cyan-500/30';
      case 'STUDIO':
        return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/30';
      default:
        return 'bg-slate-500/10 text-slate-400 border-slate-500/30';
    }
  };

  const getStatusBadge = (status: ResourceStatus) => {
    switch (status) {
      case 'ACTIVE':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            <CheckCircle2 className="w-3 h-3" /> Active
          </span>
        );
      case 'MAINTENANCE':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-amber-500/10 text-amber-400 border border-amber-500/20">
            <AlertTriangle className="w-3 h-3" /> Maintenance
          </span>
        );
      case 'INACTIVE':
        return (
          <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-medium bg-slate-500/10 text-slate-400 border border-slate-500/20">
            <XCircle className="w-3 h-3" /> Inactive
          </span>
        );
    }
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

      {/* Header Banner */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-6 rounded-2xl bg-gradient-to-r from-slate-900 via-indigo-950/40 to-slate-900 border border-[var(--color-surface-border)] shadow-xl relative overflow-hidden">
        <div className="relative z-10">
          <div className="flex items-center gap-3">
            <div className="p-2.5 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-indigo-400">
              <Box className="w-6 h-6" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-white font-[var(--font-display)]">Resource Management</h1>
              <p className="text-xs text-[var(--color-text-muted)] mt-0.5">
                Manage rooms, equipment, chairs, and physical facilities required by your services.
              </p>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-3 relative z-10">
          <button
            onClick={() => setIsAssignmentOpen(true)}
            className="flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-semibold text-[var(--color-text-primary)] bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] hover:bg-[var(--color-surface-hover)] transition-all duration-150"
          >
            <Layers className="w-4 h-4 text-indigo-400" /> Service Assignments
          </button>
          <button
            onClick={handleOpenAdd}
            className="flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 shadow-lg shadow-indigo-600/30 transition-all duration-150"
          >
            <Plus className="w-4 h-4" /> Add Resource
          </button>
        </div>
      </div>

      {/* Search & Filters */}
      <div className="p-4 rounded-xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] flex flex-wrap items-center gap-4">
        {/* Search */}
        <div className="relative flex-1 min-w-[240px]">
          <Search className="w-4 h-4 text-[var(--color-text-muted)] absolute left-3.5 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            placeholder="Search resource by name or description..."
            className="w-full pl-10 pr-4 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white placeholder-[var(--color-text-muted)] focus:outline-none focus:border-indigo-500"
          />
        </div>

        {/* Location Filter */}
        <div className="flex items-center gap-2">
          <MapPin className="w-4 h-4 text-[var(--color-text-muted)]" />
          <select
            value={selectedLocation}
            onChange={e => setSelectedLocation(e.target.value)}
            className="px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
          >
            <option value="ALL">All Locations</option>
            {SAMPLE_LOCATIONS.map(loc => (
              <option key={loc.id} value={loc.id}>
                {loc.name}
              </option>
            ))}
          </select>
        </div>

        {/* Type Filter */}
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-[var(--color-text-muted)]" />
          <select
            value={selectedType}
            onChange={e => setSelectedType(e.target.value)}
            className="px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
          >
            <option value="ALL">All Types</option>
            <option value="ROOM">Room</option>
            <option value="EQUIPMENT">Equipment</option>
            <option value="CHAIR">Chair</option>
            <option value="STUDIO">Studio</option>
            <option value="FACILITY">Facility</option>
            <option value="OTHER">Other</option>
          </select>
        </div>

        {/* Status Filter */}
        <select
          value={selectedStatus}
          onChange={e => setSelectedStatus(e.target.value)}
          className="px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
        >
          <option value="ALL">All Statuses</option>
          <option value="ACTIVE">Active</option>
          <option value="MAINTENANCE">Maintenance</option>
          <option value="INACTIVE">Inactive</option>
        </select>
      </div>

      {/* Resource Grid */}
      {filteredResources.length === 0 ? (
        <div className="p-12 text-center rounded-2xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)]">
          <Box className="w-12 h-12 text-[var(--color-text-muted)] mx-auto mb-3 opacity-50" />
          <h3 className="text-sm font-semibold text-white">No resources found</h3>
          <p className="text-xs text-[var(--color-text-muted)] mt-1">
            Try adjusting your search query or filter parameters.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredResources.map(res => {
            const locationObj = SAMPLE_LOCATIONS.find(l => l.id === res.location_id);
            return (
              <div
                key={res.id}
                className="p-5 rounded-xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] hover:border-indigo-500/40 transition-all duration-200 flex flex-col justify-between group"
              >
                <div>
                  <div className="flex items-start justify-between gap-3 mb-3">
                    <div className="flex items-center gap-2">
                      <span className={`px-2.5 py-0.5 rounded-md text-[10px] font-bold tracking-wider uppercase border ${getTypeBadgeColor(res.type)}`}>
                        {res.type}
                      </span>
                      <span className="flex items-center gap-1 text-[11px] text-[var(--color-text-muted)] bg-[var(--color-surface-base)] px-2 py-0.5 rounded border border-[var(--color-surface-border)]">
                        <Users className="w-3 h-3 text-indigo-400" /> Cap: {res.capacity}
                      </span>
                    </div>
                    {getStatusBadge(res.status)}
                  </div>

                  <h3 className="text-sm font-bold text-white font-[var(--font-display)] mb-1 group-hover:text-indigo-300 transition-colors">
                    {res.name}
                  </h3>
                  {res.description && (
                    <p className="text-xs text-[var(--color-text-muted)] line-clamp-2 mb-3">
                      {res.description}
                    </p>
                  )}
                </div>

                <div className="pt-4 mt-3 border-t border-[var(--color-surface-border)] flex items-center justify-between">
                  <div className="flex items-center gap-1.5 text-xs text-[var(--color-text-muted)]">
                    <MapPin className="w-3.5 h-3.5 text-indigo-400 flex-shrink-0" />
                    <span className="truncate max-w-[140px]">{locationObj?.name || 'All Branches'}</span>
                  </div>

                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => handleOpenEdit(res)}
                      title="Edit Resource"
                      className="p-1.5 rounded-lg text-[var(--color-text-muted)] hover:text-white hover:bg-[var(--color-surface-hover)] transition-colors"
                    >
                      <Edit2 className="w-3.5 h-3.5" />
                    </button>
                    <button
                      onClick={() => handleDeleteResource(res.id, res.name)}
                      title="Delete Resource"
                      className="p-1.5 rounded-lg text-[var(--color-text-muted)] hover:text-red-400 hover:bg-red-500/10 transition-colors"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Add / Edit Resource Modal */}
      {isFormOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
          <div className="w-full max-w-md rounded-2xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] shadow-2xl p-6">
            <div className="flex items-center justify-between pb-4 mb-4 border-b border-[var(--color-surface-border)]">
              <h2 className="text-base font-bold text-white font-[var(--font-display)]">
                {editingResource ? 'Edit Resource' : 'Add New Resource'}
              </h2>
              <button
                onClick={() => setIsFormOpen(false)}
                className="p-1 rounded-lg text-[var(--color-text-muted)] hover:text-white"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {formError && (
              <div className="mb-4 p-3 rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 text-xs flex items-center gap-2">
                <AlertTriangle className="w-4 h-4 flex-shrink-0" />
                <span>{formError}</span>
              </div>
            )}

            <form onSubmit={handleSaveResource} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">
                  Resource Name *
                </label>
                <input
                  type="text"
                  required
                  value={formName}
                  onChange={e => setFormName(e.target.value)}
                  placeholder="e.g. Treatment Room 1 or Laser Scanner"
                  className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">
                    Resource Type
                  </label>
                  <select
                    value={formType}
                    onChange={e => setFormType(e.target.value as ResourceType)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  >
                    <option value="ROOM">Room</option>
                    <option value="EQUIPMENT">Equipment</option>
                    <option value="CHAIR">Chair</option>
                    <option value="STUDIO">Studio</option>
                    <option value="FACILITY">Facility</option>
                    <option value="OTHER">Other</option>
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">
                    Capacity (Units)
                  </label>
                  <input
                    type="number"
                    min={1}
                    value={formCapacity}
                    onChange={e => setFormCapacity(parseInt(e.target.value) || 1)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">
                    Location
                  </label>
                  <select
                    value={formLocationId}
                    onChange={e => setFormLocationId(e.target.value)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  >
                    {SAMPLE_LOCATIONS.map(loc => (
                      <option key={loc.id} value={loc.id}>
                        {loc.name}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">
                    Status
                  </label>
                  <select
                    value={formStatus}
                    onChange={e => setFormStatus(e.target.value as ResourceStatus)}
                    className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                  >
                    <option value="ACTIVE">Active</option>
                    <option value="MAINTENANCE">Maintenance</option>
                    <option value="INACTIVE">Inactive</option>
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">
                  Description
                </label>
                <textarea
                  rows={3}
                  value={formDescription}
                  onChange={e => setFormDescription(e.target.value)}
                  placeholder="Optional details or specifications..."
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
                  Save Resource
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Service Resource Assignment Modal */}
      {isAssignmentOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-fade-in">
          <div className="w-full max-w-lg rounded-2xl bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)] shadow-2xl p-6">
            <div className="flex items-center justify-between pb-4 mb-4 border-b border-[var(--color-surface-border)]">
              <div className="flex items-center gap-2">
                <Layers className="w-5 h-5 text-indigo-400" />
                <h2 className="text-base font-bold text-white font-[var(--font-display)]">
                  Service-Resource Requirements
                </h2>
              </div>
              <button
                onClick={() => setIsAssignmentOpen(false)}
                className="p-1 rounded-lg text-[var(--color-text-muted)] hover:text-white"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <p className="text-xs text-[var(--color-text-muted)] mb-4">
              Specify which equipment or room resources are required to perform each service. Leave empty for services that do not require dedicated physical resources.
            </p>

            <div className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-1">
                  Select Service
                </label>
                <select
                  value={selectedServiceId}
                  onChange={e => setSelectedServiceId(e.target.value)}
                  className="w-full px-3 py-2 text-xs rounded-lg bg-[var(--color-surface-base)] border border-[var(--color-surface-border)] text-white focus:outline-none focus:border-indigo-500"
                >
                  {SAMPLE_SERVICES.map(svc => (
                    <option key={svc.id} value={svc.id}>
                      {svc.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-2">
                  Required Resources for selected service:
                </label>
                <div className="max-h-60 overflow-y-auto space-y-2 pr-1">
                  {resources.map(res => {
                    const assignedList = serviceResourceMap[selectedServiceId] || [];
                    const isChecked = assignedList.includes(res.id);
                    return (
                      <label
                        key={res.id}
                        className={`flex items-center justify-between p-3 rounded-xl border transition-all cursor-pointer ${
                          isChecked
                            ? 'bg-indigo-500/10 border-indigo-500/40 text-white'
                            : 'bg-[var(--color-surface-base)] border-[var(--color-surface-border)] text-[var(--color-text-muted)] hover:border-slate-700'
                        }`}
                      >
                        <div className="flex items-center gap-3">
                          <input
                            type="checkbox"
                            checked={isChecked}
                            onChange={() => handleToggleServiceResource(res.id)}
                            className="w-4 h-4 rounded text-indigo-600 bg-[var(--color-surface-base)] border-[var(--color-surface-border)] focus:ring-indigo-500"
                          />
                          <div>
                            <p className="text-xs font-semibold text-white">{res.name}</p>
                            <p className="text-[10px] text-[var(--color-text-muted)]">
                              Type: {res.type} • Status: {res.status}
                            </p>
                          </div>
                        </div>
                        <span className="text-[10px] px-2 py-0.5 rounded bg-[var(--color-surface-raised)] border border-[var(--color-surface-border)]">
                          Cap: {res.capacity}
                        </span>
                      </label>
                    );
                  })}
                </div>
              </div>
            </div>

            <div className="flex items-center justify-end gap-3 pt-4 mt-6 border-t border-[var(--color-surface-border)]">
              <button
                type="button"
                onClick={() => setIsAssignmentOpen(false)}
                className="px-4 py-2 rounded-lg text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500"
              >
                Done
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
