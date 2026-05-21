"use client"

import type { InsightPolicy } from "@/types/insights"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { formatRelativeTime } from "@/lib/utils"
import { ExternalLink } from "lucide-react"

function PolicyPhaseBadge({ phase, suspended }: { phase?: string; suspended?: boolean }) {
  if (suspended) {
    return (
      <Badge
        variant="outline"
        className="border-slate-400 bg-slate-100 text-slate-700 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-300"
      >
        Suspended
      </Badge>
    )
  }
  if (!phase) return <Badge variant="outline">Unknown</Badge>

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

interface PoliciesTableProps {
  policies: InsightPolicy[]
  onView?: (policy: InsightPolicy) => void
  onDelete?: (policy: InsightPolicy) => void
  onToggleSuspend?: (policy: InsightPolicy, suspended: boolean) => void
  className?: string
}

/**
 * PoliciesTable displays a list of InsightPolicies in a tabular format.
 * InsightPolicies are cluster-scoped — no namespace column.
 */
export function PoliciesTable({
  policies,
  onView,
  onDelete,
  onToggleSuspend,
  className,
}: PoliciesTableProps) {
  if (policies.length === 0) {
    return (
      <div className={`flex flex-col items-center justify-center py-12 text-center ${className}`}>
        <p className="text-lg font-medium text-muted-foreground">No policies found</p>
        <p className="text-sm text-muted-foreground">
          Create a policy to automatically generate insights for your resources.
        </p>
      </div>
    )
  }

  return (
    <div className={className}>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Target Kind</TableHead>
            <TableHead>Rules</TableHead>
            <TableHead>Matched</TableHead>
            <TableHead>Skipped</TableHead>
            <TableHead>Insights</TableHead>
            <TableHead>Phase</TableHead>
            <TableHead>Active</TableHead>
            <TableHead>Age</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {policies.map((policy) => (
            <TableRow key={policy.metadata.name}>
              <TableCell>
                <p className="font-medium">{policy.metadata.name}</p>
              </TableCell>
              <TableCell>
                <Badge variant="outline" className="font-mono text-xs">
                  {policy.spec.selector.kind}
                </Badge>
              </TableCell>
              <TableCell className="font-medium">
                {policy.spec.rules.length}
              </TableCell>
              <TableCell>
                {policy.status?.matchingResourceCount ?? 0}
              </TableCell>
              <TableCell>
                {policy.status?.skippedResourceCount ?? 0}
              </TableCell>
              <TableCell>
                {policy.status?.totalInsightCount ?? 0}
              </TableCell>
              <TableCell>
                <PolicyPhaseBadge
                  phase={policy.status?.phase}
                  suspended={policy.spec.suspended}
                />
              </TableCell>
              <TableCell>
                {onToggleSuspend ? (
                  <Switch
                    checked={!policy.spec.suspended}
                    onCheckedChange={(checked) => onToggleSuspend(policy, !checked)}
                    aria-label={`${policy.spec.suspended ? "Enable" : "Suspend"} policy ${policy.metadata.name}`}
                  />
                ) : (
                  <span className="text-sm">
                    {policy.spec.suspended ? "No" : "Yes"}
                  </span>
                )}
              </TableCell>
              <TableCell className="text-sm text-muted-foreground whitespace-nowrap">
                {policy.metadata.creationTimestamp
                  ? formatRelativeTime(policy.metadata.creationTimestamp)
                  : "-"}
              </TableCell>
              <TableCell className="text-right">
                <div className="flex items-center justify-end gap-1">
                  {onView && (
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => onView(policy)}
                      aria-label={`View policy ${policy.metadata.name}`}
                    >
                      <ExternalLink className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
