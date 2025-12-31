import { useEffect, useState } from 'react';
import { useAdminStore } from '@/stores/adminStore';
import { adminApi } from '@/api/admin';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { useToast } from '@/components/ui/use-toast';
import { formatBytes } from '@/utils/format';
import { Plus, Edit, Trash2, ChevronLeft, ChevronRight, Package } from 'lucide-react';
import type { Plan, CreatePlanRequest, UpdatePlanRequest } from '@/types';

export function AdminPlansPage() {
  const { plans, fetchPlans, isLoading } = useAdminStore();
  const { toast } = useToast();
  const [currentPage, setCurrentPage] = useState(1);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedPlan, setSelectedPlan] = useState<Plan | null>(null);

  // Form states
  const [name, setName] = useState('');
  const [quotaGB, setQuotaGB] = useState('');
  const [resetPeriod, setResetPeriod] = useState<'none' | 'daily' | 'weekly' | 'monthly' | 'yearly'>('monthly');
  const [baseMultiplier, setBaseMultiplier] = useState('1.0');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchPlans(currentPage);
  }, [currentPage]);

  const handleCreatePlan = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);

    try {
      const quotaBytes = parseFloat(quotaGB) * 1024 * 1024 * 1024; // Convert GB to bytes
      const data: CreatePlanRequest = {
        name,
        quota_bytes: quotaBytes,
        reset_period: resetPeriod,
        base_multiplier: parseFloat(baseMultiplier),
      };

      await adminApi.createPlan(data);
      toast({
        title: 'Success',
        description: 'Plan created successfully',
      });
      setCreateDialogOpen(false);
      resetForm();
      fetchPlans(currentPage);
    } catch (error: any) {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: error.response?.data?.error?.message || 'Failed to create plan',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdatePlan = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedPlan) return;
    setSubmitting(true);

    try {
      const data: UpdatePlanRequest = {};

      if (name && name !== selectedPlan.name) data.name = name;
      if (quotaGB) {
        const quotaBytes = parseFloat(quotaGB) * 1024 * 1024 * 1024;
        if (quotaBytes !== selectedPlan.quota_bytes) {
          data.quota_bytes = quotaBytes;
        }
      }
      if (resetPeriod !== selectedPlan.reset_period) data.reset_period = resetPeriod;
      if (baseMultiplier && parseFloat(baseMultiplier) !== selectedPlan.base_multiplier) {
        data.base_multiplier = parseFloat(baseMultiplier);
      }

      await adminApi.updatePlan(selectedPlan.id, data);
      toast({
        title: 'Success',
        description: 'Plan updated successfully',
      });
      setEditDialogOpen(false);
      resetForm();
      fetchPlans(currentPage);
    } catch (error: any) {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: error.response?.data?.error?.message || 'Failed to update plan',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeletePlan = async () => {
    if (!selectedPlan) return;
    setSubmitting(true);

    try {
      await adminApi.deletePlan(selectedPlan.id);
      toast({
        title: 'Success',
        description: 'Plan deleted successfully',
      });
      setDeleteDialogOpen(false);
      setSelectedPlan(null);
      fetchPlans(currentPage);
    } catch (error: any) {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: error.response?.data?.error?.message || 'Failed to delete plan',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const openEditDialog = (plan: Plan) => {
    setSelectedPlan(plan);
    setName(plan.name);
    setQuotaGB((plan.quota_bytes / (1024 * 1024 * 1024)).toString()); // Convert bytes to GB
    setResetPeriod(plan.reset_period);
    setBaseMultiplier(plan.base_multiplier.toString());
    setEditDialogOpen(true);
  };

  const openDeleteDialog = (plan: Plan) => {
    setSelectedPlan(plan);
    setDeleteDialogOpen(true);
  };

  const resetForm = () => {
    setName('');
    setQuotaGB('');
    setResetPeriod('monthly');
    setBaseMultiplier('1.0');
    setSelectedPlan(null);
  };

  if (isLoading) {
    return <div className="text-center py-12">Loading...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Plan Management</h1>
          <p className="text-gray-600 mt-1">Manage subscription plans and quotas</p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create Plan
        </Button>
      </div>

      {/* Plans Table */}
      <Card>
        <CardHeader>
          <CardTitle>Plans</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="text-left p-3">ID</th>
                  <th className="text-left p-3">Name</th>
                  <th className="text-left p-3">Quota</th>
                  <th className="text-left p-3">Reset Period</th>
                  <th className="text-left p-3">Multiplier</th>
                  <th className="text-left p-3">Labels</th>
                  <th className="text-right p-3">Actions</th>
                </tr>
              </thead>
              <tbody>
                {plans?.data.map((plan) => (
                  <tr key={plan.id} className="border-b hover:bg-gray-50">
                    <td className="p-3">{plan.id}</td>
                    <td className="p-3 font-medium">
                      <div className="flex items-center gap-2">
                        <Package className="h-4 w-4 text-gray-500" />
                        {plan.name}
                      </div>
                    </td>
                    <td className="p-3">
                      <Badge variant="outline">{formatBytes(plan.quota_bytes)}</Badge>
                    </td>
                    <td className="p-3">
                      <Badge variant="secondary">
                        {plan.reset_period === 'none' ? 'No Reset' : plan.reset_period}
                      </Badge>
                    </td>
                    <td className="p-3">{plan.base_multiplier}x</td>
                    <td className="p-3">
                      <div className="flex gap-1 flex-wrap">
                        {plan.labels?.length > 0 ? (
                          plan.labels.map((label) => (
                            <Badge key={label.id} variant="outline" className="text-xs">
                              {label.name}
                            </Badge>
                          ))
                        ) : (
                          <span className="text-sm text-muted-foreground">None</span>
                        )}
                      </div>
                    </td>
                    <td className="p-3">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openEditDialog(plan)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openDeleteDialog(plan)}
                        >
                          <Trash2 className="h-4 w-4 text-red-600" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {plans && plans.pagination.pages > 1 && (
            <div className="flex items-center justify-between mt-4 pt-4 border-t">
              <div className="text-sm text-muted-foreground">
                Page {plans.pagination.page} of {plans.pagination.pages} ({plans.pagination.total} total)
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                  disabled={currentPage === 1}
                >
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setCurrentPage(p => Math.min(plans.pagination.pages, p + 1))}
                  disabled={currentPage === plans.pagination.pages}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Plan Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Plan</DialogTitle>
            <DialogDescription>Add a new subscription plan</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreatePlan}>
            <div className="space-y-4 py-4">
              <div>
                <Label htmlFor="create-name">Plan Name</Label>
                <Input
                  id="create-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g., Standard Plan"
                  required
                />
              </div>

              <div>
                <Label htmlFor="create-quota">Quota (GB)</Label>
                <Input
                  id="create-quota"
                  type="number"
                  step="0.01"
                  value={quotaGB}
                  onChange={(e) => setQuotaGB(e.target.value)}
                  placeholder="e.g., 100"
                  required
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Total data transfer allowed per reset period
                </p>
              </div>

              <div>
                <Label htmlFor="create-reset">Reset Period</Label>
                <select
                  id="create-reset"
                  value={resetPeriod}
                  onChange={(e) => setResetPeriod(e.target.value as any)}
                  className="w-full p-2 border rounded-md"
                  required
                >
                  <option value="none">No Reset</option>
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="monthly">Monthly</option>
                  <option value="yearly">Yearly</option>
                </select>
              </div>

              <div>
                <Label htmlFor="create-multiplier">Base Multiplier</Label>
                <Input
                  id="create-multiplier"
                  type="number"
                  step="0.1"
                  value={baseMultiplier}
                  onChange={(e) => setBaseMultiplier(e.target.value)}
                  placeholder="e.g., 1.0"
                  required
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Traffic multiplier for this plan (1.0 = normal rate)
                </p>
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setCreateDialogOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? 'Creating...' : 'Create Plan'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Edit Plan Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Plan</DialogTitle>
            <DialogDescription>Update plan information</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleUpdatePlan}>
            <div className="space-y-4 py-4">
              <div>
                <Label htmlFor="edit-name">Plan Name</Label>
                <Input
                  id="edit-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>

              <div>
                <Label htmlFor="edit-quota">Quota (GB)</Label>
                <Input
                  id="edit-quota"
                  type="number"
                  step="0.01"
                  value={quotaGB}
                  onChange={(e) => setQuotaGB(e.target.value)}
                />
              </div>

              <div>
                <Label htmlFor="edit-reset">Reset Period</Label>
                <select
                  id="edit-reset"
                  value={resetPeriod}
                  onChange={(e) => setResetPeriod(e.target.value as any)}
                  className="w-full p-2 border rounded-md"
                >
                  <option value="none">No Reset</option>
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="monthly">Monthly</option>
                  <option value="yearly">Yearly</option>
                </select>
              </div>

              <div>
                <Label htmlFor="edit-multiplier">Base Multiplier</Label>
                <Input
                  id="edit-multiplier"
                  type="number"
                  step="0.1"
                  value={baseMultiplier}
                  onChange={(e) => setBaseMultiplier(e.target.value)}
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setEditDialogOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? 'Updating...' : 'Update Plan'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Plan</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete plan "{selectedPlan?.name}"? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeletePlan} disabled={submitting}>
              {submitting ? 'Deleting...' : 'Delete Plan'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
