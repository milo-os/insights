import type { Insight, InsightState } from "@/types/insights"

/**
 * Available snooze duration preset options
 */
export const SNOOZE_DURATIONS = [
  { label: "1 hour", value: "1h" },
  { label: "4 hours", value: "4h" },
  { label: "1 day", value: "24h" },
  { label: "1 week", value: "168h" },
] as const

export type SnoozeDuration = (typeof SNOOZE_DURATIONS)[number]["value"]

/**
 * Get the current state of an insight from status.state
 */
export function getInsightState(insight: Insight): InsightState {
  return insight.status?.state ?? "Active"
}

/**
 * Check if an insight is currently snoozed
 */
export function isInsightSnoozed(insight: Insight): boolean {
  return insight.status?.state === "Snoozed"
}

/**
 * Check if an insight has been acknowledged
 */
export function isInsightAcknowledged(insight: Insight): boolean {
  return insight.status?.state === "Acknowledged"
}

/**
 * Check if an insight has been resolved
 */
export function isInsightResolved(insight: Insight): boolean {
  return insight.status?.state === "Resolved"
}

/**
 * Check if an insight is persistent (no TTL/auto-expiration)
 */
export function isInsightPersistent(insight: Insight): boolean {
  return insight.spec.persistent === true
}

/**
 * Get the snooze expiration date for an insight from status.snooze.until
 */
export function getSnoozeExpiration(insight: Insight): Date | null {
  const until = insight.status?.snooze?.until
  if (!until) return null
  return new Date(until)
}

/**
 * Format the snooze remaining time for display
 */
export function formatSnoozeRemaining(insight: Insight): string {
  const expiration = getSnoozeExpiration(insight)
  if (!expiration) return ""

  const now = new Date()
  const diff = expiration.getTime() - now.getTime()

  if (diff <= 0) return "Expired"

  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))

  if (hours >= 24) {
    const days = Math.floor(hours / 24)
    return `${days}d ${hours % 24}h`
  }

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }

  return `${minutes}m`
}

/**
 * Determine which lifecycle actions are available based on the current state
 */
export interface AvailableActions {
  acknowledge: boolean
  unacknowledge: boolean
  snooze: boolean
  resolve: boolean
  assign: boolean
}

export function getAvailableActions(insight: Insight): AvailableActions {
  const state = getInsightState(insight)

  switch (state) {
    case "Active":
      return { acknowledge: true, unacknowledge: false, snooze: true, resolve: true, assign: true }
    case "Acknowledged":
      return { acknowledge: false, unacknowledge: true, snooze: true, resolve: true, assign: true }
    case "Snoozed":
      return { acknowledge: false, unacknowledge: false, snooze: false, resolve: true, assign: true }
    case "Resolved":
      return { acknowledge: false, unacknowledge: false, snooze: false, resolve: false, assign: false }
    default:
      return { acknowledge: false, unacknowledge: false, snooze: false, resolve: false, assign: false }
  }
}
