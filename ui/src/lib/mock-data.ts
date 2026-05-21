import type { Insight, InsightPolicy, InsightList, InsightPolicyList } from "@/types/insights"

/**
 * Sample insights for demonstration and testing
 */
export const mockInsights: Insight[] = [
  {
    apiVersion: "insights.miloapis.com/v1alpha1",
    kind: "Insight",
    metadata: {
      name: "deploy-nginx-no-limits",
      namespace: "default",
      creationTimestamp: new Date(Date.now() - 1000 * 60 * 30).toISOString(), // 30 minutes ago
      uid: "insight-1",
    },
    spec: {
      targetRef: {
        apiVersion: "apps/v1",
        kind: "Deployment",
        name: "nginx",
        namespace: "default",
      },
      severity: "warning",
      message: "Deployment nginx has no resource limits configured",
      description: `The deployment does not have CPU/memory limits configured.
This can lead to resource contention and affect other workloads.

Consider adding resource limits:
\`\`\`yaml
resources:
  limits:
    cpu: "500m"
    memory: "128Mi"
\`\`\``,
      category: "configuration",
      source: {
        type: "Policy",
        policyRef: {
          name: "deployment-best-practices",
          ruleName: "no-resource-limits",
        },
      },
      ttlSeconds: 86400,
    },
    status: {
      state: "Active",
      targetExists: true,
      observedGeneration: 1,
      expiresAt: new Date(Date.now() + 1000 * 60 * 60 * 23).toISOString(), // 23 hours from now
      conditions: [
        {
          type: "Ready",
          status: "True",
          lastTransitionTime: new Date(Date.now() - 1000 * 60 * 30).toISOString(),
          reason: "InsightActive",
          message: "Insight is active and target resource exists",
        },
      ],
    },
  },
  {
    apiVersion: "insights.miloapis.com/v1alpha1",
    kind: "Insight",
    metadata: {
      name: "deploy-api-single-replica",
      namespace: "production",
      creationTimestamp: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(), // 2 hours ago
      uid: "insight-2",
    },
    spec: {
      targetRef: {
        apiVersion: "apps/v1",
        kind: "Deployment",
        name: "api-server",
        namespace: "production",
      },
      severity: "critical",
      message: "Production deployment api-server has only 1 replica",
      description: `The deployment api-server in production namespace is configured with only a single replica.
This is a critical availability concern for production workloads.

Consider increasing replicas to at least 2 for high availability.`,
      category: "availability",
      source: {
        type: "Policy",
        policyRef: {
          name: "production-requirements",
          ruleName: "minimum-replicas",
        },
      },
      ttlSeconds: 3600,
    },
    status: {
      state: "Active",
      targetExists: true,
      observedGeneration: 1,
      conditions: [
        {
          type: "Ready",
          status: "True",
          lastTransitionTime: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(),
          reason: "InsightActive",
          message: "Insight is active and target resource exists",
        },
      ],
    },
  },
  {
    apiVersion: "insights.miloapis.com/v1alpha1",
    kind: "Insight",
    metadata: {
      name: "pod-privileged-container",
      namespace: "default",
      creationTimestamp: new Date(Date.now() - 1000 * 60 * 60 * 5).toISOString(), // 5 hours ago
      uid: "insight-3",
    },
    spec: {
      targetRef: {
        apiVersion: "v1",
        kind: "Pod",
        name: "debug-pod",
        namespace: "default",
      },
      severity: "critical",
      message: "Pod debug-pod is running with privileged containers",
      description: `This pod is running with privileged containers which poses a security risk.
Privileged containers can access host resources and break container isolation.`,
      category: "security",
      source: {
        type: "Manual",
        component: "security-scanner",
      },
    },
    status: {
      state: "Active",
      targetExists: true,
      observedGeneration: 1,
    },
  },
  {
    apiVersion: "insights.miloapis.com/v1alpha1",
    kind: "Insight",
    metadata: {
      name: "service-no-probes",
      namespace: "staging",
      creationTimestamp: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(), // 1 day ago
      uid: "insight-4",
    },
    spec: {
      targetRef: {
        apiVersion: "apps/v1",
        kind: "Deployment",
        name: "web-frontend",
        namespace: "staging",
      },
      severity: "info",
      message: "Deployment web-frontend has no readiness or liveness probes",
      description: `The deployment does not have health check probes configured.
Consider adding probes for better availability management.`,
      category: "configuration",
      source: {
        type: "Policy",
        policyRef: {
          name: "deployment-best-practices",
          ruleName: "missing-probes",
        },
      },
      ttlSeconds: 86400,
    },
    status: {
      state: "Resolved",
      targetExists: true,
      observedGeneration: 1,
      expiresAt: new Date(Date.now() - 1000 * 60 * 60).toISOString(), // 1 hour ago
    },
  },
]

/**
 * Sample policies for demonstration and testing
 */
