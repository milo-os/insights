"use client"

import type { InsightPolicy } from "@/types/insights"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { formatRelativeTime } from "@/lib/utils"
import { Clock, ExternalLink, FileCode, MoreVertical, Trash2 } from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

function PolicyPhaseBadge({ phase, suspended }: { phase?: string; suspended?: boolean }) {
  if (suspended) {
    return (
      <Badge variant="outline" className="border-slate-400 bg-slate-100 text-slate-700 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-300">
        Suspended
      </Badge>
    )
  }
  if (!phase) return null

  const classMap: Record<string, string> = {
    Active: "border-green-400 bg-green-100 text-green-700 dark:border-green-600 dark:bg-green-900/40 dark:text-green-300",
    Error: "border-red-400 bg-red-100 text-red-700 dark:border-red-600 dark:bg-red-900/40 dark:text-red-300",
  }

  return (
    <Badge variant="outline" className={classMap[phase] ?? ""}>
      {phase}
    </Badge>
  )
}

interface PolicyCardProps {
  policy: InsightPolicy
  onView?: (policy: InsightPolicy) => void
  onDelete?: (policy: InsightPolicy) => void
  onToggleSuspend?: (policy: InsightPolicy, suspended: boolean) => void
  className?: string
}

/**
 * PolicyCard displays a summary of an InsightPolicy in a card format.
 * Ideal for dashboards and list views.
 * InsightPolicies are cluster-scoped — no namespace display.
 */
export function PolicyCard({
  policy,
  onView,
  onDelete,
  onToggleSuspend,
  className,
}: PolicyCardProps) {
  const isSuspended = policy.spec.suspended || false

  return (
    <Card className={`transition-shadow hover:shadow-md ${className}`}>
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2">
            <PolicyPhaseBadge
              phase={policy.status?.phase}
              suspended={isSuspended}
            />
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <MoreVertical className="h-4 w-4" />
                <span className="sr-only">Actions</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {onView && (
                <DropdownMenuItem onClick={() => onView(policy)}>
                  <ExternalLink className="mr-2 h-4 w-4" />
                  View Details
                </DropdownMenuItem>
              )}
              {onDelete && (
                <DropdownMenuItem
                  onClick={() => onDelete(policy)}
                  className="text-destructive focus:text-destructive"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  Delete
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <CardTitle className="text-base leading-tight mt-2">
          {policy.metadata.name}
        </CardTitle>
        <CardDescription className="flex items-center gap-2 text-xs">
          <Clock className="h-3 w-3" />
          {policy.metadata.creationTimestamp
            ? formatRelativeTime(policy.metadata.creationTimestamp)
            : "Unknown time"}
        </CardDescription>
      </CardHeader>
      <CardContent className="space-y-3">
        {/* Selector info */}
        <div className="flex items-center gap-2">
          <FileCode className="h-4 w-4 text-muted-foreground" />
          <span className="font-mono text-sm">
            {policy.spec.selector.kind}
          </span>
          <span className="text-xs text-muted-foreground">
            ({policy.spec.selector.apiVersion})
          </span>
        </div>

        {/* Stats */}
        <div className="flex items-center gap-4 text-sm">
          <div className="flex items-center gap-1">
            <span className="font-semibold">{policy.spec.rules.length}</span>
            <span className="text-muted-foreground">rules</span>
          </div>
          {policy.status?.matchingResourceCount !== undefined && (
            <div className="flex items-center gap-1">
              <span className="font-semibold">{policy.status.matchingResourceCount}</span>
              <span className="text-muted-foreground">matched</span>
            </div>
          )}
          {policy.status?.totalInsightCount !== undefined && (
            <div className="flex items-center gap-1">
              <span className="font-semibold">{policy.status.totalInsightCount}</span>
              <span className="text-muted-foreground">insights</span>
            </div>
          )}
        </div>

        {/* Suspend toggle */}
        {onToggleSuspend && (
          <div className="flex items-center justify-between pt-2 border-t">
            <span className="text-sm text-muted-foreground">
              {isSuspended ? "Resume policy" : "Suspend policy"}
            </span>
            <Switch
              checked={!isSuspended}
              onCheckedChange={(checked) => onToggleSuspend(policy, !checked)}
              aria-label={`${isSuspended ? "Resume" : "Suspend"} policy ${policy.metadata.name}`}
            />
          </div>
        )}
      </CardContent>
    </Card>
  )
}
