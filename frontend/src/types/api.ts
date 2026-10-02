/**
 * API Response Types
 * These types must mirror the backend Go response envelope.
 * Source of truth: backend/internal/handler/v1/response.go
 */

// --- Response Envelope -------------------------------------------------------

export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface ResponseMeta {
  request_id: string;
  timestamp: string;
  pagination?: PaginationMeta;
}

export interface ApiResponse<T> {
  data: T;
  meta: ResponseMeta;
  success?: boolean;
  error?: ApiErrorBody;
}

export interface ApiListResponse<T> {
  data: T[];
  meta: ResponseMeta;
}

export interface FieldError {
  field: string;
  message: string;
}

export interface ApiErrorBody {
  code: string;
  message: string;
  details?: FieldError[];
}

export interface ApiErrorResponse {
  error: ApiErrorBody;
  meta: ResponseMeta;
}

// --- Domain Types ------------------------------------------------------------

// Organization
export interface Organization {
  id: string;
  name: string;
  slug: string;
  business_type: string;
  email: string;
  phone?: string;
  website?: string;
  logo_url?: string;
  description?: string;
  timezone: string;
  currency: string;
  country: string;
  status: 'active' | 'suspended' | 'inactive';
  settings: OrganizationSettings;
  owner_id: string;
  created_at: string;
  updated_at: string;
}

export interface OrganizationSettings {
  allow_online_booking: boolean;
  require_payment_upfront: boolean;
  auto_confirm_bookings: boolean;
  allow_guest_booking: boolean;
  min_advance_booking_hours: number;
  max_advance_days: number;
  cancellation_notice_hours: number;
  slot_granularity_minutes: number;
  send_confirmation_email: boolean;
  send_reminder_email: boolean;
  reminder_hours_before: number;
  send_sms_reminder: boolean;
  send_whatsapp_reminder: boolean;
}

// User
export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
  avatar_url?: string;
  is_email_verified: boolean;
  created_at: string;
  updated_at: string;
}

// Auth
export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: string;
  token_type: 'Bearer';
}

export interface AuthUser extends User {
  current_org?: Organization;
  role?: string;
}

// Staff
export interface Staff {
  id: string;
  organization_id: string;
  user_id?: string;
  first_name: string;
  last_name: string;
  name?: string;
  email?: string;
  phone?: string;
  avatar_url?: string;
  title?: string;
  bio?: string;
  is_active: boolean;
  accepts_online: boolean;
  display_order: number;
  created_at: string;
  updated_at: string;
}

