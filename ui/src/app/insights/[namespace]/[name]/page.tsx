"use client"

import {
  useInsight,
  useDeleteInsight,
  useAcknowledgeInsight,
  useUnacknowledgeInsight,
  useSnoozeInsight,
  useResolveInsight,
  useAssignInsight,
} from "@/hooks/use-insights"
import { InsightDetail } from "@/components/insights/insight-detail"
import { LoadingState } from "@/components/shared/loading-state"
import { ErrorState } from "@/components/shared/error-state"
import { useRouter, useParams } from "next/navigation"
import type { ActorReference } from "@/types/insights"
import type { SnoozeOptions } from "@/lib/api"

export default function InsightDetailPage() {
  const router = useRouter()
  const params = useParams()
  const namespace = params.namespace as string
  const name = params.name as string

  const { data: insight, isLoading, isError, error, refetch } = useInsight(name, namespace)

  const deleteInsight = useDeleteInsight()
  const acknowledgeInsight = useAcknowledgeInsight()
  const unacknowledgeInsight = useUnacknowledgeInsight()
  const snoozeInsight = useSnoozeInsight()
  const resolveInsight = useResolveInsight()
  const assignInsight = useAssignInsight()

  const handleBack = () => {
    router.push("/insights")
  }

  const handleDelete = async () => {
    await deleteInsight.mutateAsync({ name, namespace })
    router.push("/insights")
  }

  const handleAcknowledge = async (note?: string) => {
    await acknowledgeInsight.mutateAsync({ name, namespace, options: { note } })
  }

  const handleUnacknowledge = async () => {
    await unacknowledgeInsight.mutateAsync({ name, namespace })
  }

  const handleSnooze = async (options: SnoozeOptions) => {
    await snoozeInsight.mutateAsync({ name, namespace, options })
  }

  const handleResolve = async (note?: string) => {
    await resolveInsight.mutateAsync({ name, namespace, options: { note } })
  }

  const handleAssign = async (
    assignee: Pick<ActorReference, "type" | "name">,
    note?: string
  ) => {
    await assignInsight.mutateAsync({
      name,
      namespace,
      options: { assignee, note },
    })
  }

  if (isLoading) {
    return <LoadingState message="Loading insight..." />
  }

  if (isError || !insight) {
    return (
      <ErrorState
        title="Failed to load insight"
        message={error?.message || "The insight could not be found."}
        onRetry={() => refetch()}
      />
    )
  }

  return (
    <InsightDetail
      insight={insight}
      onBack={handleBack}
      onDelete={handleDelete}
      onAcknowledge={handleAcknowledge}
      onUnacknowledge={handleUnacknowledge}
      onSnooze={handleSnooze}
      onResolve={handleResolve}
      onAssign={handleAssign}
    />
  )
}
