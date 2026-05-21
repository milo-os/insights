/**
 * Kubernetes-style metadata
 */
export interface ObjectMeta {
  name: string
  namespace?: string
  uid?: string
  resourceVersion?: string
  creationTimestamp?: string
  labels?: Record<string, string>
  annotations?: Record<string, string>
}

/**
 * Reference to a Kubernetes resource
 */
export interface TargetRef {
  apiVersion: string
  kind: string
  name: string
  namespace?: string
}

/**
 * Actor reference — a user or service account that performed an action
 */
export interface ActorReference {
  type: "user" | "serviceaccount"
  name: string
  uid?: string
  email?: string
}

/**
 * Reference to an InsightPolicy (cluster-scoped, no namespace)
 */
export interface PolicyReference {
  name: string
  ruleName: string
}

/**
 * Reference to an InsightMuteRule
 */
export interface MuteRuleReference {
  name: string
  namespace: string
}

/**
 * Source of an Insight
 */
export interface InsightSource {
  type: "Manual" | "Policy"
  component?: string
  policyRef?: PolicyReference
}

/**
 * Insight severity levels
 */
export type InsightSeverity = "info" | "warning" | "critical"

/**
 * Insight lifecycle state
 */
export type InsightState = "Active" | "Acknowledged" | "Snoozed" | "Resolved"

/**
 * Standard Kubernetes condition
 */
export interface Condition {
  type: string
  status: "True" | "False" | "Unknown"
  lastTransitionTime: string
  reason: string
  message: string
  observedGeneration?: number
}

/**
 * Acknowledgement lifecycle info
 */
export interface AcknowledgementInfo {
  by: ActorReference
  at: string
  note?: string
}

/**
 * Snooze lifecycle info
 */
export interface SnoozeInfo {
  by: ActorReference
  at: string
  until: string
}

/**
 * Assignment lifecycle info
 */
export interface AssignmentInfo {
  by: ActorReference
  to: ActorReference
  at: string
  note?: string
}

/**
 * Resolution lifecycle info
 */
export interface ResolutionInfo {
  by: ActorReference
  at: string
  note?: string
}

/**
 * InsightSpec defines the desired state of an Insight
 */
export interface InsightSpec {
  targetRef: TargetRef
  severity: InsightSeverity
  message: string
  description?: string
  category: string
  source: InsightSource
  ttlSeconds?: number
  persistent?: boolean
}

/**
 * InsightStatus defines the observed state of an Insight
 */
export interface InsightStatus {
  state?: InsightState
  owner?: ActorReference
  acknowledgement?: AcknowledgementInfo
  snooze?: SnoozeInfo
  assignment?: AssignmentInfo
  resolution?: ResolutionInfo
  targetExists?: boolean
  expiresAt?: string
  deleteAt?: string
  muted?: boolean
  mutedBy?: MuteRuleReference
  conditions?: Condition[]
  observedGeneration?: number
}

/**
 * Insight represents an observation about a Kubernetes resource
 */
export interface Insight {
  apiVersion: string
  kind: "Insight"
  metadata: ObjectMeta
  spec: InsightSpec
  status?: InsightStatus
}

/**
 * InsightList is a list of Insights
 */
export interface InsightList {
  apiVersion: string
  kind: "InsightList"
  metadata: {
    resourceVersion?: string
    continue?: string
  }
  items: Insight[]
}

/**
 * Label selector for InsightPolicy
 */
export interface LabelSelector {
  matchLabels?: Record<string, string>
  matchExpressions?: Array<{
    key: string
    operator: "In" | "NotIn" | "Exists" | "DoesNotExist"
    values?: string[]
  }>
}

/**
 * Selector for InsightPolicy
 */
export interface PolicySelector {
  apiVersion: string
  kind: string
  namespaces?: string[]
  labelSelector?: LabelSelector
  matchExpression?: string
}

/**
 * Rule within an InsightPolicy
 */
export interface InsightRule {
  name: string
  condition: string
  severity: InsightSeverity
  category: string
  messageTemplate: string
  descriptionTemplate?: string
  ttlSeconds?: number
  persistent?: boolean
}

/**
 * Status of a rule within an InsightPolicy
 */
export interface InsightRuleStatus {
  name: string
  activeInsightCount: number
  lastEvaluationTime?: string
  lastError?: string
}

/**
 * InsightPolicy phase
 */
export type InsightPolicyPhase = "Active" | "Suspended" | "Error"

/**
 * InsightPolicySpec defines the desired state of an InsightPolicy
 */
export interface InsightPolicySpec {
  selector: PolicySelector
  rules: InsightRule[]
  suspended?: boolean
  resyncPeriodSeconds?: number
  insightNamePrefix?: string
  insightNamespace?: string
}

/**
 * InsightPolicyStatus defines the observed state of an InsightPolicy
 */
export interface InsightPolicyStatus {
  phase?: InsightPolicyPhase
  matchingResourceCount?: number
  skippedResourceCount?: number
  totalInsightCount?: number
  ruleStatuses?: InsightRuleStatus[]
  lastEvaluationTime?: string
  conditions?: Condition[]
  observedGeneration?: number
}

/**
 * InsightPolicy defines rules for automatically generating Insights
 */
export interface InsightPolicy {
  apiVersion: string
  kind: "InsightPolicy"
  metadata: ObjectMeta
  spec: InsightPolicySpec
  status?: InsightPolicyStatus
}

/**
 * InsightPolicyList is a list of InsightPolicies
 */
export interface InsightPolicyList {
  apiVersion: string
  kind: "InsightPolicyList"
  metadata: {
    resourceVersion?: string
    continue?: string
  }
  items: InsightPolicy[]
}

/**
 * InsightMuteRule match criteria
 */
export interface InsightMuteMatch {
  policyRef?: PolicyReference
  category?: string
  targetRef?: TargetRef
  severity?: InsightSeverity
  labelSelector?: LabelSelector
}

/**
 * InsightMuteRuleSpec defines the desired state of an InsightMuteRule
 */
export interface InsightMuteRuleSpec {
  match: InsightMuteMatch
  reason?: string
  expiresAt?: string
}

/**
 * InsightMuteRuleStatus defines the observed state of an InsightMuteRule
 */
export interface InsightMuteRuleStatus {
  mutedInsightCount?: number
  expired?: boolean
  conditions?: Condition[]
  observedGeneration?: number
}

/**
 * InsightMuteRule suppresses matching insights from being surfaced
 */
export interface InsightMuteRule {
  apiVersion: string
  kind: "InsightMuteRule"
  metadata: ObjectMeta
  spec: InsightMuteRuleSpec
  status?: InsightMuteRuleStatus
}

/**
 * Dashboard statistics
 */
export interface DashboardStats {
  totalInsights: number
  criticalInsights: number
  warningInsights: number
  infoInsights: number
  activeInsights: number
  acknowledgedInsights: number
  snoozedInsights: number
  resolvedInsights: number
  totalPolicies: number
  activePolicies: number
  suspendedPolicies: number
  errorPolicies: number
}