// Service
export interface ServiceCategory {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  color?: string;
  display_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface Service {
  id: string;
  organization_id?: string;
  tenant_id?: string;
  category_id?: string;
  category?: ServiceCategory;
  name: string;
  slug?: string;
  description?: string;
  status?: 'active' | 'inactive' | 'archived';
  is_active?: boolean;
  duration_minutes: number;
  buffer_before_minutes?: number;
  buffer_after_minutes?: number;
  buffer_time_before_minutes?: number;
  buffer_time_after_minutes?: number;
  price_cents: number;
  currency: string;
  requires_deposit?: boolean;
  deposit_amount_cents?: number;
  max_capacity?: number;
  requires_resource?: boolean;
  color?: string;
  image_url?: string;
  display_order?: number;
  is_public?: boolean;
  created_at: string;
  updated_at: string;
}

// Location
export interface Location {
  id: string;
  organization_id: string;
  name: string;
  description?: string;
  phone?: string;
  email?: string;
  address_line1: string;
  address_line2?: string;
  city: string;
  state?: string;
  postal_code?: string;
  country: string;
  latitude?: number;
  longitude?: number;
  timezone?: string;
  status?: 'active' | 'inactive';
  is_active?: boolean;
  is_default: boolean;
  display_order: number;
  created_at: string;
  updated_at: string;
}

// Customer
export type CustomerStatus = 'active' | 'inactive' | 'vip' | 'blocked' | 'ACTIVE' | 'INACTIVE' | 'VIP' | 'BLOCKED';
export type CustomerSource = 'online' | 'manual' | 'import' | 'referral' | 'ONLINE_BOOKING' | 'WALK_IN' | 'MANUAL';

export interface Customer {
  id: string;
  organization_id: string;
  user_id?: string;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  notes?: string;
  status?: CustomerStatus;
  source?: CustomerSource;
  total_appointments: number;
  completed_appointments: number;
  no_show_count: number;
  last_appointment_at?: string;
  total_spent_cents: number;
  created_at: string;
  updated_at: string;
}

// Appointment
export type AppointmentStatus =
  | 'pending'
  | 'confirmed'
  | 'rescheduled'
  | 'completed'
  | 'cancelled'
  | 'no_show'
  | 'PENDING'
  | 'CONFIRMED'
  | 'RESCHEDULED'
  | 'COMPLETED'
  | 'CANCELLED'
  | 'NO_SHOW'
  | 'PENDING_DEPOSIT';

export interface Appointment {
  id: string;
  organization_id?: string;
  tenant_id?: string;
  location_id: string;
  service_id: string;
  staff_id: string;
  customer_id?: string;
  resource_id?: string;
  start_time: string;
  end_time: string;
  timezone: string;
  status: AppointmentStatus;
  price_cents: number;
  currency: string;
  guest_name?: string;
  guest_email?: string;
  guest_phone?: string;
  notes?: string;
  internal_notes?: string;
  cancel_reason?: string;
  source: 'online' | 'manual' | 'api' | 'ONLINE_BOOKING';
  payment_id?: string;
  created_by?: string;
  created_at: string;
  updated_at: string;
  // Joined fields (returned by API with full details)
  customer?: Customer;
  staff?: Staff;
  service?: Service;
  location?: Location;
}

export interface AppointmentStatusHistory {
  id: string;
  appointment_id: string;
  from_status: AppointmentStatus;
  to_status: AppointmentStatus;
  reason?: string;
  changed_by: string;
  changed_at: string;
}

// Availability
export interface TimeSlot {
  start_time: string; // ISO 8601
  end_time: string;   // ISO 8601
  staff_id: string;
  available: boolean;
}

export interface AvailabilityResult {
  date: string;
  timezone: string;
  service_id: string;
  slots: TimeSlot[];
}

// Payment
export type PaymentStatus =
  | 'pending'
  | 'paid'
  | 'failed'
  | 'refunded'
  | 'partially_refunded'
  | 'expired'
  | 'cancelled';

export interface Payment {
  id: string;
  organization_id: string;
  appointment_id: string;
  customer_id?: string;
  status: PaymentStatus;
  provider: 'stripe' | 'midtrans' | 'xendit' | 'manual';
  amount_cents: number;
  currency: string;
  refunded_amount_cents: number;
  external_id?: string;
  payment_url?: string;
  receipt_url?: string;
  notes?: string;
  paid_at?: string;
  expires_at?: string;
  created_at: string;
  updated_at: string;
}

// --- Request Types -----------------------------------------------------------

export interface CreateAppointmentRequest {
  location_id: string;
  service_id: string;
  staff_id: string;
  customer_id?: string;
  resource_id?: string;
  start_time: string;
  timezone: string;
  guest_name?: string;
  guest_email?: string;
  guest_phone?: string;
  notes?: string;
}

export interface UpdateAppointmentStatusRequest {
  status: AppointmentStatus;
  reason?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  phone?: string;
}

export interface ListParams {
  page?: number;
  per_page?: number;
  search?: string;
}

export interface AppointmentListParams extends ListParams {
  status?: AppointmentStatus;
  staff_id?: string;
  location_id?: string;
  service_id?: string;
  date_from?: string;
  date_to?: string;
}

// --- Resource Types ----------------------------------------------------------

export type ResourceType = 'ROOM' | 'EQUIPMENT' | 'CHAIR' | 'STUDIO' | 'FACILITY' | 'OTHER';
export type ResourceStatus = 'ACTIVE' | 'INACTIVE' | 'MAINTENANCE';

export interface Resource {
  id: string;
  organization_id: string;
  location_id?: string;
  name: string;
  type: ResourceType;
  capacity: number;
  status: ResourceStatus;
  description?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateResourceRequest {
  location_id?: string;
  name: string;
  type: ResourceType;
  capacity: number;
  status: ResourceStatus;
  description?: string;
}

export interface UpdateResourceRequest {
  location_id?: string;
  name: string;
  type: ResourceType;
  capacity: number;
  status: ResourceStatus;
  description?: string;
}

export interface Refund {
  id: string;
  payment_id: string;
  amount_cents: number;
  currency: string;
  reason?: string;
  status: 'pending' | 'succeeded' | 'failed';
  created_at: string;
}

export interface BusinessHours {
  day_of_week: number;
  start_time: string;
  end_time: string;
  is_closed: boolean;
}

export interface StaffSchedule {
  id: string;
  staff_id: string;
  day_of_week: number;
  start_time: string;
  end_time: string;
}

export interface StaffTimeOff {
  id: string;
  staff_id: string;
  start_at: string;
  end_at: string;
  reason?: string;
}


