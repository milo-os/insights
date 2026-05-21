"use client"

import { useState } from "react"
import { usePolicies, useDeletePolicy, useTogglePolicySuspension } from "@/hooks/use-insights"
import { PoliciesTable } from "@/components/policies/policies-table"
import { PolicyCard } from "@/components/policies/policy-card"
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
import type { InsightPolicy } from "@/types/insights"
import { useRouter } from "next/navigation"
import { Grid3X3, List, Plus, Search } from "lucide-react"

export default function PoliciesPage() {
  const router = useRouter()
  const { data, isLoading, isError, error, refetch } = usePolicies()
  const deletePolicy = useDeletePolicy()
  const toggleSuspension = useTogglePolicySuspension()
  const [viewMode, setViewMode] = useState<"grid" | "table">("table")
  const [searchQuery, setSearchQuery] = useState("")
  const [phaseFilter, setPhaseFilter] = useState<string>("all")

  // InsightPolicies are cluster-scoped — use only the name in the route
  const handleViewPolicy = (policy: InsightPolicy) => {
    router.push(`/policies/${policy.metadata.name}`)
  }

  const handleDeletePolicy = async (policy: InsightPolicy) => {
    await deletePolicy.mutateAsync({ name: policy.metadata.name })
  }

  const handleToggleSuspend = async (policy: InsightPolicy, suspended: boolean) => {
    await toggleSuspension.mutateAsync({
      name: policy.metadata.name,
      suspended,
    })
  }

  if (isLoading) {
    return <LoadingState message="Loading policies..." />
  }

  if (isError) {
    return (
      <ErrorState
        title="Failed to load policies"
        message={error?.message || "An error occurred while loading policies."}
        onRetry={() => refetch()}
      />
    )
  }

  const policies = data?.items || []

  // Apply filters
  const filteredPolicies = policies.filter((policy) => {
    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      const matchesSearch =
        policy.metadata.name.toLowerCase().includes(query) ||
        policy.spec.selector.kind.toLowerCase().includes(query)
      if (!matchesSearch) return false
    }

    // Phase filter
    if (phaseFilter !== "all") {
      if (phaseFilter === "Suspended") {
        if (!policy.spec.suspended) return false
      } else {
        if (policy.status?.phase !== phaseFilter) return false
      }
    }

    return true
  })

  // Sort by creation date (most recent first)
  const sortedPolicies = [...filteredPolicies].sort((a, b) => {
    const dateA = new Date(a.metadata.creationTimestamp || 0).getTime()
    const dateB = new Date(b.metadata.creationTimestamp || 0).getTime()
    return dateB - dateA
  })

  const hasFilters = searchQuery !== "" || phaseFilter !== "all"

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Policies</h1>
          <p className="text-muted-foreground">
            Manage InsightPolicies that automatically generate insights.
          </p>
        </div>
        <Button onClick={() => router.push("/policies/new")}>
          <Plus className="mr-2 h-4 w-4" />
          New Policy
        </Button>
      </div>

      {/* Filters */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div className="flex flex-1 items-center gap-3 flex-wrap">
          <div className="relative flex-1 min-w-[200px] max-w-sm">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search policies..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-8"
              aria-label="Search policies"
            />
          </div>
          <Select value={phaseFilter} onValueChange={setPhaseFilter}>
            <SelectTrigger className="w-[150px]" aria-label="Filter by phase">
              <SelectValue placeholder="Phase" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Phases</SelectItem>
              <SelectItem value="Active">Active</SelectItem>
              <SelectItem value="Suspended">Suspended</SelectItem>
              <SelectItem value="Error">Error</SelectItem>
            </SelectContent>
          </Select>
          {hasFilters && (
            <Button
              variant="ghost"
              size="sm"
              onClick={() => {
                setSearchQuery("")
                setPhaseFilter("all")
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
        Showing {sortedPolicies.length} of {policies.length} policies
      </p>

      {/* Content */}
      {sortedPolicies.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          {hasFilters ? (
            <>
              <p className="font-medium text-muted-foreground">No policies match your filters</p>
              <Button
                variant="link"
                className="mt-2"
                onClick={() => {
                  setSearchQuery("")
                  setPhaseFilter("all")
                }}
              >
                Clear filters
              </Button>
            </>
          ) : (
            <>
              <p className="font-medium text-muted-foreground">No policies yet</p>
              <p className="text-sm text-muted-foreground mt-1">
                Create a policy to automatically generate insights for your resources.
              </p>
              <Button className="mt-4" onClick={() => router.push("/policies/new")}>
                <Plus className="mr-2 h-4 w-4" />
                New Policy
              </Button>
            </>
          )}
        </div>
      ) : viewMode === "table" ? (
        <PoliciesTable
          policies={sortedPolicies}
          onView={handleViewPolicy}
          onDelete={handleDeletePolicy}
          onToggleSuspend={handleToggleSuspend}
        />
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {sortedPolicies.map((policy) => (
            <PolicyCard
              key={policy.metadata.name}
              policy={policy}
              onView={handleViewPolicy}
              onDelete={handleDeletePolicy}
              onToggleSuspend={handleToggleSuspend}
            />
          ))}
        </div>
      )}
    </div>
  )
}
