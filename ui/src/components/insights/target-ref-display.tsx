"use client"

import type { TargetRef } from "@/types/insights"
import { Badge } from "@/components/ui/badge"
import { Box, FileCode } from "lucide-react"

interface TargetRefDisplayProps {
  targetRef: TargetRef
  compact?: boolean
  className?: string
}

/**
 * TargetRefDisplay shows information about the Kubernetes resource
 * that an insight references. Can be used in tables, cards, or detail views.
 */
export function TargetRefDisplay({ targetRef, compact = false, className }: TargetRefDisplayProps) {
  if (compact) {
    return (
      <div className={`flex items-center gap-2 ${className}`}>
        <Box className="h-4 w-4 text-muted-foreground" />
        <span className="font-mono text-sm">
          {targetRef.kind}/{targetRef.name}
        </span>
      </div>
    )
  }

  return (
    <div className={`space-y-1 ${className}`}>
      <div className="flex items-center gap-2">
        <Badge variant="outline" className="font-mono text-xs">
          {targetRef.apiVersion}
        </Badge>
        <Badge variant="secondary" className="font-mono text-xs">
          {targetRef.kind}
        </Badge>
      </div>
      <div className="flex items-center gap-2 text-sm">
        <FileCode className="h-4 w-4 text-muted-foreground" />
        <span className="font-mono">{targetRef.name}</span>
        {targetRef.namespace && (
          <span className="text-muted-foreground">in {targetRef.namespace}</span>
        )}
      </div>
    </div>
  )
}
