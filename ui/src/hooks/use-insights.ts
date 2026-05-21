"use client"

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import {
  apiClient,
  calculateDashboardStats,
  type AcknowledgeOptions,
  type UnacknowledgeOptions,
  type SnoozeOptions,
  type ResolveOptions,
  type AssignOptions,
} from "@/lib/api"
import type { InsightPolicy, InsightPolicySpec } from "@/types/insights"

/**
 * Hook to fetch all insights, optionally scoped to a namespace
 */
export function useInsights(namespace?: string) {
  return useQuery({
    queryKey: ["insights", namespace],
    queryFn: () => namespace ? apiClient.listInsights(namespace) : apiClient.listAllInsights(),
    refetchInterval: 30000,
  })
}

/**
 * Hook to fetch insights filtered by a label selector (e.g. for a policy)
 */
export function useInsightsByLabel(labelSelector: string, namespace?: string) {
  return useQuery({
    queryKey: ["insights-by-label", labelSelector, namespace],
    queryFn: () => apiClient.listInsightsByLabel(labelSelector, namespace),
    enabled: !!labelSelector,
    refetchInterval: 30000,
  })
}

/**
 * Hook to fetch a single insight
 */
export function useInsight(name: string, namespace: string) {
  return useQuery({
    queryKey: ["insight", namespace, name],
    queryFn: () => apiClient.getInsight(name, namespace),
    enabled: !!(name && namespace),
  })
}

/**
 * Hook to delete an insight
 */
export function useDeleteInsight() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ name, namespace }: { name: string; namespace: string }) =>
      apiClient.deleteInsight(name, namespace),
    onSuccess: (_, { namespace }) => {
      queryClient.invalidateQueries({ queryKey: ["insights"] })
      queryClient.invalidateQueries({ queryKey: ["insights", namespace] })
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] })
    },
  })
}

/**
 * Hook to acknowledge an insight via the /acknowledge subresource
 */
export function useAcknowledgeInsight() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      name,
      namespace,
      options,
    }: {
      name: string
      namespace: string
      options?: AcknowledgeOptions
    }) => apiClient.acknowledgeInsight(name, namespace, options),
    onSuccess: (data) => {
      const ns = data.metadata.namespace ?? ""
      queryClient.invalidateQueries({ queryKey: ["insight", ns, data.metadata.name] })
      queryClient.invalidateQueries({ queryKey: ["insights"] })
      queryClient.invalidateQueries({ queryKey: ["insights", ns] })
    },
  })
}

/**
 * Hook to unacknowledge an insight via the /unacknowledge subresource
 */
export function useUnacknowledgeInsight() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      name,
      namespace,
      options,
    }: {
      name: string
      namespace: string
      options?: UnacknowledgeOptions
    }) => apiClient.unacknowledgeInsight(name, namespace, options),
    onSuccess: (data) => {
      const ns = data.metadata.namespace ?? ""
      queryClient.invalidateQueries({ queryKey: ["insight", ns, data.metadata.name] })
      queryClient.invalidateQueries({ queryKey: ["insights"] })
      queryClient.invalidateQueries({ queryKey: ["insights", ns] })
    },
  })
}

/**
 * Hook to snooze an insight via the /snooze subresource
 */
export function useSnoozeInsight() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      name,
      namespace,
      options,
    }: {
      name: string
      namespace: string
      options: SnoozeOptions
    }) => apiClient.snoozeInsight(name, namespace, options),
    onSuccess: (data) => {
      const ns = data.metadata.namespace ?? ""
      queryClient.invalidateQueries({ queryKey: ["insight", ns, data.metadata.name] })
      queryClient.invalidateQueries({ queryKey: ["insights"] })
      queryClient.invalidateQueries({ queryKey: ["insights", ns] })
    },
  })
}

/**
 * Hook to resolve an insight via the /resolve subresource
 */
