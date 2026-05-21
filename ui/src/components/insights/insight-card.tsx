"use client"

import type { Insight } from "@/types/insights"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { SeverityBadge } from "./severity-badge"
import { InsightStateBadge } from "./insight-state-badge"
import { TargetRefDisplay } from "./target-ref-display"
import { Button } from "@/components/ui/button"
import { formatRelativeTime } from "@/lib/utils"
import { Clock, ExternalLink, MoreVertical, Trash2 } from "lucide-react"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"

interface InsightCardProps {
  insight: Insight
  onView?: (insight: Insight) => void
  onDelete?: (insight: Insight) => void
  className?: string
}

/**
 * InsightCard displays a summary of an Insight in a card format.
 * Ideal for dashboards and list views. Can be embedded in any portal.
 */
export function InsightCard({ insight, onView, onDelete, className }: InsightCardProps) {
  const handleCardClick = () => {
    if (onView) {
      onView(insight)
    }
  }

  return (
    <Card
      className={`transition-shadow hover:shadow-md ${onView ? "cursor-pointer" : ""} ${className}`}
      onClick={handleCardClick}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-2 flex-wrap">
            <SeverityBadge severity={insight.spec.severity} />
            <InsightStateBadge insight={insight} />
          </div>
          <div onClick={(e) => e.stopPropagation()}>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8">
                  <MoreVertical className="h-4 w-4" />
                  <span className="sr-only">Actions</span>
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {onView && (
                  <DropdownMenuItem onClick={() => onView(insight)}>
                    <ExternalLink className="mr-2 h-4 w-4" />
                    View Details
                  </DropdownMenuItem>
                )}
                {onDelete && (
                  <DropdownMenuItem
                    onClick={() => onDelete(insight)}
                    className="text-destructive focus:text-destructive"
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </div>
        <CardTitle className="text-base leading-tight mt-2">
          {insight.spec.message}
        </CardTitle>
        <CardDescription className="flex items-center gap-2 text-xs">
          <Clock className="h-3 w-3" />
          {insight.metadata.creationTimestamp
            ? formatRelativeTime(insight.metadata.creationTimestamp)
            : "Unknown time"}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <TargetRefDisplay targetRef={insight.spec.targetRef} compact />
      </CardContent>
    </Card>
  )
}
