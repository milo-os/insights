import { NextResponse } from "next/server"
import { createMockInsightList, createMockPolicyList, mockInsights, mockPolicies } from "@/lib/mock-data"

/**
 * Mock API endpoint for development without a live Kubernetes backend.
 * Enable by setting NEXT_PUBLIC_USE_MOCK_DATA=true
 */
export async function GET(request: Request) {
  const { searchParams } = new URL(request.url)
  const resource = searchParams.get("resource")
  const namespace = searchParams.get("namespace")
  const name = searchParams.get("name")

  // Simulate network delay
  await new Promise((resolve) => setTimeout(resolve, 200))

  if (resource === "insights") {
    if (name && namespace) {
      // Get single insight
      const insight = mockInsights.find(
        (i) => i.metadata.name === name && i.metadata.namespace === namespace
      )
      if (insight) {
        return NextResponse.json(insight)
      }
      return NextResponse.json({ message: "Insight not found" }, { status: 404 })
    }

    // List insights
    let filtered = mockInsights
    if (namespace) {
      filtered = filtered.filter((i) => i.metadata.namespace === namespace)
    }
    return NextResponse.json(createMockInsightList(filtered))
  }

  if (resource === "policies") {
    if (name && namespace) {
      // Get single policy
      const policy = mockPolicies.find(
        (p) => p.metadata.name === name && p.metadata.namespace === namespace
      )
      if (policy) {
        return NextResponse.json(policy)
      }
      return NextResponse.json({ message: "Policy not found" }, { status: 404 })
    }

    // List policies
    let filtered = mockPolicies
    if (namespace) {
      filtered = filtered.filter((p) => p.metadata.namespace === namespace)
    }
    return NextResponse.json(createMockPolicyList(filtered))
  }

  return NextResponse.json({ message: "Unknown resource type" }, { status: 400 })
}

export async function POST(request: Request) {
  const body = await request.json()

  // Simulate network delay
  await new Promise((resolve) => setTimeout(resolve, 300))

  // Return the created resource with generated metadata
  return NextResponse.json({
    ...body,
    metadata: {
      ...body.metadata,
      uid: `mock-${Date.now()}`,
      creationTimestamp: new Date().toISOString(),
      resourceVersion: "1",
    },
    status: body.kind === "Insight"
      ? { phase: "Active", targetExists: true }
      : { phase: "Active", matchingResourceCount: 0, totalInsightCount: 0 },
  })
}

export async function DELETE(request: Request) {
  // Simulate network delay
  await new Promise((resolve) => setTimeout(resolve, 200))

  return NextResponse.json({ status: "Success" })
}
