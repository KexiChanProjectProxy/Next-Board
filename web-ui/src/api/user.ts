import apiClient from './client';
import type { User, Plan, Node, Usage, TelegramLinkResponse } from '@/types';

export const userApi = {
  getProfile: async (): Promise<User> => {
    const response = await apiClient.get<{ user: User }>('/api/v1/me');
    return response.data.user;
  },

  getPlan: async (): Promise<Plan | null> => {
    const response = await apiClient.get<{ plan: Plan | null }>('/api/v1/me/plan');
    return response.data.plan;
  },

  getNodes: async (): Promise<Node[]> => {
    const response = await apiClient.get<{ nodes: Node[] }>('/api/v1/me/nodes');
    return response.data.nodes;
  },

  getUsage: async (): Promise<Usage> => {
    const response = await apiClient.get<{ usage: Usage }>('/api/v1/me/usage');
    return response.data.usage;
  },

  generateTelegramLink: async (): Promise<string> => {
    const response = await apiClient.post<TelegramLinkResponse>('/api/v1/me/telegram/link');
    return response.data.link_token;
  },
};
