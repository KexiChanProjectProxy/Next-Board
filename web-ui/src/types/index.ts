// Core API Models
export interface User {
  id: number;
  email: string;
  role: 'admin' | 'user';
  plan_id: number | null;
  telegram_chat_id: number | null;
  telegram_linked_at: string | null;
  created_at: string;
}

export interface Plan {
  id: number;
  name: string;
  quota_bytes: number;
  reset_period: 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly';
  base_multiplier: number;
  labels: Label[];
  created_at: string;
}

export interface Node {
  id: number;
  name: string;
  node_type: string;
  host: string;
  port: number;
  node_multiplier: number;
  status: 'active' | 'inactive';
  labels: Label[];
  protocol_config?: Record<string, any>;
}

export interface Label {
  id: number;
  name: string;
  description: string;
  multiplier: number;
}

export interface Usage {
  real_bytes_up: number;
  real_bytes_down: number;
  billable_bytes_up: number;
  billable_bytes_down: number;
  period_start: string;
  period_end: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: {
    total: number;
    page: number;
    limit: number;
    pages: number;
  };
}

// API Request/Response Types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  access_token: string;
  refresh_token: string;
}

export interface RefreshRequest {
  refresh_token: string;
}

export interface RefreshResponse {
  access_token: string;
}

export interface CreateUserRequest {
  email: string;
  password: string;
  plan_id?: number;
  role?: 'admin' | 'user';
}

export interface UpdateUserRequest {
  email?: string;
  plan_id?: number;
  banned?: boolean;
}

export interface CreateNodeRequest {
  name: string;
  node_type: string;
  host: string;
  port: number;
  protocol_config?: Record<string, any>;
  node_multiplier: number;
  label_ids?: number[];
}

export interface UpdateNodeRequest {
  name?: string;
  node_type?: string;
  host?: string;
  port?: number;
  protocol_config?: Record<string, any>;
  node_multiplier?: number;
  status?: 'active' | 'inactive';
}

export interface CreatePlanRequest {
  name: string;
  quota_bytes: number;
  reset_period: 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly';
  base_multiplier: number;
  label_ids?: number[];
}

export interface UpdatePlanRequest {
  name?: string;
  quota_bytes?: number;
  reset_period?: 'none' | 'daily' | 'weekly' | 'monthly' | 'yearly';
  base_multiplier?: number;
}

export interface CreateLabelRequest {
  name: string;
  description: string;
}

export interface UpdateLabelRequest {
  name?: string;
  description?: string;
}

export interface TelegramLinkResponse {
  link_token: string;
}

export interface APIError {
  code: string;
  message: string;
}
