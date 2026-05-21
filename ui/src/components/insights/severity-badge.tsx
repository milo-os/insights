"use client"

import { Badge } from "@/components/ui/badge"
import type { InsightSeverity } from "@/types/insights"
import { AlertCircle, AlertTriangle, Info } from "lucide-react"

interface SeverityBadgeProps {
  severity: InsightSeverity
  showIcon?: boolean
  className?: string
}

/**
 * SeverityBadge displays a color-coded badge for insight severity levels.
 * Can be embedded in any UI portal that displays insights.
 */
export function SeverityBadge({ severity, showIcon = true, className }: SeverityBadgeProps) {
  const Icon = severity === "critical" ? AlertCircle : severity === "warning" ? AlertTriangle : Info

  return (
    <Badge variant={severity} className={className}>
      {showIcon && <Icon className="mr-1 h-3 w-3" />}
      {severity.charAt(0).toUpperCase() + severity.slice(1)}
    </Badge>
  )
}
