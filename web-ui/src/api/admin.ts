import apiClient from './client';
import type {
  User,
  Node,
  Plan,
  Label,
  PaginatedResponse,
  CreateUserRequest,
  UpdateUserRequest,
  CreateNodeRequest,
  UpdateNodeRequest,
  CreatePlanRequest,
  UpdatePlanRequest,
  CreateLabelRequest,
  UpdateLabelRequest,
} from '@/types';

export const adminApi = {
  // Users
  getUsers: async (page = 1, limit = 20): Promise<PaginatedResponse<User>> => {
    const response = await apiClient.get<PaginatedResponse<User>>('/api/v1/admin/users', {
      params: { page, limit },
    });
    return response.data;
  },

  createUser: async (data: CreateUserRequest): Promise<User> => {
    const response = await apiClient.post<{ user: User }>('/api/v1/admin/users', data);
    return response.data.user;
  },

  updateUser: async (id: number, data: UpdateUserRequest): Promise<User> => {
    const response = await apiClient.put<{ user: User }>(`/api/v1/admin/users/${id}`, data);
    return response.data.user;
  },

  deleteUser: async (id: number): Promise<void> => {
    await apiClient.delete(`/api/v1/admin/users/${id}`);
  },

  // Nodes
  getNodes: async (page = 1, limit = 20): Promise<PaginatedResponse<Node>> => {
    const response = await apiClient.get<PaginatedResponse<Node>>('/api/v1/admin/nodes', {
      params: { page, limit },
    });
    return response.data;
  },

  createNode: async (data: CreateNodeRequest): Promise<Node> => {
    const response = await apiClient.post<{ node: Node }>('/api/v1/admin/nodes', data);
    return response.data.node;
  },

  updateNode: async (id: number, data: UpdateNodeRequest): Promise<Node> => {
    const response = await apiClient.put<{ node: Node }>(`/api/v1/admin/nodes/${id}`, data);
    return response.data.node;
  },

  deleteNode: async (id: number): Promise<void> => {
    await apiClient.delete(`/api/v1/admin/nodes/${id}`);
  },

  assignNodeLabels: async (nodeId: number, labelIds: number[]): Promise<void> => {
    await apiClient.post(`/api/v1/admin/nodes/${nodeId}/labels`, { label_ids: labelIds });
  },

  // Plans
  getPlans: async (page = 1, limit = 20): Promise<PaginatedResponse<Plan>> => {
    const response = await apiClient.get<PaginatedResponse<Plan>>('/api/v1/admin/plans', {
      params: { page, limit },
    });
    return response.data;
  },

  createPlan: async (data: CreatePlanRequest): Promise<Plan> => {
    const response = await apiClient.post<{ plan: Plan }>('/api/v1/admin/plans', data);
    return response.data.plan;
  },

  updatePlan: async (id: number, data: UpdatePlanRequest): Promise<Plan> => {
    const response = await apiClient.put<{ plan: Plan }>(`/api/v1/admin/plans/${id}`, data);
    return response.data.plan;
  },

  deletePlan: async (id: number): Promise<void> => {
    await apiClient.delete(`/api/v1/admin/plans/${id}`);
  },

  assignPlanLabels: async (planId: number, labelIds: number[]): Promise<void> => {
    await apiClient.post(`/api/v1/admin/plans/${planId}/labels`, { label_ids: labelIds });
  },

  // Labels
  getLabels: async (page = 1, limit = 100): Promise<PaginatedResponse<Label>> => {
    const response = await apiClient.get<PaginatedResponse<Label>>('/api/v1/admin/labels', {
      params: { page, limit },
    });
    return response.data;
  },

  createLabel: async (data: CreateLabelRequest): Promise<Label> => {
    const response = await apiClient.post<{ label: Label }>('/api/v1/admin/labels', data);
    return response.data.label;
  },

  updateLabel: async (id: number, data: UpdateLabelRequest): Promise<Label> => {
    const response = await apiClient.put<{ label: Label }>(`/api/v1/admin/labels/${id}`, data);
    return response.data.label;
  },

  deleteLabel: async (id: number): Promise<void> => {
    await apiClient.delete(`/api/v1/admin/labels/${id}`);
  },
};
