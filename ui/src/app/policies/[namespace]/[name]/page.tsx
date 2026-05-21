"use client"

import { useEffect } from "react"
import { useRouter, useParams } from "next/navigation"

/**
 * Legacy redirect — InsightPolicies are cluster-scoped.
 * This route was wrong; redirect to the correct /policies/[name] route.
 */
export default function LegacyPolicyDetailPage() {
  const router = useRouter()
  const params = useParams()
  const name = params.name as string

  useEffect(() => {
    router.replace(`/policies/${name}`)
  }, [router, name])

  return null
}