export const mockPolicies: InsightPolicy[] = [
  {
    apiVersion: "insights.miloapis.com/v1alpha1",
    kind: "InsightPolicy",
    metadata: {
      name: "deployment-best-practices",
      namespace: "default",
      creationTimestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 * 7).toISOString(), // 1 week ago
      uid: "policy-1",
    },
    spec: {
      selector: {
        apiVersion: "apps/v1",
        kind: "Deployment",
        namespaces: ["default", "staging"],
      },
      rules: [
        {
          name: "no-resource-limits",
          condition: `!has(object.spec.template.spec.containers) ||
object.spec.template.spec.containers.size() == 0 ||
!has(object.spec.template.spec.containers[0].resources) ||
!has(object.spec.template.spec.containers[0].resources.limits)`,
          severity: "warning",
          category: "configuration",
          messageTemplate: "Deployment {{ object.metadata.name }} has no resource limits",
          descriptionTemplate: `The deployment {{ object.metadata.name }} in namespace {{ object.metadata.namespace }}
does not have resource limits configured.

This can lead to:
- Resource contention with other workloads
- Unpredictable application behavior under load
- Difficulty in capacity planning`,
          ttlSeconds: 3600,
        },
        {
          name: "single-replica",
          condition: "has(object.spec.replicas) && object.spec.replicas == 1",
          severity: "info",
          category: "availability",
          messageTemplate: "Deployment {{ object.metadata.name }} has only 1 replica",
          descriptionTemplate: `The deployment {{ object.metadata.name }} is configured with a single replica.
Consider increasing replicas for better availability.
Current replicas: {{ object.spec.replicas }}`,
          ttlSeconds: 3600,
        },
      ],
      suspended: false,
      resyncPeriodSeconds: 300,
      insightNamePrefix: "best-practices",
    },
    status: {
      phase: "Active",
      matchingResourceCount: 12,
      totalInsightCount: 4,
      ruleStatuses: [
        {
          name: "no-resource-limits",
          activeInsightCount: 3,
        },
        {
          name: "single-replica",
          activeInsightCount: 1,
        },
      ],
      lastEvaluationTime: new Date(Date.now() - 1000 * 60 * 5).toISOString(), // 5 minutes ago
      observedGeneration: 2,
      conditions: [
        {
          type: "Ready",
          status: "True",
          lastTransitionTime: new Date(Date.now() - 1000 * 60 * 60 * 24).toISOString(),
          reason: "PolicyActive",
          message: "Policy is actively evaluating resources",
        },
      ],
    },
  },
  {
    apiVersion: "insights.miloapis.com/v1alpha1",
    kind: "InsightPolicy",
    metadata: {
      name: "production-requirements",
      namespace: "production",
      creationTimestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 * 3).toISOString(), // 3 days ago
      uid: "policy-2",
    },
    spec: {
      selector: {
        apiVersion: "apps/v1",
        kind: "Deployment",
        namespaces: ["production"],
      },
      rules: [
        {
          name: "minimum-replicas",
          condition: "!has(object.spec.replicas) || object.spec.replicas < 2",
          severity: "critical",
          category: "availability",
          messageTemplate: "Production deployment {{ object.metadata.name }} has insufficient replicas",
          descriptionTemplate: `Production workloads should have at least 2 replicas for high availability.
Current replicas: {{ has(object.spec.replicas) ? object.spec.replicas : 1 }}`,
        },
      ],
      suspended: false,
      resyncPeriodSeconds: 120,
    },
    status: {
      phase: "Active",
      matchingResourceCount: 5,
      totalInsightCount: 2,
      ruleStatuses: [
        {
          name: "minimum-replicas",
          activeInsightCount: 2,
        },
      ],
      lastEvaluationTime: new Date(Date.now() - 1000 * 60 * 2).toISOString(),
      observedGeneration: 1,
    },
  },
  {
    apiVersion: "insights.miloapis.com/v1alpha1",
    kind: "InsightPolicy",
    metadata: {
      name: "security-checks",
      namespace: "default",
      creationTimestamp: new Date(Date.now() - 1000 * 60 * 60 * 24 * 5).toISOString(), // 5 days ago
      uid: "policy-3",
    },
    spec: {
      selector: {
        apiVersion: "v1",
        kind: "Pod",
      },
      rules: [
        {
          name: "privileged-container",
          condition: `object.spec.containers.exists(c,
  has(c.securityContext) &&
  has(c.securityContext.privileged) &&
  c.securityContext.privileged == true
)`,
          severity: "critical",
          category: "security",
          messageTemplate: "Pod {{ object.metadata.name }} has privileged containers",
          descriptionTemplate: `This pod is running with privileged containers.
Privileged containers can access host resources and break container isolation.`,
        },
      ],
      suspended: true,
      resyncPeriodSeconds: 300,
    },
    status: {
      phase: "Suspended",
      matchingResourceCount: 0,
      totalInsightCount: 0,
      ruleStatuses: [],
      observedGeneration: 1,
    },
  },
]

/**
 * Create mock InsightList response
 */
export function createMockInsightList(insights: Insight[] = mockInsights): InsightList {
  return {
    apiVersion: "insights.miloapis.com/v1alpha1",
    kind: "InsightList",
    metadata: {
      resourceVersion: "12345",
    },
    items: insights,
  }
}

/**
 * Create mock InsightPolicyList response
 */
export function createMockPolicyList(policies: InsightPolicy[] = mockPolicies): InsightPolicyList {
  return {
    apiVersion: "insights.miloapis.com/v1alpha1",
    kind: "InsightPolicyList",
    metadata: {
      resourceVersion: "12345",
    },
    items: policies,
  }
}
