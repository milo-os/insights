"use client"

import type { MuteRuleReference } from "@/types/insights"
import { Card, CardContent } from "@/components/ui/card"
import { VolumeX } from "lucide-react"

interface InsightMutedBannerProps {
  mutedBy?: MuteRuleReference
}

/**
 * InsightMutedBanner is shown on the insight detail page when status.muted is true.
 * It indicates that a mute rule is suppressing this insight.
 */
export function InsightMutedBanner({ mutedBy }: InsightMutedBannerProps) {
  return (
    <Card className="border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/30">
      <CardContent className="py-3">
        <div className="flex items-center gap-3">
          <VolumeX className="h-5 w-5 flex-shrink-0 text-slate-500 dark:text-slate-400" />
          <div>
            <p className="font-medium text-slate-700 dark:text-slate-300">
              This insight is muted
            </p>
            {mutedBy ? (
              <p className="text-sm text-slate-600 dark:text-slate-400">
                Suppressed by mute rule{" "}
                <span className="font-mono font-medium">
                  {mutedBy.namespace}/{mutedBy.name}
                </span>
              </p>
            ) : (
              <p className="text-sm text-slate-600 dark:text-slate-400">
                This insight is being suppressed by a mute rule.
              </p>
            )}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
