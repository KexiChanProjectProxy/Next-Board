import { useEffect } from 'react';
import { useUserStore } from '@/stores/userStore';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Badge } from '@/components/ui/badge';
import { formatBytes, formatDate, calculateUsagePercentage, getUsageColorClass } from '@/utils/format';
import { Server, Package, TrendingUp, TrendingDown } from 'lucide-react';

export function DashboardPage() {
  const { profile, plan, usage, nodes, fetchAll, isLoading } = useUserStore();

  useEffect(() => {
    fetchAll();
  }, []);

  if (isLoading) {
    return <div className="text-center py-12">Loading...</div>;
  }

  const usagePercentage = usage && plan
    ? calculateUsagePercentage(usage.billable_bytes_up, usage.billable_bytes_down, plan.quota_bytes)
    : 0;

  const totalRealTraffic = usage
    ? usage.real_bytes_up + usage.real_bytes_down
    : 0;

  const totalBillableTraffic = usage
    ? usage.billable_bytes_up + usage.billable_bytes_down
    : 0;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Dashboard</h1>
        <p className="text-gray-600 mt-1">Welcome back, {profile?.email}</p>
      </div>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Plan</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{plan?.name || 'No Plan'}</div>
            <p className="text-xs text-muted-foreground">
              {plan ? `${formatBytes(plan.quota_bytes)} quota` : 'Not assigned'}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Available Nodes</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{nodes.length}</div>
            <p className="text-xs text-muted-foreground">
              {nodes.filter(n => n.status === 'active').length} active
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Upload</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatBytes(usage?.billable_bytes_up || 0)}
            </div>
            <p className="text-xs text-muted-foreground">
              Real: {formatBytes(usage?.real_bytes_up || 0)}
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Download</CardTitle>
            <TrendingDown className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatBytes(usage?.billable_bytes_down || 0)}
            </div>
            <p className="text-xs text-muted-foreground">
              Real: {formatBytes(usage?.real_bytes_down || 0)}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Usage Summary */}
      <Card>
        <CardHeader>
          <CardTitle>Usage Summary</CardTitle>
          <CardDescription>
            Current billing period: {usage?.period_start ? formatDate(usage.period_start) : 'N/A'}
            {' - '}
            {usage?.period_end ? formatDate(usage.period_end) : 'N/A'}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium">Quota Usage</span>
              <span className="text-sm text-muted-foreground">
                {formatBytes(totalBillableTraffic)} / {formatBytes(plan?.quota_bytes || 0)}
              </span>
            </div>
            <Progress
              value={usagePercentage}
              className={getUsageColorClass(usagePercentage)}
            />
            <p className="text-xs text-muted-foreground mt-2">
              {usagePercentage.toFixed(2)}% used
            </p>
          </div>

          <div className="grid grid-cols-2 gap-4 pt-4">
            <div>
              <p className="text-sm text-muted-foreground">Real Traffic</p>
              <p className="text-lg font-semibold">{formatBytes(totalRealTraffic)}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Billable Traffic</p>
              <p className="text-lg font-semibold">{formatBytes(totalBillableTraffic)}</p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Plan Details */}
      {plan && (
        <Card>
          <CardHeader>
            <CardTitle>Plan Details</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-muted-foreground">Plan Name</p>
                <p className="font-medium">{plan.name}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Reset Period</p>
                <p className="font-medium capitalize">{plan.reset_period}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Quota</p>
                <p className="font-medium">{formatBytes(plan.quota_bytes)}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Base Multiplier</p>
                <p className="font-medium">{plan.base_multiplier}x</p>
              </div>
            </div>

            {plan.labels && plan.labels.length > 0 && (
              <div>
                <p className="text-sm text-muted-foreground mb-2">Included Labels</p>
                <div className="flex flex-wrap gap-2">
                  {plan.labels.map((label) => (
                    <Badge key={label.id} variant="secondary">
                      {label.name} ({label.multiplier}x)
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
