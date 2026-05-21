"use client"

import type { InsightStatus, ActorReference } from "@/types/insights"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import { formatDate, formatRelativeTime } from "@/lib/utils"
import {
  CheckCircle2,
  BellOff,
  UserRound,
  CheckCircle,
  History,
} from "lucide-react"

interface InsightLifecycleHistoryProps {
  status: InsightStatus
}

function ActorBadge({ actor }: { actor: ActorReference }) {
  return (
    <span className="inline-flex items-center gap-1 text-sm font-medium">
      <UserRound className="h-3.5 w-3.5 text-muted-foreground" />
      <span className="font-mono">{actor.name}</span>
      {actor.type === "serviceaccount" && (
        <Badge variant="outline" className="text-xs px-1 py-0">
          SA
        </Badge>
      )}
    </span>
  )
}

function TimestampCell({ timestamp }: { timestamp: string }) {
  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <span className="text-sm text-muted-foreground cursor-default">
            {formatRelativeTime(timestamp)}
          </span>
        </TooltipTrigger>
        <TooltipContent>
          <p>{formatDate(timestamp)}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}

/**
 * InsightLifecycleHistory shows a timeline of lifecycle events for an insight.
 * Each event (acknowledgement, snooze, assignment, resolution) is shown when present.
 */
export function InsightLifecycleHistory({ status }: InsightLifecycleHistoryProps) {
  const hasHistory =
    status.acknowledgement ||
    status.snooze ||
    status.assignment ||
    status.resolution

  if (!hasHistory) {
    return null
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="flex items-center gap-2 text-base">
          <History className="h-4 w-4" />
          Lifecycle History
        </CardTitle>
      </CardHeader>
      <CardContent>
        <ol className="relative border-l border-border space-y-6 pl-6">
          {status.acknowledgement && (
            <li className="relative">
              <div className="absolute -left-[1.625rem] flex h-6 w-6 items-center justify-center rounded-full bg-purple-100 dark:bg-purple-900/50 ring-2 ring-background">
                <CheckCircle2 className="h-3.5 w-3.5 text-purple-600 dark:text-purple-400" />
              </div>
              <div className="space-y-1">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-medium text-sm">Acknowledged by</span>
                  <ActorBadge actor={status.acknowledgement.by} />
                  <TimestampCell timestamp={status.acknowledgement.at} />
                </div>
                {status.acknowledgement.note && (
                  <p className="text-sm text-muted-foreground italic">
                    &ldquo;{status.acknowledgement.note}&rdquo;
                  </p>
                )}
              </div>
            </li>
          )}

          {status.snooze && (
            <li className="relative">
              <div className="absolute -left-[1.625rem] flex h-6 w-6 items-center justify-center rounded-full bg-slate-100 dark:bg-slate-800 ring-2 ring-background">
                <BellOff className="h-3.5 w-3.5 text-slate-600 dark:text-slate-400" />
              </div>
              <div className="space-y-1">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-medium text-sm">Snoozed by</span>
                  <ActorBadge actor={status.snooze.by} />
                  <TimestampCell timestamp={status.snooze.at} />
                </div>
                <p className="text-sm text-muted-foreground">
                  Until {formatDate(status.snooze.until)}
                </p>
              </div>
            </li>
          )}

          {status.assignment && (
            <li className="relative">
              <div className="absolute -left-[1.625rem] flex h-6 w-6 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900/50 ring-2 ring-background">
                <UserRound className="h-3.5 w-3.5 text-blue-600 dark:text-blue-400" />
              </div>
              <div className="space-y-1">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-medium text-sm">Assigned by</span>
                  <ActorBadge actor={status.assignment.by} />
                  <span className="text-sm text-muted-foreground">to</span>
                  <ActorBadge actor={status.assignment.to} />
                  <TimestampCell timestamp={status.assignment.at} />
                </div>
                {status.assignment.note && (
                  <p className="text-sm text-muted-foreground italic">
                    &ldquo;{status.assignment.note}&rdquo;
                  </p>
                )}
              </div>
            </li>
          )}

          {status.resolution && (
            <li className="relative">
              <div className="absolute -left-[1.625rem] flex h-6 w-6 items-center justify-center rounded-full bg-green-100 dark:bg-green-900/50 ring-2 ring-background">
                <CheckCircle className="h-3.5 w-3.5 text-green-600 dark:text-green-400" />
              </div>
              <div className="space-y-1">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="font-medium text-sm">Resolved by</span>
                  <ActorBadge actor={status.resolution.by} />
                  <TimestampCell timestamp={status.resolution.at} />
                </div>
                {status.resolution.note && (
                  <p className="text-sm text-muted-foreground italic">
                    &ldquo;{status.resolution.note}&rdquo;
                  </p>
                )}
              </div>
            </li>
          )}
        </ol>
      </CardContent>
    </Card>
  )
}
