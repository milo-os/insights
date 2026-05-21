"use client"

import type { Insight } from "@/types/insights"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { SeverityBadge } from "./severity-badge"
import { InsightStateBadge } from "./insight-state-badge"
import { formatRelativeTime } from "@/lib/utils"

interface InsightsTableProps {
  insights: Insight[]
  onView?: (insight: Insight) => void
  className?: string
}

/**
 * InsightsTable displays a list of insights in a tabular format.
 * Suitable for admin dashboards and detailed views.
 */
export function InsightsTable({ insights, onView, className }: InsightsTableProps) {
  if (insights.length === 0) {
    return (
      <div className={`flex flex-col items-center justify-center py-12 text-center ${className}`}>
        <p className="text-lg font-medium text-muted-foreground">No insights found</p>
        <p className="text-sm text-muted-foreground">
          Insights will appear here when they are created by policies or manually.
        </p>
      </div>
    )
  }

  return (
    <div className={className}>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Insight</TableHead>
            <TableHead>Target</TableHead>
            <TableHead>Category</TableHead>
            <TableHead>State</TableHead>
            <TableHead>Age</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {insights.map((insight) => (
            <TableRow
              key={`${insight.metadata.namespace}/${insight.metadata.name}`}
              className={onView ? "cursor-pointer" : ""}
              onClick={() => onView?.(insight)}
              tabIndex={onView ? 0 : undefined}
              onKeyDown={(e) => {
                if (onView && (e.key === "Enter" || e.key === " ")) {
                  e.preventDefault()
                  onView(insight)
                }
              }}
              role={onView ? "button" : undefined}
              aria-label={onView ? `View insight: ${insight.spec.message}` : undefined}
            >
              <TableCell className="max-w-md">
                <div className="flex items-start gap-2">
                  <SeverityBadge severity={insight.spec.severity} showIcon={false} className="mt-0.5 shrink-0" />
                  <div className="min-w-0">
                    <p className="truncate font-medium">{insight.spec.message}</p>
                    <p className="truncate text-xs text-muted-foreground font-mono">
                      {insight.metadata.namespace}/{insight.metadata.name}
                    </p>
                  </div>
                </div>
              </TableCell>
              <TableCell>
                <span className="font-mono text-sm">
                  {insight.spec.targetRef.kind}/{insight.spec.targetRef.name}
                </span>
              </TableCell>
              <TableCell>
                <span className="text-sm text-muted-foreground">
                  {insight.spec.category}
                </span>
              </TableCell>
              <TableCell>
                <InsightStateBadge insight={insight} />
              </TableCell>
              <TableCell className="text-sm text-muted-foreground whitespace-nowrap">
                {insight.metadata.creationTimestamp
                  ? formatRelativeTime(insight.metadata.creationTimestamp)
                  : "-"}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
