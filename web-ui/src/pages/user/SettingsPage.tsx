import { useEffect, useState } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { useUserStore } from '@/stores/userStore';
import { userApi } from '@/api/user';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { useToast } from '@/components/ui/use-toast';
import { formatDate } from '@/utils/format';
import { User, Link as LinkIcon, CheckCircle2, XCircle } from 'lucide-react';

export function SettingsPage() {
  const { user } = useAuthStore();
  const { profile, plan, fetchProfile, fetchPlan } = useUserStore();
  const { toast } = useToast();
  const [linkToken, setLinkToken] = useState<string>('');
  const [linkDialogOpen, setLinkDialogOpen] = useState(false);
  const [isGenerating, setIsGenerating] = useState(false);

  useEffect(() => {
    fetchProfile();
    fetchPlan();
  }, []);

  const handleGenerateLink = async () => {
    setIsGenerating(true);
    try {
      const token = await userApi.generateTelegramLink();
      setLinkToken(token);
      setLinkDialogOpen(true);
      toast({
        title: 'Link Token Generated',
        description: 'Use this token to link your Telegram account',
      });
    } catch (error: any) {
      toast({
        variant: 'destructive',
        title: 'Error',
        description: error.response?.data?.message || 'Failed to generate link token',
      });
    } finally {
      setIsGenerating(false);
    }
  };

  const copyToken = () => {
    navigator.clipboard.writeText(linkToken);
    toast({
      title: 'Copied',
      description: 'Link token copied to clipboard',
    });
  };

  const isLinked = profile?.telegram_chat_id !== null;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Settings</h1>
        <p className="text-gray-600 mt-1">Manage your account settings</p>
      </div>

      {/* Profile Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <User className="h-5 w-5" />
            Profile Information
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <p className="text-sm text-muted-foreground">Email</p>
              <p className="font-medium">{profile?.email}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Role</p>
              <Badge variant={profile?.role === 'admin' ? 'default' : 'secondary'}>
                {profile?.role}
              </Badge>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Account Created</p>
              <p className="font-medium">
                {profile?.created_at ? formatDate(profile.created_at) : 'N/A'}
              </p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Current Plan</p>
              <p className="font-medium">{plan?.name || 'No plan assigned'}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Telegram Integration */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <LinkIcon className="h-5 w-5" />
            Telegram Integration
          </CardTitle>
          <CardDescription>
            Link your Telegram account to receive notifications
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              {isLinked ? (
                <CheckCircle2 className="h-5 w-5 text-green-600" />
              ) : (
                <XCircle className="h-5 w-5 text-gray-400" />
              )}
              <div>
                <p className="font-medium">
                  {isLinked ? 'Telegram Linked' : 'Not Linked'}
                </p>
                {isLinked && profile?.telegram_chat_id && (
                  <p className="text-sm text-muted-foreground">
                    Chat ID: {profile.telegram_chat_id}
                  </p>
                )}
                {isLinked && profile?.telegram_linked_at && (
                  <p className="text-xs text-muted-foreground">
                    Linked {formatDate(profile.telegram_linked_at)}
                  </p>
                )}
              </div>
            </div>
            {!isLinked && (
              <Button onClick={handleGenerateLink} disabled={isGenerating}>
                {isGenerating ? 'Generating...' : 'Generate Link Token'}
              </Button>
            )}
          </div>

          {!isLinked && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <p className="text-sm text-blue-900">
                <strong>How to link:</strong>
              </p>
              <ol className="text-sm text-blue-800 mt-2 space-y-1 list-decimal list-inside">
                <li>Click "Generate Link Token" above</li>
                <li>Find the bot on Telegram (check with your administrator)</li>
                <li>Send the command: /link &lt;your_token&gt;</li>
                <li>Your account will be linked automatically</li>
              </ol>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Security Section (Placeholder) */}
      <Card>
        <CardHeader>
          <CardTitle>Security</CardTitle>
          <CardDescription>Manage your account security</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Change Password</p>
                <p className="text-sm text-muted-foreground">
                  Update your password to keep your account secure
                </p>
              </div>
              <Button variant="outline" disabled>
                Coming Soon
              </Button>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Active Sessions</p>
                <p className="text-sm text-muted-foreground">
                  View and manage your active sessions
                </p>
              </div>
              <Button variant="outline" disabled>
                Coming Soon
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Telegram Link Dialog */}
      <Dialog open={linkDialogOpen} onOpenChange={setLinkDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Telegram Link Token</DialogTitle>
            <DialogDescription>
              Use this token to link your Telegram account
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="bg-gray-100 rounded-lg p-4 font-mono text-center text-lg">
              {linkToken}
            </div>
            <Button onClick={copyToken} className="w-full">
              Copy Token
            </Button>
            <div className="text-sm text-muted-foreground">
              <p className="font-medium mb-2">Instructions:</p>
              <ol className="list-decimal list-inside space-y-1">
                <li>Open Telegram and find the bot</li>
                <li>Send: <code className="bg-gray-200 px-1 rounded">/link {linkToken}</code></li>
                <li>Wait for confirmation</li>
              </ol>
            </div>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
