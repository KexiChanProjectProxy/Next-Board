import { format, formatDistanceToNow } from 'date-fns';

/**
 * Format bytes to human-readable format
 */
export function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 Bytes';

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB'];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

/**
 * Convert GB to bytes
 */
export function gbToBytes(gb: number): number {
  return gb * 1024 * 1024 * 1024;
}

/**
 * Convert bytes to GB
 */
export function bytesToGB(bytes: number): number {
  return bytes / (1024 * 1024 * 1024);
}

/**
 * Format date to readable string
 */
export function formatDate(dateString: string): string {
  return format(new Date(dateString), 'yyyy-MM-dd HH:mm:ss');
}

/**
 * Format relative time (e.g., "2 hours ago")
 */
export function formatRelativeTime(dateString: string): string {
  return formatDistanceToNow(new Date(dateString), { addSuffix: true });
}

/**
 * Calculate usage percentage
 */
export function calculateUsagePercentage(
  billableUp: number,
  billableDown: number,
  quota: number
): number {
  if (quota === 0) return 0;
  const totalUsed = billableUp + billableDown;
  return Math.min((totalUsed / quota) * 100, 100);
}

/**
 * Get usage color based on percentage
 */
export function getUsageColor(percentage: number): string {
  if (percentage < 50) return 'green';
  if (percentage < 80) return 'yellow';
  if (percentage < 95) return 'orange';
  return 'red';
}

/**
 * Get Tailwind color class based on usage percentage
 */
export function getUsageColorClass(percentage: number): string {
  if (percentage < 50) return 'bg-green-500';
  if (percentage < 80) return 'bg-yellow-500';
  if (percentage < 95) return 'bg-orange-500';
  return 'bg-red-500';
}
