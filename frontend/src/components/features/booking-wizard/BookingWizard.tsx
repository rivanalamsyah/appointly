'use client';

import React, { useState } from 'react';
import { 
  Check, 
  ChevronRight, 
  ChevronLeft, 
  Clock, 
  Sparkles, 
  User, 
  Calendar as CalendarIcon, 
  ShieldCheck, 
  CreditCard,
  Building2,
  Info
} from 'lucide-react';
import type { Service, Staff, AvailableSlot, Organization } from '@/types/api';

interface BookingWizardProps {
  organization?: Partial<Organization>;
  services?: Service[];
  staffMembers?: Staff[];
  tenantSlug?: string;
}

export default function BookingWizard({
  organization = { name: 'Luxe Salon & Spa', currency: 'USD' },
  services = [
    {
      id: 'srv-1',
      tenant_id: 't-1',
      name: 'Signature Haircut & Styling',
      slug: 'haircut-styling',
      description: 'Precision cut tailored to your face shape, includes deep cleansing shampoo and professional blow-dry styling.',
      duration_minutes: 45,
      buffer_time_before_minutes: 0,
      buffer_time_after_minutes: 15,
      price_cents: 6500,
      currency: 'USD',
      requires_deposit: true,
      deposit_amount_cents: 2000,
      is_active: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
    {
      id: 'srv-2',
      tenant_id: 't-1',
      name: 'Beard Grooming & Hot Towel Treatment',
      slug: 'beard-grooming',
      description: 'Traditional straight razor edge-up with hot towel treatment and moisturizing beard oil finish.',
      duration_minutes: 30,
      buffer_time_before_minutes: 0,
      buffer_time_after_minutes: 10,
      price_cents: 3500,
      currency: 'USD',
      requires_deposit: false,
      deposit_amount_cents: 0,
      is_active: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
    {
      id: 'srv-3',
      tenant_id: 't-1',
      name: 'Full Color & Highlights Package',
      slug: 'full-color-highlights',
      description: 'Custom balayage or foil highlights paired with gloss toner and restorative keratin treatment.',
      duration_minutes: 120,
      buffer_time_before_minutes: 10,
      buffer_time_after_minutes: 20,
      price_cents: 18000,
      currency: 'USD',
      requires_deposit: true,
      deposit_amount_cents: 5000,
      is_active: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }
  ],
  staffMembers = [
    {
      id: 'stf-1',
      tenant_id: 't-1',
      user_id: 'u-1',
      name: 'Alexander Wright',
      email: 'alex@example.com',
      avatar_url: '',
      title: 'Master Stylist',
      bio: '10+ years experience in precision cutting and modern texturizing techniques.',
      color_hex: '#8b5cf6',
      is_active: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    },
    {
      id: 'stf-2',
      tenant_id: 't-1',
      user_id: 'u-2',
      name: 'Elena Rostova',
      email: 'elena@example.com',
      avatar_url: '',
      title: 'Color Specialist',
      bio: 'Certified balayage artist specializing in dimensional color and hair restoration.',
      color_hex: '#ec4899',
      is_active: true,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }
  ],
}: BookingWizardProps) {
  const [currentStep, setCurrentStep] = useState<number>(1);
  const [selectedService, setSelectedService] = useState<Service | null>(services[0] || null);
  const [selectedStaff, setSelectedStaff] = useState<Staff | null>(null); // null means 'Any Staff'
  const [selectedDate, setSelectedDate] = useState<string>(new Date().toISOString().split('T')[0]);
  const [selectedSlot, setSelectedSlot] = useState<string | null>(null);

  // Form details
  const [customerName, setCustomerName] = useState('');
  const [customerEmail, setCustomerEmail] = useState('');
  const [customerPhone, setCustomerPhone] = useState('');
  const [notes, setNotes] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [bookingSuccess, setBookingSuccess] = useState(false);

  // Sample slots
  const mockSlots = [
    '09:00', '09:45', '10:30', '11:15', '13:00', '13:45', '14:30', '15:15', '16:00'
  ];

  const formatPrice = (cents: number, currency: string = 'USD') => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
    }).format(cents / 100);
  };

  const handleNext = () => {
    if (currentStep === 1 && !selectedService) return;
    if (currentStep === 3 && !selectedSlot) return;
    if (currentStep < 5) setCurrentStep((prev) => prev + 1);
  };

  const handlePrev = () => {
    if (currentStep > 1) setCurrentStep((prev) => prev - 1);
  };

  const handleSubmitBooking = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    // Simulate booking API call
    setTimeout(() => {
      setIsSubmitting(false);
      setBookingSuccess(true);
    }, 1200);
  };

  if (bookingSuccess) {
    return (
      <div className="max-w-2xl mx-auto bg-slate-900/80 border border-slate-800 rounded-3xl p-8 md:p-12 text-center shadow-2xl backdrop-blur-xl animate-fade-in">
        <div className="w-20 h-20 bg-emerald-500/10 border border-emerald-500/30 rounded-full flex items-center justify-center mx-auto mb-6">
          <Check className="w-10 h-10 text-emerald-400" />
        </div>
        <h2 className="text-3xl font-extrabold text-white mb-3 tracking-tight">Booking Confirmed!</h2>
        <p className="text-slate-300 text-sm max-w-md mx-auto mb-6">
          Thank you, <span className="font-semibold text-white">{customerName}</span>. Your appointment for{' '}
          <span className="text-violet-400 font-semibold">{selectedService?.name}</span> has been successfully scheduled.
        </p>
        
        <div className="bg-slate-950/60 border border-slate-800/80 rounded-2xl p-6 mb-8 text-left space-y-3">
          <div className="flex justify-between items-center text-xs text-slate-400 pb-3 border-b border-slate-800">
            <span>Date & Time</span>
            <span className="font-mono text-slate-200 font-semibold">{selectedDate} at {selectedSlot}</span>
          </div>
          <div className="flex justify-between items-center text-xs text-slate-400 pb-3 border-b border-slate-800">
            <span>Staff Member</span>
            <span className="text-slate-200 font-medium">{selectedStaff ? selectedStaff.name : 'Any Available Specialist'}</span>
          </div>
          <div className="flex justify-between items-center text-xs text-slate-400">
            <span>Total Price</span>
            <span className="text-emerald-400 font-bold text-sm">{formatPrice(selectedService?.price_cents || 0)}</span>
          </div>
        </div>

        <button
          onClick={() => {
            setBookingSuccess(false);
            setCurrentStep(1);
          }}
          className="bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold px-6 py-3 rounded-xl transition-colors"
        >
          Book Another Appointment
        </button>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto space-y-6">
      {/* Multi-step progress indicator */}
      <div className="bg-slate-900/60 backdrop-blur-md border border-slate-800 rounded-2xl p-4 shadow-xl">
        <div className="flex items-center justify-between text-xs font-medium text-slate-400 px-2">
          {[
            { step: 1, label: 'Service' },
            { step: 2, label: 'Specialist' },
            { step: 3, label: 'Date & Time' },
            { step: 4, label: 'Details' },
            { step: 5, label: 'Confirm' },
          ].map((s) => (
            <div key={s.step} className="flex items-center gap-2">
              <div
                className={`w-8 h-8 rounded-full flex items-center justify-center font-bold text-xs transition-all ${
                  currentStep === s.step
                    ? 'bg-violet-600 text-white shadow-lg shadow-violet-900/50 ring-2 ring-violet-500/30'
                    : currentStep > s.step
                    ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/40'
                    : 'bg-slate-800 text-slate-500'
                }`}
              >
                {currentStep > s.step ? <Check className="w-4 h-4" /> : s.step}
              </div>
              <span className={`hidden md:inline ${currentStep === s.step ? 'text-white font-semibold' : ''}`}>
                {s.label}
              </span>
            </div>
          ))}
        </div>
      </div>

      {/* Step Content Card */}
      <div className="bg-slate-900/70 border border-slate-800 rounded-3xl p-6 md:p-8 shadow-2xl backdrop-blur-xl">
        {/* Step 1: Select Service */}
        {currentStep === 1 && (
          <div className="space-y-6 animate-fade-in">
            <div>
              <h3 className="text-xl font-bold text-white tracking-tight">Select a Service</h3>
              <p className="text-xs text-slate-400 mt-1">Choose the service you would like to book today.</p>
            </div>

            <div className="space-y-3">
              {services.map((service) => {
                const isSelected = selectedService?.id === service.id;
                return (
                  <div
                    key={service.id}
                    onClick={() => setSelectedService(service)}
                    className={`p-5 rounded-2xl border transition-all cursor-pointer flex flex-col md:flex-row md:items-center justify-between gap-4 ${
                      isSelected
                        ? 'bg-violet-950/30 border-violet-500/60 shadow-lg shadow-violet-950/50 ring-1 ring-violet-500/40'
                        : 'bg-slate-950/40 border-slate-800/80 hover:border-slate-700/80 hover:bg-slate-800/30'
                    }`}
                  >
                    <div className="space-y-1">
                      <div className="flex items-center gap-3">
                        <h4 className="font-semibold text-white text-base">{service.name}</h4>
                        {service.requires_deposit && (
                          <span className="text-[10px] bg-amber-500/10 text-amber-400 border border-amber-500/30 px-2 py-0.5 rounded-full font-medium">
                            Deposit Required
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-slate-400 line-clamp-2 max-w-xl">{service.description}</p>
                      <div className="flex items-center gap-4 text-xs font-mono text-slate-400 pt-1">
                        <span className="flex items-center gap-1">
                          <Clock className="w-3.5 h-3.5 text-violet-400" />
                          {service.duration_minutes} min
                        </span>
                      </div>
                    </div>

                    <div className="text-right flex md:flex-col items-center md:items-end justify-between border-t md:border-t-0 border-slate-800/60 pt-3 md:pt-0">
                      <div className="text-lg font-bold text-emerald-400">
                        {formatPrice(service.price_cents, service.currency)}
                      </div>
                      <div className={`w-5 h-5 rounded-full border flex items-center justify-center mt-1 ${
                        isSelected ? 'border-violet-500 bg-violet-600 text-white' : 'border-slate-700'
                      }`}>
                        {isSelected && <Check className="w-3 h-3" />}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* Step 2: Select Staff */}
        {currentStep === 2 && (
          <div className="space-y-6 animate-fade-in">
            <div>
              <h3 className="text-xl font-bold text-white tracking-tight">Select Specialist</h3>
              <p className="text-xs text-slate-400 mt-1">Pick your preferred staff member or choose anyone available.</p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Option: Any Available */}
              <div
                onClick={() => setSelectedStaff(null)}
                className={`p-5 rounded-2xl border transition-all cursor-pointer ${
                  selectedStaff === null
                    ? 'bg-violet-950/30 border-violet-500/60 ring-1 ring-violet-500/40 shadow-lg'
                    : 'bg-slate-950/40 border-slate-800/80 hover:bg-slate-800/30'
                }`}
              >
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-2xl bg-gradient-to-br from-violet-600 to-indigo-600 flex items-center justify-center text-white font-bold shadow-md">
                    <Sparkles className="w-6 h-6" />
                  </div>
                  <div>
                    <h4 className="font-semibold text-white">Any Available Specialist</h4>
                    <p className="text-xs text-slate-400">Maximum flexibility for your schedule</p>
                  </div>
                </div>
              </div>

              {/* Specific Staff */}
              {staffMembers.map((staff) => {
                const isSelected = selectedStaff?.id === staff.id;
                return (
                  <div
                    key={staff.id}
                    onClick={() => setSelectedStaff(staff)}
                    className={`p-5 rounded-2xl border transition-all cursor-pointer ${
                      isSelected
                        ? 'bg-violet-950/30 border-violet-500/60 ring-1 ring-violet-500/40 shadow-lg'
                        : 'bg-slate-950/40 border-slate-800/80 hover:bg-slate-800/30'
                    }`}
                  >
                    <div className="flex items-center gap-4">
                      <div className="w-12 h-12 rounded-2xl bg-slate-800 border border-slate-700 flex items-center justify-center text-violet-400 font-bold text-lg">
                        {staff.name.charAt(0)}
                      </div>
                      <div>
                        <h4 className="font-semibold text-white">{staff.name}</h4>
                        <p className="text-xs text-slate-400">{staff.title}</p>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* Step 3: Date & Slot Picker */}
        {currentStep === 3 && (
          <div className="space-y-6 animate-fade-in">
            <div>
              <h3 className="text-xl font-bold text-white tracking-tight">Select Date & Time</h3>
              <p className="text-xs text-slate-400 mt-1">Available time slots are generated based on real-time schedule checks.</p>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-2">
                  Select Date
                </label>
                <input
                  type="date"
                  value={selectedDate}
                  min={new Date().toISOString().split('T')[0]}
                  onChange={(e) => setSelectedDate(e.target.value)}
                  className="bg-slate-950 border border-slate-800 rounded-xl px-4 py-2.5 text-slate-200 text-sm focus:outline-none focus:border-violet-500 transition-colors"
                />
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-2">
                  Available Slots
                </label>
                <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 gap-3">
                  {mockSlots.map((slot) => {
                    const isSelected = selectedSlot === slot;
                    return (
                      <button
                        key={slot}
                        type="button"
                        onClick={() => setSelectedSlot(slot)}
                        className={`py-3 px-2 rounded-xl text-xs font-mono font-semibold transition-all border ${
                          isSelected
                            ? 'bg-gradient-to-r from-violet-600 to-indigo-600 text-white border-violet-400 shadow-md shadow-violet-900/50'
                            : 'bg-slate-950/60 border-slate-800 text-slate-300 hover:border-slate-700 hover:bg-slate-800/40'
                        }`}
                      >
                        {slot}
                      </button>
                    );
                  })}
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Step 4: Customer Details Form */}
        {currentStep === 4 && (
          <div className="space-y-6 animate-fade-in">
            <div>
              <h3 className="text-xl font-bold text-white tracking-tight">Your Details</h3>
              <p className="text-xs text-slate-400 mt-1">Provide your contact information so we can send confirmation details.</p>
            </div>

            <div className="space-y-4">
              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Full Name *</label>
                <input
                  type="text"
                  required
                  placeholder="John Doe"
                  value={customerName}
                  onChange={(e) => setCustomerName(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-sm focus:outline-none focus:border-violet-500 transition-colors"
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">Email Address *</label>
                  <input
                    type="email"
                    required
                    placeholder="john@example.com"
                    value={customerEmail}
                    onChange={(e) => setCustomerEmail(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-sm focus:outline-none focus:border-violet-500 transition-colors"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-slate-300 mb-1">Phone Number *</label>
                  <input
                    type="tel"
                    required
                    placeholder="+1 (555) 000-0000"
                    value={customerPhone}
                    onChange={(e) => setCustomerPhone(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-sm focus:outline-none focus:border-violet-500 transition-colors"
                  />
                </div>
              </div>

              <div>
                <label className="block text-xs font-semibold text-slate-300 mb-1">Special Instructions / Notes</label>
                <textarea
                  rows={3}
                  placeholder="Any allergies, preferences, or specific requests..."
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl px-4 py-3 text-slate-200 text-sm focus:outline-none focus:border-violet-500 transition-colors resize-none"
                />
              </div>
            </div>
          </div>
        )}

        {/* Step 5: Confirmation & Summary */}
        {currentStep === 5 && (
          <form onSubmit={handleSubmitBooking} className="space-y-6 animate-fade-in">
            <div>
              <h3 className="text-xl font-bold text-white tracking-tight">Review & Confirm</h3>
              <p className="text-xs text-slate-400 mt-1">Please double check your booking summary before submitting.</p>
            </div>

            <div className="bg-slate-950/60 border border-slate-800 rounded-2xl p-6 space-y-4">
              <div className="flex justify-between items-center pb-4 border-b border-slate-800">
                <div>
                  <h4 className="font-semibold text-white">{selectedService?.name}</h4>
                  <p className="text-xs text-slate-400 font-mono">{selectedService?.duration_minutes} Minutes</p>
                </div>
                <div className="text-lg font-bold text-emerald-400">
                  {formatPrice(selectedService?.price_cents || 0)}
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4 text-xs">
                <div>
                  <span className="text-slate-500 block">Date & Time</span>
                  <span className="text-slate-200 font-semibold">{selectedDate} at {selectedSlot}</span>
                </div>
                <div>
                  <span className="text-slate-500 block">Specialist</span>
                  <span className="text-slate-200 font-semibold">{selectedStaff ? selectedStaff.name : 'Any Specialist'}</span>
                </div>
                <div>
                  <span className="text-slate-500 block">Customer</span>
                  <span className="text-slate-200 font-semibold">{customerName}</span>
                </div>
                <div>
                  <span className="text-slate-500 block">Contact</span>
                  <span className="text-slate-200 font-semibold">{customerEmail}</span>
                </div>
              </div>

              {selectedService?.requires_deposit && (
                <div className="bg-amber-500/10 border border-amber-500/20 rounded-xl p-3 flex items-center gap-3 text-xs text-amber-300">
                  <CreditCard className="w-5 h-5 text-amber-400 shrink-0" />
                  <span>
                    This service requires a deposit of{' '}
                    <strong className="font-bold">{formatPrice(selectedService.deposit_amount_cents)}</strong> to guarantee your slot.
                  </span>
                </div>
              )}
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full bg-gradient-to-r from-violet-600 to-indigo-600 hover:from-violet-500 hover:to-indigo-500 text-white font-bold py-4 rounded-xl transition-all shadow-xl shadow-violet-900/40 disabled:opacity-50 flex items-center justify-center gap-2"
            >
              {isSubmitting ? (
                <div className="w-5 h-5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
              ) : (
                <>
                  <ShieldCheck className="w-5 h-5" />
                  Confirm & Complete Booking
                </>
              )}
            </button>
          </form>
        )}

        {/* Wizard Footer Navigation */}
        <div className="flex items-center justify-between pt-6 mt-6 border-t border-slate-800/80">
          {currentStep > 1 ? (
            <button
              type="button"
              onClick={handlePrev}
              className="flex items-center gap-2 text-xs font-semibold text-slate-400 hover:text-white transition-colors py-2 px-4 rounded-xl hover:bg-slate-800"
            >
              <ChevronLeft className="w-4 h-4" />
              Back
            </button>
          ) : (
            <div />
          )}

          {currentStep < 5 && (
            <button
              type="button"
              onClick={handleNext}
              disabled={
                (currentStep === 1 && !selectedService) ||
                (currentStep === 3 && !selectedSlot)
              }
              className="flex items-center gap-2 bg-violet-600 hover:bg-violet-500 disabled:opacity-50 text-white text-xs font-bold px-6 py-3 rounded-xl transition-all shadow-lg shadow-violet-900/30"
            >
              Next Step
              <ChevronRight className="w-4 h-4" />
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
