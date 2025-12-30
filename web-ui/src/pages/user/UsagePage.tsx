import { useEffect } from 'react';
import { useUserStore } from '@/stores/userStore';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { formatBytes, formatDate, calculateUsagePercentage, getUsageColorClass } from '@/utils/format';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { TrendingUp, TrendingDown, Activity } from 'lucide-react';

export function UsagePage() {
  const { usage, plan, fetchUsage, fetchPlan, isLoading } = useUserStore();

  useEffect(() => {
    fetchUsage();
    fetchPlan();
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

  // Chart data
  const chartData = [
    {
      name: 'Upload',
      Real: usage?.real_bytes_up || 0,
      Billable: usage?.billable_bytes_up || 0,
    },
    {
      name: 'Download',
      Real: usage?.real_bytes_down || 0,
      Billable: usage?.billable_bytes_down || 0,
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Usage</h1>
        <p className="text-gray-600 mt-1">Track your data usage and quota</p>
      </div>

      {/* Current Period */}
      <Card>
        <CardHeader>
          <CardTitle>Current Billing Period</CardTitle>
          <CardDescription>
            {usage?.period_start && usage?.period_end ? (
              <>
                {formatDate(usage.period_start)} - {formatDate(usage.period_end)}
              </>
            ) : (
              'No active period'
            )}
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
              {usagePercentage > 80 && (
                <span className="text-orange-600 ml-2">⚠ Approaching quota limit</span>
              )}
            </p>
          </div>
        </CardContent>
      </Card>

      {/* Stats Cards */}
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Upload Traffic</CardTitle>
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
            <CardTitle className="text-sm font-medium">Download Traffic</CardTitle>
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

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Traffic</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatBytes(totalBillableTraffic)}
            </div>
            <p className="text-xs text-muted-foreground">
              Real: {formatBytes(totalRealTraffic)}
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Traffic Breakdown Chart */}
      <Card>
        <CardHeader>
          <CardTitle>Traffic Breakdown</CardTitle>
          <CardDescription>Real vs Billable traffic comparison</CardDescription>
        </CardHeader>
        <CardContent>
          <ResponsiveContainer width="100%" height={300}>
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" />
              <XAxis dataKey="name" />
              <YAxis tickFormatter={(value) => formatBytes(value)} />
              <Tooltip
                formatter={(value: number) => formatBytes(value)}
                labelStyle={{ color: '#000' }}
              />
              <Legend />
              <Bar dataKey="Real" fill="#3b82f6" name="Real Traffic" />
              <Bar dataKey="Billable" fill="#ef4444" name="Billable Traffic" />
            </BarChart>
          </ResponsiveContainer>
        </CardContent>
      </Card>

      {/* Multiplier Effect */}
      <Card>
        <CardHeader>
          <CardTitle>Multiplier Effect</CardTitle>
          <CardDescription>How multipliers affect your usage</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <p className="text-sm text-muted-foreground">Real Traffic</p>
              <p className="text-2xl font-bold">{formatBytes(totalRealTraffic)}</p>
            </div>
            <div>
              <p className="text-sm text-muted-foreground">Billable Traffic</p>
              <p className="text-2xl font-bold">{formatBytes(totalBillableTraffic)}</p>
            </div>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">
              Billable traffic is calculated by applying node multipliers and plan multipliers to your real traffic usage.
            </p>
          </div>
          {plan && (
            <div className="pt-2 border-t">
              <p className="text-sm font-medium">Plan Base Multiplier: {plan.base_multiplier}x</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Usage History Placeholder */}
      <Card>
        <CardHeader>
          <CardTitle>Usage History</CardTitle>
          <CardDescription>Historical data will be available soon</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-12 text-muted-foreground">
            <Activity className="h-12 w-12 mx-auto mb-4 opacity-50" />
            <p>Historical usage data tracking is not yet implemented</p>
            <p className="text-sm mt-2">Check back later for detailed usage trends</p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
