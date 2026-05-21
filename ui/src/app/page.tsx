"use client"

import { useDashboardStats, useDeleteInsight, useDeletePolicy, useTogglePolicySuspension } from "@/hooks/use-insights"
import { DashboardOverview } from "@/components/shared/dashboard-overview"
import { LoadingState } from "@/components/shared/loading-state"
import { ErrorState } from "@/components/shared/error-state"
import type { Insight, InsightPolicy } from "@/types/insights"
import { useRouter } from "next/navigation"

export default function DashboardPage() {
  const router = useRouter()
  const { data: stats, isLoading, isError, error, insights, policies } = useDashboardStats()
  const deleteInsight = useDeleteInsight()
  const deletePolicy = useDeletePolicy()
  const toggleSuspension = useTogglePolicySuspension()

  const handleViewInsight = (insight: Insight) => {
    router.push(`/insights/${insight.metadata.namespace}/${insight.metadata.name}`)
  }

  // InsightPolicies are cluster-scoped — no namespace in route
  const handleViewPolicy = (policy: InsightPolicy) => {
    router.push(`/policies/${policy.metadata.name}`)
  }

  const handleDeleteInsight = async (insight: Insight) => {
    await deleteInsight.mutateAsync({
      name: insight.metadata.name,
      namespace: insight.metadata.namespace ?? "default",
    })
  }

  const handleDeletePolicy = async (policy: InsightPolicy) => {
    await deletePolicy.mutateAsync({ name: policy.metadata.name })
  }

  const handleTogglePolicySuspend = async (policy: InsightPolicy, suspended: boolean) => {
    await toggleSuspension.mutateAsync({
      name: policy.metadata.name,
      suspended,
    })
  }

  if (isLoading) {
    return <LoadingState message="Loading dashboard..." />
  }

  if (isError || !stats) {
    return (
      <ErrorState
        title="Failed to load dashboard"
        message={error?.message || "An error occurred while loading the dashboard data."}
      />
    )
  }

  // Sort insights by creation time (most recent first)
  const sortedInsights = [...insights].sort((a, b) => {
    const dateA = new Date(a.metadata.creationTimestamp || 0).getTime()
    const dateB = new Date(b.metadata.creationTimestamp || 0).getTime()
    return dateB - dateA
  })

  // Sort policies by creation time (most recent first)
  const sortedPolicies = [...policies].sort((a, b) => {
    const dateA = new Date(a.metadata.creationTimestamp || 0).getTime()
    const dateB = new Date(b.metadata.creationTimestamp || 0).getTime()
    return dateB - dateA
  })

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
          Overview of insights and policies across your Kubernetes resources.
        </p>
      </div>

      <DashboardOverview
        stats={stats}
        recentInsights={sortedInsights}
        recentPolicies={sortedPolicies}
        onViewInsight={handleViewInsight}
        onViewPolicy={handleViewPolicy}
        onDeleteInsight={handleDeleteInsight}
        onDeletePolicy={handleDeletePolicy}
        onTogglePolicySuspend={handleTogglePolicySuspend}
      />
    </div>
  )
}
