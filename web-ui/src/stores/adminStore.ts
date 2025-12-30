import { create } from 'zustand';
import { adminApi } from '@/api/admin';
import type { User, Node, Plan, Label, PaginatedResponse } from '@/types';

interface AdminState {
  users: PaginatedResponse<User> | null;
  nodes: PaginatedResponse<Node> | null;
  plans: PaginatedResponse<Plan> | null;
  labels: PaginatedResponse<Label> | null;
  isLoading: boolean;
  error: string | null;
  fetchUsers: (page?: number, limit?: number) => Promise<void>;
  fetchNodes: (page?: number, limit?: number) => Promise<void>;
  fetchPlans: (page?: number, limit?: number) => Promise<void>;
  fetchLabels: (page?: number, limit?: number) => Promise<void>;
}

export const useAdminStore = create<AdminState>((set) => ({
  users: null,
  nodes: null,
  plans: null,
  labels: null,
  isLoading: false,
  error: null,

  fetchUsers: async (page = 1, limit = 20) => {
    set({ isLoading: true, error: null });
    try {
      const users = await adminApi.getUsers(page, limit);
      set({ users, isLoading: false });
    } catch (error: any) {
      const message = error.response?.data?.message || 'Failed to fetch users';
      set({ error: message, isLoading: false });
    }
  },

  fetchNodes: async (page = 1, limit = 20) => {
    set({ isLoading: true, error: null });
    try {
      const nodes = await adminApi.getNodes(page, limit);
      set({ nodes, isLoading: false });
    } catch (error: any) {
      const message = error.response?.data?.message || 'Failed to fetch nodes';
      set({ error: message, isLoading: false });
    }
  },

  fetchPlans: async (page = 1, limit = 20) => {
    set({ isLoading: true, error: null });
    try {
      const plans = await adminApi.getPlans(page, limit);
      set({ plans, isLoading: false });
    } catch (error: any) {
      const message = error.response?.data?.message || 'Failed to fetch plans';
      set({ error: message, isLoading: false });
    }
  },

  fetchLabels: async (page = 1, limit = 100) => {
    set({ isLoading: true, error: null });
    try {
      const labels = await adminApi.getLabels(page, limit);
      set({ labels, isLoading: false });
    } catch (error: any) {
      const message = error.response?.data?.message || 'Failed to fetch labels';
      set({ error: message, isLoading: false });
    }
  },
}));
