# Security Considerations

## RBAC Permissions

The Insights controller requires broad read permissions across the cluster to evaluate InsightPolicy rules against arbitrary resource types. This is an intentional design decision to support the following use cases:

- Policies that target resources across multiple namespaces
- Policies that evaluate cluster-scoped resources (Nodes, Namespaces, PersistentVolumes, etc.)
- Cross-namespace reference validation

### Current Permissions

The controller's ServiceAccount is granted:

```yaml
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["get", "list", "watch"]
```

### Security Implications

1. **Secret Access**: The controller can read Secrets in all namespaces. While Secrets are not typically targeted by InsightPolicies, a misconfigured or malicious policy could potentially expose Secret data through Insight messages if CEL templates reference Secret fields.

2. **Information Disclosure**: CEL templates in InsightPolicy rules can extract and expose data from any readable resource. The extracted data appears in Insight `.spec.message` and `.spec.description` fields.

3. **Cluster-wide Visibility**: The controller has visibility into all cluster resources, which may be a concern in multi-tenant environments.

### Mitigation Strategies

#### For Operators

1. **Restrict InsightPolicy Creation**: Use RBAC to limit who can create InsightPolicy resources:
   ```yaml
   apiVersion: rbac.authorization.k8s.io/v1
   kind: ClusterRole
   metadata:
     name: insightpolicy-admin
   rules:
   - apiGroups: ["insights.miloapis.com"]
     resources: ["insightpolicies"]
     verbs: ["create", "update", "patch", "delete"]
   ```

2. **Policy Review Process**: Implement a review process for InsightPolicy changes, especially those targeting sensitive resource types.

3. **Network Policies**: Use the included NetworkPolicy to restrict controller network access to only the Kubernetes API server.

4. **Audit Logging**: Enable Kubernetes audit logging to track InsightPolicy creation and modifications.

#### For Policy Authors

1. **Avoid Targeting Secrets**: Do not create policies that select Secret or ConfigMap resources unless absolutely necessary.

2. **Minimize Data Exposure**: Use CEL templates carefully to avoid exposing sensitive data in Insight messages:
   ```yaml
   # Good: Reference non-sensitive metadata
   messageTemplate: "Deployment {{ object.metadata.name }} has issue"

   # Bad: Could expose sensitive data
   messageTemplate: "Config value: {{ object.data.password }}"
   ```

3. **Use Specific Selectors**: Target specific resources rather than broad selectors to minimize unintended data access.

### Future Improvements

We are considering the following enhancements:

1. **Resource Type Allowlist**: Configuration option to restrict which resource types can be targeted by policies.

2. **Admission Webhook Validation**: Reject policies that target sensitive resource types (Secrets, ServiceAccounts).

3. **Namespace-scoped InsightPolicy**: A namespace-scoped variant that can only target resources within its own namespace.

4. **CEL Expression Sandboxing**: Additional restrictions on CEL expressions to prevent accessing certain fields.

## Pod Security

The controller runs with the following security context:

- `runAsNonRoot: true`
- `readOnlyRootFilesystem: true`
- `allowPrivilegeEscalation: false`
- All capabilities dropped
- Seccomp profile: RuntimeDefault

This configuration meets the Kubernetes "restricted" Pod Security Standard.

## Network Security

When enabled, NetworkPolicy restricts:
- Ingress: Only from pods with `metrics: enabled` label (for Prometheus scraping)
- Egress: Only to Kubernetes API server

## Reporting Security Issues

Please report security vulnerabilities to security@datum.net. Do not open public issues for security concerns.
