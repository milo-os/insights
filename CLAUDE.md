# Insights Service

Kubernetes operator for generating and managing Insights about cluster resources using CEL-based policies.

## Project Context

| Property | Value |
|----------|-------|
| Module | `github.com/datum-cloud/insights` |
| API Group | `insights.miloapis.com` |
| Version | `v1alpha1` |
| Go Version | 1.23 |

## Resources

| Resource | Description |
|----------|-------------|
| `Insight` | Observation/finding about a target resource |
| `InsightPolicy` | CEL-based rules for generating insights |

## Tech Stack

- **Backend**: Go, controller-runtime (Kubebuilder)
- **Expression**: CEL (Common Expression Language)
- **UI**: Next.js, React, TypeScript, Tailwind, shadcn/ui
- **Deployment**: Kustomize, Kind
- **CI/CD**: GitHub Actions, Task

## Repository Structure

```
api/v1alpha1/           # API type definitions
internal/
├── controller/         # Reconcilers
└── cel/                # CEL evaluation engine
cmd/main.go             # Entry point
config/                 # Kustomize manifests
├── crd/bases/          # Generated CRDs
├── default/            # Main overlay
├── manager/            # Controller deployment
├── rbac/               # RBAC configuration
└── samples/            # Example resources
ui/                     # Next.js frontend
├── src/app/            # Pages
├── src/components/     # React components
└── src/lib/            # Utilities
.claude/                # Agent architecture
├── agents/             # Agent definitions
├── skills/             # Knowledge skills
├── pipeline/           # Feature pipeline
└── decisions/          # ADRs
```

## Key Patterns

### Controller
- Reconciler with finalizer pattern
- Status conditions via `meta.SetStatusCondition`
- ObservedGeneration tracking
- Requeue with intervals

### CEL
- Singleton environment via `sync.Once`
- Compiled programs cached
- Template syntax: `{{ expression }}`

### Types
- Kubebuilder markers for validation
- Spec/Status separation
- List types for every resource

## Build Commands

```bash
task generate     # Generate deepcopy, CRDs, RBAC
task test         # Run unit tests
task lint         # Run linter
task build        # Build binary
task run          # Run locally
task cluster-up   # Create Kind cluster
task install      # Install CRDs
task deploy       # Deploy controller
task dev          # Start Tilt environment
```

## Agent Delegation

| Task | Agent |
|------|-------|
| Feature ideas | product-discovery |
| Specs/requirements | product-planner |
| Pricing decisions | commercial-strategist |
| System design | architect |
| UI patterns | design-system |
| Go implementation | api-dev |
| React implementation | frontend-dev |
| Infrastructure | sre |
| Code review | code-reviewer |
| Testing | test-engineer |
| Documentation | tech-writer |
| Communications | gtm-comms |
| Support issues | support-triage |
| Bug investigation | debugger |

## Pipeline

Feature development follows stages in `.claude/pipeline/`. Read the `pipeline-conductor` skill for routing logic.

## Skills

| Skill | Use For |
|-------|---------|
| `k8s-apiserver-patterns` | Controller, types, CEL |
| `go-conventions` | Code style, testing |
| `kustomize-patterns` | Deployment config |
| `datum-ci` | Build, CI/CD |
| `capability-*` | Platform integrations |
| `pipeline-conductor` | Feature orchestration |
| `design-tokens` | UI patterns |
| `runbooks` | Agent-specific knowledge |

## Conventions

### Imports
Three groups: stdlib, external, internal (blank lines between)

### Tests
- File: `foo.go` → `foo_test.go`
- Framework: Ginkgo/Gomega
- Integration: envtest

### Markers
```go
// +kubebuilder:validation:Required
// +kubebuilder:subresource:status
// +kubebuilder:rbac:groups=...
```

### Conditions
- Type: `Ready`, `TargetExists`, etc.
- Reason: CamelCase single word
- Message: Descriptive text
