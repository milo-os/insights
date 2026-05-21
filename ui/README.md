# Datum Insights UI

A Next.js-based UI for managing Kubernetes Insights and InsightPolicies. This UI demonstrates all functionality exposed by the Insights API and provides reusable components that can be integrated into other UI portals.

## Features

- **Dashboard Overview**: View statistics and recent insights/policies at a glance
- **Insights Management**: List, view, filter, and delete insights
- **Policy Management**: List, view, create, suspend, and delete InsightPolicies
- **Dark/Light Theme**: Toggle between themes with Datum branding
- **Responsive Design**: Works on desktop and mobile devices

## Architecture

```
ui/
├── src/
│   ├── app/                    # Next.js App Router pages
│   │   ├── insights/           # Insights list and detail pages
│   │   ├── policies/           # Policies list, detail, and create pages
│   │   ├── layout.tsx          # Root layout with header/footer
│   │   └── page.tsx            # Dashboard home page
│   ├── components/
│   │   ├── ui/                 # shadcn/ui base components
│   │   ├── insights/           # Insight-specific components
│   │   │   ├── severity-badge.tsx
│   │   │   ├── phase-badge.tsx
│   │   │   ├── target-ref-display.tsx
│   │   │   ├── insight-card.tsx
│   │   │   ├── insight-detail.tsx
│   │   │   └── insights-table.tsx
│   │   ├── policies/           # Policy-specific components
│   │   │   ├── policy-card.tsx
│   │   │   ├── policy-detail.tsx
│   │   │   └── policies-table.tsx
│   │   └── shared/             # Shared/reusable components
│   │       ├── stats-card.tsx
│   │       ├── dashboard-overview.tsx
│   │       ├── theme-toggle.tsx
│   │       ├── namespace-selector.tsx
│   │       ├── loading-state.tsx
│   │       └── error-state.tsx
│   ├── hooks/                  # React Query hooks for API calls
│   │   └── use-insights.ts
│   ├── lib/
│   │   ├── api.ts              # API client for Insights/Policies
│   │   ├── mock-data.ts        # Mock data for development
│   │   └── utils.ts            # Utility functions
│   └── types/
│       └── insights.ts         # TypeScript types for API resources
├── tailwind.config.ts          # Tailwind config with Datum brand colors
└── package.json
```

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn
- Access to a Kubernetes cluster with the Insights controller installed

### Installation

```bash
cd ui
npm install
```

### Development

1. Start kubectl proxy to access the Kubernetes API:
   ```bash
   kubectl proxy --port=8001
   ```

2. Copy the environment file:
   ```bash
   cp .env.example .env.local
   ```

3. Start the development server:
   ```bash
   npm run dev
   ```

4. Open [http://localhost:3000](http://localhost:3000)

### Production Build

```bash
npm run build
npm start
```

The build is optimized for self-hosting with the `standalone` output mode.

## Reusable Components

The components in this UI are designed to be reusable in other portal applications.

### Insight Components

```tsx
import {
  SeverityBadge,
  PhaseBadge,
  TargetRefDisplay,
  InsightCard,
  InsightDetail,
  InsightsTable,
} from "@/components/insights"

// Display a severity badge
<SeverityBadge severity="warning" />

// Display insight cards in a grid
{insights.map(insight => (
  <InsightCard
    insight={insight}
    onView={handleView}
    onDelete={handleDelete}
  />
))}
```

### Policy Components

```tsx
import {
  PolicyCard,
  PolicyDetail,
  PoliciesTable,
} from "@/components/policies"

// Display policy cards with suspend toggle
{policies.map(policy => (
  <PolicyCard
    policy={policy}
    onView={handleView}
    onToggleSuspend={handleToggle}
  />
))}
```

### Shared Components

```tsx
import {
  StatsCard,
  DashboardOverview,
  ThemeToggle,
  NamespaceSelector,
  LoadingState,
  ErrorState,
} from "@/components/shared"

// Full dashboard view
<DashboardOverview
  stats={dashboardStats}
  recentInsights={insights}
  recentPolicies={policies}
  onViewInsight={handleViewInsight}
  onViewPolicy={handleViewPolicy}
/>
```

## Theming

The UI uses Datum's brand colors defined in `tailwind.config.ts`:

| Color | Light Theme | Dark Theme | Hex |
|-------|-------------|------------|-----|
| Primary | Midnight Fjord | Aurora Moss | #0C1D31 / #E6F59F |
| Secondary | Pine Forge | Pine Forge | #4D6356 |
| Accent | Aurora Moss | Blush Quartz | #E6F59F / #ECD0D0 |
| Background | Glacier Mist | Midnight Fjord | #E8E7E4 / #0C1D31 |

### Customizing the Theme

Edit `src/app/globals.css` to modify CSS variables:

```css
:root {
  --primary: 212 60% 12%;     /* Midnight Fjord */
  --accent: 68 80% 79%;       /* Aurora Moss */
  /* ... */
}

.dark {
  --primary: 68 80% 79%;      /* Aurora Moss */
  --background: 212 60% 12%;  /* Midnight Fjord */
  /* ... */
}
```

## API Integration

The UI communicates with the Kubernetes API server to manage Insights and InsightPolicies. The API client is configured in `src/lib/api.ts`.

### Configuration

Set the API base URL via environment variable:

```env
NEXT_PUBLIC_API_BASE_URL=http://localhost:8001
```

For in-cluster deployment, use a Kubernetes service account with appropriate RBAC:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: insights-ui
rules:
  - apiGroups: ["insights.miloapis.com"]
    resources: ["insights", "insightpolicies"]
    verbs: ["get", "list", "watch", "create", "update", "delete"]
```

## Deployment

### Docker

```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:18-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
COPY --from=builder /app/public ./public
EXPOSE 3000
CMD ["node", "server.js"]
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: insights-ui
spec:
  replicas: 2
  selector:
    matchLabels:
      app: insights-ui
  template:
    metadata:
      labels:
        app: insights-ui
    spec:
      containers:
        - name: insights-ui
          image: your-registry/insights-ui:latest
          ports:
            - containerPort: 3000
          env:
            - name: NEXT_PUBLIC_API_BASE_URL
              value: "/api/k8s"
```

## License

See the main project license.
