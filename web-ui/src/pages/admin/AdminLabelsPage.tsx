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
import { Plus, Edit, Trash2, ChevronLeft, ChevronRight, Tag } from 'lucide-react';
import type { Label as LabelType, CreateLabelRequest, UpdateLabelRequest } from '@/types';

export function AdminLabelsPage() {
  const { labels, fetchLabels, isLoading } = useAdminStore();
  const { toast } = useToast();
  const [currentPage, setCurrentPage] = useState(1);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedLabel, setSelectedLabel] = useState<LabelType | null>(null);

  // Form states
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchLabels(currentPage);
  }, [currentPage]);

  const handleCreateLabel = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);

    try {
      const data: CreateLabelRequest = {
        name,
        description,
      };

      await adminApi.createLabel(data);
      toast({
        title: 'Success',
        description: 'Label created successfully',
      });
      setCreateDialogOpen(false);
      resetForm();
      fetchLabels(currentPage);
    } catch (error: any) {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: error.response?.data?.error?.message || 'Failed to create label',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdateLabel = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedLabel) return;
    setSubmitting(true);

    try {
      const data: UpdateLabelRequest = {};

      if (name && name !== selectedLabel.name) data.name = name;
      if (description && description !== selectedLabel.description) data.description = description;

      await adminApi.updateLabel(selectedLabel.id, data);
      toast({
        title: 'Success',
        description: 'Label updated successfully',
      });
      setEditDialogOpen(false);
      resetForm();
      fetchLabels(currentPage);
    } catch (error: any) {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: error.response?.data?.error?.message || 'Failed to update label',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteLabel = async () => {
    if (!selectedLabel) return;
    setSubmitting(true);

    try {
      await adminApi.deleteLabel(selectedLabel.id);
      toast({
        title: 'Success',
        description: 'Label deleted successfully',
      });
      setDeleteDialogOpen(false);
      setSelectedLabel(null);
      fetchLabels(currentPage);
    } catch (error: any) {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: error.response?.data?.error?.message || 'Failed to delete label',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const openEditDialog = (label: LabelType) => {
    setSelectedLabel(label);
    setName(label.name);
    setDescription(label.description);
    setEditDialogOpen(true);
  };

  const openDeleteDialog = (label: LabelType) => {
    setSelectedLabel(label);
    setDeleteDialogOpen(true);
  };

  const resetForm = () => {
    setName('');
    setDescription('');
    setSelectedLabel(null);
  };

  if (isLoading) {
    return <div className="text-center py-12">Loading...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Label Management</h1>
          <p className="text-gray-600 mt-1">Manage labels for organizing nodes and plans</p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create Label
        </Button>
      </div>

      {/* Labels Table */}
      <Card>
        <CardHeader>
          <CardTitle>Labels</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="text-left p-3">ID</th>
                  <th className="text-left p-3">Name</th>
                  <th className="text-left p-3">Description</th>
                  <th className="text-left p-3">Multiplier</th>
                  <th className="text-right p-3">Actions</th>
                </tr>
              </thead>
              <tbody>
                {labels?.data.map((label) => (
                  <tr key={label.id} className="border-b hover:bg-gray-50">
                    <td className="p-3">{label.id}</td>
                    <td className="p-3 font-medium">
                      <div className="flex items-center gap-2">
                        <Tag className="h-4 w-4 text-gray-500" />
                        <Badge variant="outline">{label.name}</Badge>
                      </div>
                    </td>
                    <td className="p-3 text-sm text-muted-foreground">
                      {label.description || 'No description'}
                    </td>
                    <td className="p-3">{label.multiplier}x</td>
                    <td className="p-3">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openEditDialog(label)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openDeleteDialog(label)}
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
          {labels && labels.pagination.pages > 1 && (
            <div className="flex items-center justify-between mt-4 pt-4 border-t">
              <div className="text-sm text-muted-foreground">
                Page {labels.pagination.page} of {labels.pagination.pages} ({labels.pagination.total} total)
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
                  onClick={() => setCurrentPage(p => Math.min(labels.pagination.pages, p + 1))}
                  disabled={currentPage === labels.pagination.pages}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Label Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Label</DialogTitle>
            <DialogDescription>Add a new label for organizing resources</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreateLabel}>
            <div className="space-y-4 py-4">
              <div>
                <Label htmlFor="create-name">Label Name</Label>
                <Input
                  id="create-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                  placeholder="e.g., Premium, US Region"
                  required
                />
              </div>

              <div>
                <Label htmlFor="create-description">Description</Label>
                <textarea
                  id="create-description"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full p-2 border rounded-md min-h-[80px]"
                  placeholder="Optional description for this label"
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setCreateDialogOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? 'Creating...' : 'Create Label'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Edit Label Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Label</DialogTitle>
            <DialogDescription>Update label information</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleUpdateLabel}>
            <div className="space-y-4 py-4">
              <div>
                <Label htmlFor="edit-name">Label Name</Label>
                <Input
                  id="edit-name"
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              </div>

              <div>
                <Label htmlFor="edit-description">Description</Label>
                <textarea
                  id="edit-description"
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  className="w-full p-2 border rounded-md min-h-[80px]"
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setEditDialogOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? 'Updating...' : 'Update Label'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Label</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete label "{selectedLabel?.name}"? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteLabel} disabled={submitting}>
              {submitting ? 'Deleting...' : 'Delete Label'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
