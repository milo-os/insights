import type {
  Insight,
  InsightList,
  InsightPolicy,
  InsightPolicyList,
  InsightSpec,
  InsightPolicySpec,
  DashboardStats,
} from "@/types/insights"

/**
 * Configuration for the Insights API client
 */
export interface ApiConfig {
  baseUrl: string
  token?: string
  namespace?: string
}

/**
 * Default API configuration - can be overridden via environment variables
 */
const defaultConfig: ApiConfig = {
  baseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || "/api/k8s",
  namespace: process.env.NEXT_PUBLIC_DEFAULT_NAMESPACE || "default",
}

/**
 * Create headers for API requests
 */
function createHeaders(config: ApiConfig): HeadersInit {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  }
  if (config.token) {
    headers["Authorization"] = `Bearer ${config.token}`
  }
  return headers
}

/**
 * Handle API response and errors
 */
async function handleResponse<T>(response: Response): Promise<T> {
  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: response.statusText }))
    throw new Error(error.message || `API error: ${response.status}`)
  }
  return response.json()
}

/**
 * Request bodies for lifecycle subresource actions
 */
export interface AcknowledgeOptions {
  note?: string
}

export interface UnacknowledgeOptions {
  note?: string
}

export interface SnoozeOptions {
  duration?: string
  until?: string
}

export interface ResolveOptions {
  note?: string
}

export interface AssignOptions {
  assignee: {
    type: "user" | "serviceaccount"
    name: string
    email?: string
  }
  note?: string
}

/**
 * Namespace list from the Kubernetes API
 */
export interface NamespaceList {
  items: Array<{ metadata: { name: string } }>
}

/**
 * Insights API client
 */
export class InsightsApiClient {
  private config: ApiConfig

  constructor(config: Partial<ApiConfig> = {}) {
    this.config = { ...defaultConfig, ...config }
  }