export function useResolveInsight() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      name,
      namespace,
      options,
    }: {
      name: string
      namespace: string
      options?: ResolveOptions
    }) => apiClient.resolveInsight(name, namespace, options),
    onSuccess: (data) => {
      const ns = data.metadata.namespace ?? ""
      queryClient.invalidateQueries({ queryKey: ["insight", ns, data.metadata.name] })
      queryClient.invalidateQueries({ queryKey: ["insights"] })
      queryClient.invalidateQueries({ queryKey: ["insights", ns] })
    },
  })
}

/**
 * Hook to assign an insight via the /assign subresource
 */
export function useAssignInsight() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      name,
      namespace,
      options,
    }: {
      name: string
      namespace: string
      options: AssignOptions
    }) => apiClient.assignInsight(name, namespace, options),
    onSuccess: (data) => {
      const ns = data.metadata.namespace ?? ""
      queryClient.invalidateQueries({ queryKey: ["insight", ns, data.metadata.name] })
      queryClient.invalidateQueries({ queryKey: ["insights"] })
      queryClient.invalidateQueries({ queryKey: ["insights", ns] })
    },
  })
}

/**
 * Hook to fetch all policies (cluster-scoped)
 */
export function usePolicies() {
  return useQuery({
    queryKey: ["policies"],
    queryFn: () => apiClient.listAllPolicies(),
    refetchInterval: 30000,
  })
}

/**
 * Hook to fetch a single policy (cluster-scoped)
 */
export function usePolicy(name: string) {
  return useQuery({
    queryKey: ["policy", name],
    queryFn: () => apiClient.getPolicy(name),
    enabled: !!name,
  })
}

/**
 * Hook to create a policy (cluster-scoped)
 */
export function useCreatePolicy() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (policy: { metadata: { name: string }; spec: InsightPolicySpec }) =>
      apiClient.createPolicy(policy),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["policies"] })
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] })
    },
  })
}

/**
 * Hook to update a policy (cluster-scoped)
 */
export function useUpdatePolicy() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (policy: InsightPolicy) => apiClient.updatePolicy(policy),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["policies"] })
      queryClient.invalidateQueries({ queryKey: ["policy", data.metadata.name] })
    },
  })
}

/**
 * Hook to delete a policy (cluster-scoped)
 */
export function useDeletePolicy() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ name }: { name: string }) => apiClient.deletePolicy(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["policies"] })
      queryClient.invalidateQueries({ queryKey: ["dashboard-stats"] })
    },
  })
}

/**
 * Hook to toggle policy suspension (cluster-scoped)
 */
export function useTogglePolicySuspension() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ name, suspended }: { name: string; suspended: boolean }) =>
      apiClient.togglePolicySuspension(name, suspended),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: ["policies"] })
      queryClient.invalidateQueries({ queryKey: ["policy", data.metadata.name] })
    },
  })
}

/**
 * Hook to fetch namespaces for the namespace selector
 */
export function useNamespaces() {
  return useQuery({
    queryKey: ["namespaces"],
    queryFn: async () => {
      const result = await apiClient.listNamespaces()
      return result.items.map((ns) => ns.metadata.name)
    },
    refetchInterval: 60000,
    staleTime: 30000,
  })
}

/**
 * Hook to fetch dashboard statistics
 */
export function useDashboardStats(namespace?: string) {
  const insightsQuery = useInsights(namespace)
  const policiesQuery = usePolicies()

  const isLoading = insightsQuery.isLoading || policiesQuery.isLoading
  const isError = insightsQuery.isError || policiesQuery.isError
  const error = insightsQuery.error || policiesQuery.error

  const stats = !isLoading && !isError
    ? calculateDashboardStats(
        insightsQuery.data?.items || [],
        policiesQuery.data?.items || []
      )
    : null

  return {
    data: stats,
    isLoading,
    isError,
    error,
    insights: insightsQuery.data?.items || [],
    policies: policiesQuery.data?.items || [],
  }
}
