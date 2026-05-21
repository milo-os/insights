"use client"

import { Badge } from "@/components/ui/badge"
import type { InsightPolicyPhase } from "@/types/insights"

interface PhaseBadgeProps {
  phase?: InsightPolicyPhase
  className?: string
}

/**
 * PhaseBadge displays a color-coded badge for InsightPolicy phases.
 * Use InsightStateBadge for Insight lifecycle states.
 */
export function PhaseBadge({ phase, className }: PhaseBadgeProps) {
  if (!phase) {
    return <Badge variant="outline" className={className}>Unknown</Badge>
  }

  const classMap: Record<InsightPolicyPhase, string> = {
    Active: "border-green-400 bg-green-100 text-green-700 dark:border-green-600 dark:bg-green-900/40 dark:text-green-300",
    Suspended: "border-slate-400 bg-slate-100 text-slate-700 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-300",
    Error: "border-red-400 bg-red-100 text-red-700 dark:border-red-600 dark:bg-red-900/40 dark:text-red-300",
  }

  return (
    <Badge variant="outline" className={`${classMap[phase]} ${className ?? ""}`}>
      {phase}
    </Badge>
  )
}
