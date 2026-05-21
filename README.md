# insights

A Kubernetes operator for generating and managing Insights about cluster resources using CEL-based policies.

## Overview

The Insights service evaluates CEL (Common Expression Language) expressions against Kubernetes resources and surfaces findings as `Insight` objects. Policies are defined via `InsightPolicy` resources.

| Resource | Description |
|----------|-------------|
| `InsightPolicy` | CEL-based rules for generating insights |
| `Insight` | Observation or finding about a target resource |

## Quick Start

```bash
# Create a Kind cluster and install CRDs
task cluster-up
task install

# Run the controller locally
task run

# Start full dev environment (Tilt)
task dev
```

## Requirements

- Go 1.23+
- kubectl
- Kind (for local development)
- Task (`brew install go-task`)

## License

Apache License 2.0
