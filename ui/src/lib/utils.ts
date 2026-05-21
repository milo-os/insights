import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDate(date: string | Date): string {
  return new Date(date).toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  })
}

export function formatRelativeTime(date: string | Date): string {
  const now = new Date()
  const target = new Date(date)
  const diffMs = now.getTime() - target.getTime()
  const diffSecs = Math.floor(diffMs / 1000)
  const diffMins = Math.floor(diffSecs / 60)
  const diffHours = Math.floor(diffMins / 60)
  const diffDays = Math.floor(diffHours / 24)

  if (diffSecs < 60) return "just now"
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return formatDate(date)
}

export function severityToColor(severity: string): string {
  switch (severity) {
    case "critical":
      return "text-red-600 dark:text-red-400"
    case "warning":
      return "text-amber-600 dark:text-amber-400"
    case "info":
    default:
      return "text-blue-600 dark:text-blue-400"
  }
}

export function severityToBgColor(severity: string): string {
  switch (severity) {
    case "critical":
      return "bg-red-100 dark:bg-red-900/30"
    case "warning":
      return "bg-amber-100 dark:bg-amber-900/30"
    case "info":
    default:
      return "bg-blue-100 dark:bg-blue-900/30"
  }
}

export function phaseToColor(phase: string): string {
  switch (phase) {
    case "Active":
      return "text-green-600 dark:text-green-400"
    case "Expired":
      return "text-gray-500 dark:text-gray-400"
    case "Resolved":
      return "text-blue-600 dark:text-blue-400"
    case "Suspended":
      return "text-amber-600 dark:text-amber-400"
    case "Error":
      return "text-red-600 dark:text-red-400"
    default:
      return "text-gray-600 dark:text-gray-400"
  }
}

export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  return `${days}d ${hours}h`
}

export function calculateTTLProgress(
  creationTimestamp: string | undefined,
  ttlSeconds: number | undefined,
  expiresAt: string | undefined
): { percentage: number; remaining: string; isExpired: boolean } {
  if (!ttlSeconds || ttlSeconds <= 0) {
    return { percentage: 100, remaining: "No expiration", isExpired: false }
  }

  const now = new Date()
  let expiration: Date

  if (expiresAt) {
    expiration = new Date(expiresAt)
  } else if (creationTimestamp) {
    expiration = new Date(new Date(creationTimestamp).getTime() + ttlSeconds * 1000)
  } else {
    return { percentage: 100, remaining: "Unknown", isExpired: false }
  }

  const remainingMs = expiration.getTime() - now.getTime()

  if (remainingMs <= 0) {
    return { percentage: 0, remaining: "Expired", isExpired: true }
  }

  const remainingSecs = Math.floor(remainingMs / 1000)
  const percentage = Math.min(100, Math.max(0, (remainingSecs / ttlSeconds) * 100))

  return {
    percentage,
    remaining: formatDuration(remainingSecs),
    isExpired: false
  }
}
