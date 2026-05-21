"use client"

import type { DashboardStats, Insight, InsightPolicy } from "@/types/insights"
import { StatsCard } from "./stats-card"
import { InsightCard } from "@/components/insights/insight-card"
import { PolicyCard } from "@/components/policies/policy-card"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
  AlertCircle,
  AlertTriangle,
  Lightbulb,
  ScrollText,
  Shield,
} from "lucide-react"

interface DashboardOverviewProps {
  stats: DashboardStats
  recentInsights: Insight[]
  recentPolicies: InsightPolicy[]
  onViewInsight?: (insight: Insight) => void
  onViewPolicy?: (policy: InsightPolicy) => void
  onDeleteInsight?: (insight: Insight) => void
  onDeletePolicy?: (policy: InsightPolicy) => void
  onTogglePolicySuspend?: (policy: InsightPolicy, suspended: boolean) => void
  className?: string
}

/**
 * DashboardOverview provides a complete dashboard view with stats and recent items.
 * Can be embedded in any portal as the main insights dashboard.
 */
export function DashboardOverview({
  stats,
  recentInsights,
  recentPolicies,
  onViewInsight,
  onViewPolicy,
  onDeleteInsight,
  onDeletePolicy,
  onTogglePolicySuspend,
  className,
}: DashboardOverviewProps) {
  // Filter critical insights for attention section
  const criticalInsights = recentInsights.filter((i) => i.spec.severity === "critical")

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Stats Overview - Single row with key metrics */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatsCard
          title="Critical"
          value={stats.criticalInsights}
          icon={AlertCircle}
          variant="critical"
        />
        <StatsCard
          title="Warnings"
          value={stats.warningInsights}
          icon={AlertTriangle}
          variant="warning"
        />
        <StatsCard
          title="Active Insights"
          value={stats.activeInsights}
          icon={Lightbulb}
        />
        <StatsCard
          title="Active Policies"
          value={stats.activePolicies}
          description={stats.suspendedPolicies > 0 ? `${stats.suspendedPolicies} suspended` : undefined}
          icon={Shield}
        />
      </div>

      {/* Critical Alerts - Only shown when there are critical insights */}
      {criticalInsights.length > 0 && (
        <Card className="border-red-200 dark:border-red-800">
          <CardHeader className="pb-3">
            <CardTitle className="flex items-center gap-2 text-base text-red-600 dark:text-red-400">
              <AlertCircle className="h-5 w-5" />
              Attention Required
            </CardTitle>
          </CardHeader>
          <CardContent>
            <ScrollArea className="max-h-64">
              <div className="grid gap-3 md:grid-cols-2">
                {criticalInsights.slice(0, 4).map((insight) => (
                  <InsightCard
                    key={`${insight.metadata.namespace}/${insight.metadata.name}`}
                    insight={insight}
                    onView={onViewInsight}
                    onDelete={onDeleteInsight}
                  />
                ))}
              </div>
            </ScrollArea>
          </CardContent>
        </Card>
      )}

      {/* Recent Items Tabs */}
      <Tabs defaultValue="insights" className="space-y-4">
        <TabsList>
          <TabsTrigger value="insights">Recent Insights</TabsTrigger>
          <TabsTrigger value="policies">Active Policies</TabsTrigger>
        </TabsList>

        <TabsContent value="insights" className="space-y-4">
          {recentInsights.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <Lightbulb className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-lg font-medium text-muted-foreground">No insights yet</p>
                <p className="text-sm text-muted-foreground">
                  Insights will appear here when created by policies or manually.
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {recentInsights.slice(0, 6).map((insight) => (
                <InsightCard
                  key={`${insight.metadata.namespace}/${insight.metadata.name}`}
                  insight={insight}
                  onView={onViewInsight}
                  onDelete={onDeleteInsight}
                />
              ))}
            </div>
          )}
        </TabsContent>

        <TabsContent value="policies" className="space-y-4">
          {recentPolicies.length === 0 ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-12">
                <ScrollText className="h-12 w-12 text-muted-foreground mb-4" />
                <p className="text-lg font-medium text-muted-foreground">No policies yet</p>
                <p className="text-sm text-muted-foreground">
                  Create a policy to automatically generate insights for your resources.
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {recentPolicies.slice(0, 6).map((policy) => (
                <PolicyCard
                  key={policy.metadata.name}
                  policy={policy}
                  onView={onViewPolicy}
                  onDelete={onDeletePolicy}
                  onToggleSuspend={onTogglePolicySuspend}
                />
              ))}
            </div>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}
