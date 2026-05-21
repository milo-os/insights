"use client"

import { useState } from "react"
import { useInsights } from "@/hooks/use-insights"
import { InsightsTable } from "@/components/insights/insights-table"
import { InsightCard } from "@/components/insights/insight-card"
import { LoadingState } from "@/components/shared/loading-state"
import { ErrorState } from "@/components/shared/error-state"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type { Insight, InsightSeverity, InsightState } from "@/types/insights"
import { useRouter } from "next/navigation"
import { Grid3X3, List, Search } from "lucide-react"

export default function InsightsPage() {
  const router = useRouter()
  const { data, isLoading, isError, error, refetch } = useInsights()

  const [viewMode, setViewMode] = useState<"grid" | "table">("table")
  const [searchQuery, setSearchQuery] = useState("")
  const [severityFilter, setSeverityFilter] = useState<InsightSeverity | "all">("all")
  const [stateFilter, setStateFilter] = useState<InsightState | "all">("all")

  const handleViewInsight = (insight: Insight) => {
    router.push(`/insights/${insight.metadata.namespace}/${insight.metadata.name}`)
  }

  if (isLoading) {
    return <LoadingState message="Loading insights..." />
  }

  if (isError) {
    return (
      <ErrorState
        title="Failed to load insights"
        message={error?.message || "An error occurred while loading insights."}
        onRetry={() => refetch()}
      />
    )
  }

  const insights = data?.items || []

  // Apply filters using status.state
  const filteredInsights = insights.filter((insight) => {
    // State filter
    if (stateFilter !== "all") {
      const state = insight.status?.state ?? "Active"
      if (state !== stateFilter) return false
    }

    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      const matchesSearch =
        insight.metadata.name.toLowerCase().includes(query) ||
        insight.spec.message.toLowerCase().includes(query) ||
        insight.spec.category.toLowerCase().includes(query) ||
        insight.spec.targetRef.name.toLowerCase().includes(query)
      if (!matchesSearch) return false
    }

    // Severity filter
    if (severityFilter !== "all" && insight.spec.severity !== severityFilter) {
      return false
    }

    return true
  })

  // Sort by severity (critical first) then by date
  const sortedInsights = [...filteredInsights].sort((a, b) => {
    const severityOrder = { critical: 0, warning: 1, info: 2 }
    const severityDiff =
      severityOrder[a.spec.severity] - severityOrder[b.spec.severity]
    if (severityDiff !== 0) return severityDiff

    const dateA = new Date(a.metadata.creationTimestamp || 0).getTime()
    const dateB = new Date(b.metadata.creationTimestamp || 0).getTime()
    return dateB - dateA
  })

  const hasFilters = searchQuery !== "" || severityFilter !== "all" || stateFilter !== "all"

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Insights</h1>
          <p className="text-muted-foreground">
            View and manage insights about your Kubernetes resources.
          </p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div className="flex flex-1 items-center gap-3 flex-wrap">
          <div className="relative flex-1 min-w-[200px] max-w-sm">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search insights..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-8"
              aria-label="Search insights"
            />
          </div>
          <Select
            value={severityFilter}
            onValueChange={(value) => setSeverityFilter(value as InsightSeverity | "all")}
          >
            <SelectTrigger className="w-[150px]" aria-label="Filter by severity">
              <SelectValue placeholder="Severity" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Severities</SelectItem>
              <SelectItem value="critical">Critical</SelectItem>
              <SelectItem value="warning">Warning</SelectItem>
              <SelectItem value="info">Info</SelectItem>
            </SelectContent>
          </Select>
          <Select
            value={stateFilter}
            onValueChange={(value) => setStateFilter(value as InsightState | "all")}
          >
            <SelectTrigger className="w-[150px]" aria-label="Filter by state">
              <SelectValue placeholder="State" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All States</SelectItem>
              <SelectItem value="Active">Active</SelectItem>
              <SelectItem value="Acknowledged">Acknowledged</SelectItem>
              <SelectItem value="Snoozed">Snoozed</SelectItem>
              <SelectItem value="Resolved">Resolved</SelectItem>
            </SelectContent>
          </Select>
          {hasFilters && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setSearchQuery("")
                setSeverityFilter("all")
                setStateFilter("all")
              }}
            >
              Clear filters
            </Button>
          )}
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant={viewMode === "table" ? "default" : "outline"}
            size="icon"
            onClick={() => setViewMode("table")}
            aria-label="Table view"
            aria-pressed={viewMode === "table"}
          >
            <List className="h-4 w-4" />
          </Button>
          <Button
            variant={viewMode === "grid" ? "default" : "outline"}
            size="icon"
            onClick={() => setViewMode("grid")}
            aria-label="Grid view"
            aria-pressed={viewMode === "grid"}
          >
            <Grid3X3 className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Results count */}
      <p className="text-sm text-muted-foreground">
        Showing {sortedInsights.length} of {insights.length} insights
      </p>

      {/* Content */}
      {sortedInsights.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <Search className="h-10 w-10 text-muted-foreground/40 mb-3" />
          {hasFilters ? (
            <>
              <p className="font-medium text-muted-foreground">No insights match your filters</p>
              <p className="text-sm text-muted-foreground mt-1">
                Try adjusting the severity or state filter, or clear the search query.
              </p>
              <Button
                variant="link"
                className="mt-2"
                onClick={() => {
                  setSearchQuery("")
                  setSeverityFilter("all")
                  setStateFilter("all")
                }}
              >
                Clear filters
              </Button>
            </>
          ) : (
            <>
              <p className="font-medium text-muted-foreground">No insights yet</p>
              <p className="text-sm text-muted-foreground mt-1">
                Insights are generated automatically when policy conditions match.
              </p>
            </>
          )}
        </div>
      ) : viewMode === "table" ? (
        <InsightsTable
          insights={sortedInsights}
          onView={handleViewInsight}
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {sortedInsights.map((insight) => (
            <InsightCard
              key={`${insight.metadata.namespace}/${insight.metadata.name}`}
              insight={insight}
              onView={handleViewInsight}
            />
          ))}
        </div>
      )}
    </div>
  )
}
