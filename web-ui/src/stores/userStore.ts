import { create } from 'zustand';
import { userApi } from '@/api/user';
import type { User, Plan, Node, Usage } from '@/types';

interface UserState {
  profile: User | null;
  plan: Plan | null;
  nodes: Node[];
  usage: Usage | null;
  isLoading: boolean;
  error: string | null;
  fetchProfile: () => Promise<void>;
  fetchPlan: () => Promise<void>;
  fetchNodes: () => Promise<void>;
  fetchUsage: () => Promise<void>;
  fetchAll: () => Promise<void>;
}

export const useUserStore = create<UserState>((set) => ({
  profile: null,
  plan: null,
  nodes: [],
  usage: null,
  isLoading: false,
  error: null,

  fetchProfile: async () => {
    try {
      const profile = await userApi.getProfile();
      set({ profile });
    } catch (error: any) {
      const message = error.response?.data?.message || 'Failed to fetch profile';
      set({ error: message });
    }
  },

  fetchPlan: async () => {
    try {
      const plan = await userApi.getPlan();
      set({ plan });
    } catch (error: any) {
      const message = error.response?.data?.message || 'Failed to fetch plan';
      set({ error: message });
    }
  },

  fetchNodes: async () => {
    try {
      const nodes = await userApi.getNodes();
      set({ nodes });
    } catch (error: any) {
      const message = error.response?.data?.message || 'Failed to fetch nodes';
      set({ error: message });
    }
  },

  fetchUsage: async () => {
    try {
      const usage = await userApi.getUsage();
      set({ usage });
    } catch (error: any) {
      const message = error.response?.data?.message || 'Failed to fetch usage';
      set({ error: message });
    }
  },

  fetchAll: async () => {
    set({ isLoading: true, error: null });
    try {
      await Promise.all([
        useUserStore.getState().fetchProfile(),
        useUserStore.getState().fetchPlan(),
        useUserStore.getState().fetchNodes(),
        useUserStore.getState().fetchUsage(),
      ]);
    } finally {
      set({ isLoading: false });
    }
  },
}));
