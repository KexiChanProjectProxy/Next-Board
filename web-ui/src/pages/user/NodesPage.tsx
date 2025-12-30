import { useEffect, useState } from 'react';
import { useUserStore } from '@/stores/userStore';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Copy, QrCode, Search } from 'lucide-react';
import { useToast } from '@/components/ui/use-toast';
import { generateQRCode, generateNodeConfig } from '@/utils/qrcode';
import type { Node } from '@/types';

export function NodesPage() {
  const { nodes, fetchNodes, isLoading } = useUserStore();
  const { toast } = useToast();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedType, setSelectedType] = useState<string>('all');
  const [qrCodeUrl, setQrCodeUrl] = useState<string>('');
  const [qrDialogOpen, setQrDialogOpen] = useState(false);

  useEffect(() => {
    fetchNodes();
  }, []);

  const filteredNodes = nodes.filter((node) => {
    const matchesSearch = node.name.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesType = selectedType === 'all' || node.node_type === selectedType;
    return matchesSearch && matchesType;
  });

  const nodeTypes = ['all', ...new Set(nodes.map((n) => n.node_type))];

  const copyNodeInfo = (node: Node) => {
    const config = generateNodeConfig(node);
    navigator.clipboard.writeText(config);
    toast({
      title: 'Copied',
      description: 'Node configuration copied to clipboard',
    });
  };

  const showQRCode = async (node: Node) => {
    const config = generateNodeConfig(node);
    const qrCode = await generateQRCode(config);
    setQrCodeUrl(qrCode);
    setQrDialogOpen(true);
  };

  if (isLoading) {
    return <div className="text-center py-12">Loading...</div>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Nodes</h1>
        <p className="text-gray-600 mt-1">Available proxy nodes</p>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-3 h-4 w-4 text-gray-400" />
          <Input
            placeholder="Search nodes..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-10"
          />
        </div>
        <div className="flex gap-2 flex-wrap">
          {nodeTypes.map((type) => (
            <Button
              key={type}
              variant={selectedType === type ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSelectedType(type)}
            >
              {type === 'all' ? 'All Types' : type.toUpperCase()}
            </Button>
          ))}
        </div>
      </div>

      {/* Nodes Grid */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {filteredNodes.map((node) => (
          <Card key={node.id}>
            <CardHeader>
              <div className="flex items-start justify-between">
                <div>
                  <CardTitle className="text-lg">{node.name}</CardTitle>
                  <Badge variant="secondary" className="mt-2">
                    {node.node_type.toUpperCase()}
                  </Badge>
                </div>
                <Badge variant={node.status === 'active' ? 'default' : 'destructive'}>
                  {node.status}
                </Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Host:</span>
                  <span className="font-medium">{node.host}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Port:</span>
                  <span className="font-medium">{node.port}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Multiplier:</span>
                  <span className="font-medium">{node.node_multiplier}x</span>
                </div>
              </div>

              {node.labels && node.labels.length > 0 && (
                <div>
                  <p className="text-sm text-muted-foreground mb-2">Labels</p>
                  <div className="flex flex-wrap gap-1">
                    {node.labels.map((label) => (
                      <Badge key={label.id} variant="outline" className="text-xs">
                        {label.name}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}

              <div className="flex gap-2 pt-2">
                <Button
                  size="sm"
                  variant="outline"
                  className="flex-1"
                  onClick={() => copyNodeInfo(node)}
                >
                  <Copy className="h-4 w-4 mr-1" />
                  Copy
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  className="flex-1"
                  onClick={() => showQRCode(node)}
                >
                  <QrCode className="h-4 w-4 mr-1" />
                  QR Code
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {filteredNodes.length === 0 && (
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            No nodes found matching your criteria
          </CardContent>
        </Card>
      )}

      {/* QR Code Dialog */}
      <Dialog open={qrDialogOpen} onOpenChange={setQrDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Node QR Code</DialogTitle>
            <DialogDescription>
              Scan this QR code with your proxy client
            </DialogDescription>
          </DialogHeader>
          {qrCodeUrl && (
            <div className="flex justify-center py-4">
              <img src={qrCodeUrl} alt="QR Code" className="w-64 h-64" />
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
