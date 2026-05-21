"use client"

import type { Insight } from "@/types/insights"
import { Badge } from "@/components/ui/badge"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { formatDate } from "@/lib/utils"
import { BellOff, CheckCircle2, CheckCircle } from "lucide-react"

interface InsightStateBadgeProps {
  insight: Insight
  className?: string
}

/**
 * InsightStateBadge displays the current lifecycle state of an insight.
 * Reads from status.state — not annotations.
 *
 * State to badge mapping:
 *   Active       — no badge (default state, omit)
 *   Acknowledged — purple badge with checkmark icon
 *   Snoozed      — slate badge with bell-off icon, tooltip showing snooze.until
 *   Resolved     — green badge with check-circle icon
 */
export function InsightStateBadge({ insight, className }: InsightStateBadgeProps) {
  const state = insight.status?.state

  if (!state || state === "Active") {
    return null
  }

  if (state === "Snoozed") {
    const snoozedUntil = insight.status?.snooze?.until

    return (
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Badge
              variant="outline"
              className={`gap-1 border-slate-400 bg-slate-100 text-slate-700 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-300 ${className ?? ""}`}
            >
              <BellOff className="h-3 w-3" />
              Snoozed
            </Badge>
          </TooltipTrigger>
          {snoozedUntil && (
            <TooltipContent>
              <p>Snoozed until {formatDate(snoozedUntil)}</p>
            </TooltipContent>
          )}
        </Tooltip>
      </TooltipProvider>
    )
  }

  if (state === "Acknowledged") {
    return (
      <Badge
        variant="outline"
        className={`gap-1 border-purple-400 bg-purple-100 text-purple-700 dark:border-purple-600 dark:bg-purple-900/40 dark:text-purple-300 ${className ?? ""}`}
      >
        <CheckCircle2 className="h-3 w-3" />
        Acknowledged
      </Badge>
    )
  }

  if (state === "Resolved") {
    return (
      <Badge
        variant="outline"
        className={`gap-1 border-green-400 bg-green-100 text-green-700 dark:border-green-600 dark:bg-green-900/40 dark:text-green-300 ${className ?? ""}`}
      >
        <CheckCircle className="h-3 w-3" />
        Resolved
      </Badge>
    )
  }

  return null
}