  /**
   * List all Insights in a namespace
   */
  async listInsights(namespace: string): Promise<InsightList> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/namespaces/${namespace}/insights`,
      { headers: createHeaders(this.config) }
    )
    return handleResponse<InsightList>(response)
  }

  /**
   * List all Insights across all namespaces
   */
  async listAllInsights(): Promise<InsightList> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/insights`,
      { headers: createHeaders(this.config) }
    )
    return handleResponse<InsightList>(response)
  }

  /**
   * List insights filtered by a label selector
   */
  async listInsightsByLabel(labelSelector: string, namespace?: string): Promise<InsightList> {
    const path = namespace
      ? `/apis/insights.miloapis.com/v1alpha1/namespaces/${namespace}/insights`
      : `/apis/insights.miloapis.com/v1alpha1/insights`
    const url = `${this.config.baseUrl}${path}?labelSelector=${encodeURIComponent(labelSelector)}`
    const response = await fetch(url, { headers: createHeaders(this.config) })
    return handleResponse<InsightList>(response)
  }

  /**
   * Get a specific Insight
   */
  async getInsight(name: string, namespace: string): Promise<Insight> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/namespaces/${namespace}/insights/${name}`,
      { headers: createHeaders(this.config) }
    )
    return handleResponse<Insight>(response)
  }

  /**
   * Create a new Insight
   */
  async createInsight(insight: { metadata: { name: string; namespace: string }; spec: InsightSpec }): Promise<Insight> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/namespaces/${insight.metadata.namespace}/insights`,
      {
        method: "POST",
        headers: createHeaders(this.config),
        body: JSON.stringify({
          apiVersion: "insights.miloapis.com/v1alpha1",
          kind: "Insight",
          ...insight,
        }),
      }
    )
    return handleResponse<Insight>(response)
  }

  /**
   * Delete an Insight
   */
  async deleteInsight(name: string, namespace: string): Promise<void> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/namespaces/${namespace}/insights/${name}`,
      {
        method: "DELETE",
        headers: createHeaders(this.config),
      }
    )
    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: response.statusText }))
      throw new Error(error.message || `Failed to delete insight: ${response.status}`)
    }
  }

  /**
   * Acknowledge an Insight via subresource endpoint
   */
  async acknowledgeInsight(name: string, namespace: string, options: AcknowledgeOptions = {}): Promise<Insight> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/namespaces/${namespace}/insights/${name}/acknowledge`,
      {
        method: "POST",
        headers: createHeaders(this.config),
        body: JSON.stringify(options),
      }
    )
    return handleResponse<Insight>(response)
  }

  /**
   * Unacknowledge an Insight via subresource endpoint
   */
  async unacknowledgeInsight(name: string, namespace: string, options: UnacknowledgeOptions = {}): Promise<Insight> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/namespaces/${namespace}/insights/${name}/unacknowledge`,
      {
        method: "POST",
        headers: createHeaders(this.config),
        body: JSON.stringify(options),
      }
    )
    return handleResponse<Insight>(response)
  }

  /**
   * Snooze an Insight via subresource endpoint
   */
  async snoozeInsight(name: string, namespace: string, options: SnoozeOptions): Promise<Insight> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/namespaces/${namespace}/insights/${name}/snooze`,
      {
        method: "POST",
        headers: createHeaders(this.config),
        body: JSON.stringify(options),
      }
    )
    return handleResponse<Insight>(response)
  }

  /**
   * Resolve an Insight via subresource endpoint
   */
  async resolveInsight(name: string, namespace: string, options: ResolveOptions = {}): Promise<Insight> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/namespaces/${namespace}/insights/${name}/resolve`,
      {
        method: "POST",
        headers: createHeaders(this.config),
        body: JSON.stringify(options),
      }
    )
    return handleResponse<Insight>(response)
  }

  /**
   * Assign an Insight via subresource endpoint
   */
  async assignInsight(name: string, namespace: string, options: AssignOptions): Promise<Insight> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/namespaces/${namespace}/insights/${name}/assign`,
      {
        method: "POST",
        headers: createHeaders(this.config),
        body: JSON.stringify(options),
      }
    )
    return handleResponse<Insight>(response)
  }

  /**
   * List all InsightPolicies (cluster-scoped — no namespace segment)
   */
  async listAllPolicies(): Promise<InsightPolicyList> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/insightpolicies`,
      { headers: createHeaders(this.config) }
    )
    return handleResponse<InsightPolicyList>(response)
  }

  /**
   * Get a specific InsightPolicy (cluster-scoped)
   */
  async getPolicy(name: string): Promise<InsightPolicy> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/insightpolicies/${name}`,
      { headers: createHeaders(this.config) }
    )
    return handleResponse<InsightPolicy>(response)
  }

  /**
   * Create a new InsightPolicy (cluster-scoped)
   */
  async createPolicy(policy: { metadata: { name: string }; spec: InsightPolicySpec }): Promise<InsightPolicy> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/insightpolicies`,
      {
        method: "POST",
        headers: createHeaders(this.config),
        body: JSON.stringify({
          apiVersion: "insights.miloapis.com/v1alpha1",
          kind: "InsightPolicy",
          ...policy,
        }),
      }
    )
    return handleResponse<InsightPolicy>(response)
  }

  /**
   * Update an InsightPolicy (cluster-scoped)
   */
  async updatePolicy(policy: InsightPolicy): Promise<InsightPolicy> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/insightpolicies/${policy.metadata.name}`,
      {
        method: "PUT",
        headers: createHeaders(this.config),
        body: JSON.stringify(policy),
      }
    )
    return handleResponse<InsightPolicy>(response)
  }

  /**
   * Delete an InsightPolicy (cluster-scoped)
   */
  async deletePolicy(name: string): Promise<void> {
    const response = await fetch(
      `${this.config.baseUrl}/apis/insights.miloapis.com/v1alpha1/insightpolicies/${name}`,
      {
        method: "DELETE",
        headers: createHeaders(this.config),
      }
    )
    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: response.statusText }))
      throw new Error(error.message || `Failed to delete policy: ${response.status}`)
    }
  }

  /**
   * Suspend or resume an InsightPolicy (cluster-scoped)
   */
  async togglePolicySuspension(name: string, suspended: boolean): Promise<InsightPolicy> {
    const policy = await this.getPolicy(name)
    policy.spec.suspended = suspended
    return this.updatePolicy(policy)
  }

  /**
   * List all namespaces
   */
  async listNamespaces(): Promise<NamespaceList> {
    const response = await fetch(
      `${this.config.baseUrl}/api/v1/namespaces`,
      { headers: createHeaders(this.config) }
    )
    return handleResponse<NamespaceList>(response)
  }
}

/**
 * Calculate dashboard statistics from insights and policies
 */
export function calculateDashboardStats(
  insights: Insight[],
  policies: InsightPolicy[]
): DashboardStats {
  const stats: DashboardStats = {
    totalInsights: insights.length,
    criticalInsights: 0,
    warningInsights: 0,
    infoInsights: 0,
    activeInsights: 0,
    acknowledgedInsights: 0,
    snoozedInsights: 0,
    resolvedInsights: 0,
    totalPolicies: policies.length,
    activePolicies: 0,
    suspendedPolicies: 0,
    errorPolicies: 0,
  }

  for (const insight of insights) {
    switch (insight.spec.severity) {
      case "critical":
        stats.criticalInsights++
        break
      case "warning":
        stats.warningInsights++
        break
      case "info":
        stats.infoInsights++
        break
    }

    switch (insight.status?.state) {
      case "Active":
        stats.activeInsights++
        break
      case "Acknowledged":
        stats.acknowledgedInsights++
        break
      case "Snoozed":
        stats.snoozedInsights++
        break
      case "Resolved":
        stats.resolvedInsights++
        break
    }
  }

  for (const policy of policies) {
    switch (policy.status?.phase) {
      case "Active":
        stats.activePolicies++
        break
      case "Suspended":
        stats.suspendedPolicies++
        break
      case "Error":
        stats.errorPolicies++
        break
    }
  }

  return stats
}

/**
 * Default API client instance
 */
export const apiClient = new InsightsApiClient()
