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
import { Plus, Edit, Trash2, ChevronLeft, ChevronRight, Server } from 'lucide-react';
import type { Node, CreateNodeRequest, UpdateNodeRequest } from '@/types';

export function AdminNodesPage() {
  const { nodes, fetchNodes, isLoading } = useAdminStore();
  const { toast } = useToast();
  const [currentPage, setCurrentPage] = useState(1);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);

  // Form states
  const [name, setName] = useState('');
  const [nodeType, setNodeType] = useState('');
  const [host, setHost] = useState('');
  const [port, setPort] = useState('');
  const [nodeMultiplier, setNodeMultiplier] = useState('1.0');
  const [status, setStatus] = useState<'active' | 'inactive'>('active');
  const [protocolConfig, setProtocolConfig] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    fetchNodes(currentPage);
  }, [currentPage]);

  const handleCreateNode = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);

    try {
      const data: CreateNodeRequest = {
        name,
        node_type: nodeType,
        host,
        port: parseInt(port),
        node_multiplier: parseFloat(nodeMultiplier),
      };

      if (protocolConfig) {
        try {
          data.protocol_config = JSON.parse(protocolConfig);
        } catch {
          toast({
            variant: 'destructive',
            title: 'Error',
            description: 'Invalid JSON in protocol config',
          });
          setSubmitting(false);
          return;
        }
      }

      await adminApi.createNode(data);
      toast({
        title: 'Success',
        description: 'Node created successfully',
      });
      setCreateDialogOpen(false);
      resetForm();
      fetchNodes(currentPage);
    } catch (error: any) {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: error.response?.data?.error?.message || 'Failed to create node',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleUpdateNode = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedNode) return;
    setSubmitting(true);

    try {
      const data: UpdateNodeRequest = {};

      if (name && name !== selectedNode.name) data.name = name;
      if (nodeType && nodeType !== selectedNode.node_type) data.node_type = nodeType;
      if (host && host !== selectedNode.host) data.host = host;
      if (port && parseInt(port) !== selectedNode.port) data.port = parseInt(port);
      if (nodeMultiplier && parseFloat(nodeMultiplier) !== selectedNode.node_multiplier) {
        data.node_multiplier = parseFloat(nodeMultiplier);
      }
      if (status !== selectedNode.status) data.status = status;

      if (protocolConfig) {
        try {
          data.protocol_config = JSON.parse(protocolConfig);
        } catch {
          toast({
            variant: 'destructive',
            title: 'Error',
            description: 'Invalid JSON in protocol config',
          });
          setSubmitting(false);
          return;
        }
      }

      await adminApi.updateNode(selectedNode.id, data);
      toast({
        title: 'Success',
        description: 'Node updated successfully',
      });
      setEditDialogOpen(false);
      resetForm();
      fetchNodes(currentPage);
    } catch (error: any) {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: error.response?.data?.error?.message || 'Failed to update node',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteNode = async () => {
    if (!selectedNode) return;
    setSubmitting(true);

    try {
      await adminApi.deleteNode(selectedNode.id);
      toast({
        title: 'Success',
        description: 'Node deleted successfully',
      });
      setDeleteDialogOpen(false);
      setSelectedNode(null);
      fetchNodes(currentPage);
    } catch (error: any) {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: error.response?.data?.error?.message || 'Failed to delete node',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const openEditDialog = (node: Node) => {
    setSelectedNode(node);
    setName(node.name);
    setNodeType(node.node_type);
    setHost(node.host);
    setPort(node.port.toString());
    setNodeMultiplier(node.node_multiplier.toString());
    setStatus(node.status);
    setProtocolConfig(node.protocol_config ? JSON.stringify(node.protocol_config, null, 2) : '');
    setEditDialogOpen(true);
  };

  const openDeleteDialog = (node: Node) => {
    setSelectedNode(node);
    setDeleteDialogOpen(true);
  };

  const resetForm = () => {
    setName('');
    setNodeType('');
    setHost('');
    setPort('');
    setNodeMultiplier('1.0');
    setStatus('active');
    setProtocolConfig('');
    setSelectedNode(null);
  };

  if (isLoading) {
    return <div className="text-center py-12">Loading...</div>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Node Management</h1>
          <p className="text-gray-600 mt-1">Manage proxy nodes and servers</p>
        </div>
        <Button onClick={() => setCreateDialogOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create Node
        </Button>
      </div>

      {/* Nodes Table */}
      <Card>
        <CardHeader>
          <CardTitle>Nodes</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b">
                  <th className="text-left p-3">ID</th>
                  <th className="text-left p-3">Name</th>
                  <th className="text-left p-3">Type</th>
                  <th className="text-left p-3">Address</th>
                  <th className="text-left p-3">Multiplier</th>
                  <th className="text-left p-3">Status</th>
                  <th className="text-left p-3">Labels</th>
                  <th className="text-right p-3">Actions</th>
                </tr>
              </thead>
              <tbody>
                {nodes?.data.map((node) => (
                  <tr key={node.id} className="border-b hover:bg-gray-50">
                    <td className="p-3">{node.id}</td>
                    <td className="p-3 font-medium">
                      <div className="flex items-center gap-2">
                        <Server className="h-4 w-4 text-gray-500" />
                        {node.name}
                      </div>
                    </td>
                    <td className="p-3">
                      <Badge variant="outline">{node.node_type}</Badge>
                    </td>
                    <td className="p-3 text-sm text-muted-foreground">
                      {node.host}:{node.port}
                    </td>
                    <td className="p-3">{node.node_multiplier}x</td>
                    <td className="p-3">
                      <Badge variant={node.status === 'active' ? 'default' : 'secondary'}>
                        {node.status}
                      </Badge>
                    </td>
                    <td className="p-3">
                      <div className="flex gap-1 flex-wrap">
                        {node.labels?.length > 0 ? (
                          node.labels.map((label) => (
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
                          onClick={() => openEditDialog(node)}
                        >
                          <Edit className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => openDeleteDialog(node)}
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
          {nodes && nodes.pagination.pages > 1 && (
            <div className="flex items-center justify-between mt-4 pt-4 border-t">
              <div className="text-sm text-muted-foreground">
                Page {nodes.pagination.page} of {nodes.pagination.pages} ({nodes.pagination.total} total)
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
                  onClick={() => setCurrentPage(p => Math.min(nodes.pagination.pages, p + 1))}
                  disabled={currentPage === nodes.pagination.pages}
                >
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Create Node Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Create New Node</DialogTitle>
            <DialogDescription>Add a new proxy node to the system</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreateNode}>
            <div className="space-y-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="create-name">Name</Label>
                  <Input
                    id="create-name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    placeholder="e.g., US-West-1"
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="create-type">Node Type</Label>
                  <Input
                    id="create-type"
                    value={nodeType}
                    onChange={(e) => setNodeType(e.target.value)}
                    placeholder="e.g., vmess, trojan"
                    required
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="create-host">Host</Label>
                  <Input
                    id="create-host"
                    value={host}
                    onChange={(e) => setHost(e.target.value)}
                    placeholder="e.g., node1.example.com"
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="create-port">Port</Label>
                  <Input
                    id="create-port"
                    type="number"
                    value={port}
                    onChange={(e) => setPort(e.target.value)}
                    placeholder="e.g., 443"
                    required
                  />
                </div>
              </div>

              <div>
                <Label htmlFor="create-multiplier">Node Multiplier</Label>
                <Input
                  id="create-multiplier"
                  type="number"
                  step="0.1"
                  value={nodeMultiplier}
                  onChange={(e) => setNodeMultiplier(e.target.value)}
                  placeholder="e.g., 1.0"
                  required
                />
                <p className="text-xs text-muted-foreground mt-1">
                  Traffic multiplier for this node (1.0 = normal rate)
                </p>
              </div>

              <div>
                <Label htmlFor="create-protocol">Protocol Config (JSON, Optional)</Label>
                <textarea
                  id="create-protocol"
                  value={protocolConfig}
                  onChange={(e) => setProtocolConfig(e.target.value)}
                  className="w-full p-2 border rounded-md font-mono text-sm min-h-[100px]"
                  placeholder='{"method": "aes-256-gcm", "network": "tcp"}'
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setCreateDialogOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? 'Creating...' : 'Create Node'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Edit Node Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Edit Node</DialogTitle>
            <DialogDescription>Update node information</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleUpdateNode}>
            <div className="space-y-4 py-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="edit-name">Name</Label>
                  <Input
                    id="edit-name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                  />
                </div>
                <div>
                  <Label htmlFor="edit-type">Node Type</Label>
                  <Input
                    id="edit-type"
                    value={nodeType}
                    onChange={(e) => setNodeType(e.target.value)}
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="edit-host">Host</Label>
                  <Input
                    id="edit-host"
                    value={host}
                    onChange={(e) => setHost(e.target.value)}
                  />
                </div>
                <div>
                  <Label htmlFor="edit-port">Port</Label>
                  <Input
                    id="edit-port"
                    type="number"
                    value={port}
                    onChange={(e) => setPort(e.target.value)}
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="edit-multiplier">Node Multiplier</Label>
                  <Input
                    id="edit-multiplier"
                    type="number"
                    step="0.1"
                    value={nodeMultiplier}
                    onChange={(e) => setNodeMultiplier(e.target.value)}
                  />
                </div>
                <div>
                  <Label htmlFor="edit-status">Status</Label>
                  <select
                    id="edit-status"
                    value={status}
                    onChange={(e) => setStatus(e.target.value as 'active' | 'inactive')}
                    className="w-full p-2 border rounded-md"
                  >
                    <option value="active">Active</option>
                    <option value="inactive">Inactive</option>
                  </select>
                </div>
              </div>

              <div>
                <Label htmlFor="edit-protocol">Protocol Config (JSON)</Label>
                <textarea
                  id="edit-protocol"
                  value={protocolConfig}
                  onChange={(e) => setProtocolConfig(e.target.value)}
                  className="w-full p-2 border rounded-md font-mono text-sm min-h-[100px]"
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setEditDialogOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={submitting}>
                {submitting ? 'Updating...' : 'Update Node'}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Node</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete node "{selectedNode?.name}"? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteNode} disabled={submitting}>
              {submitting ? 'Deleting...' : 'Delete Node'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
