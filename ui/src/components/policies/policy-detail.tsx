"use client"

import type { InsightPolicy } from "@/types/insights"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { SeverityBadge } from "@/components/insights/severity-badge"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Switch } from "@/components/ui/switch"
import { Separator } from "@/components/ui/separator"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { PolicyInsightsTab } from "./policy-insights-tab"
import { formatDate, formatRelativeTime } from "@/lib/utils"
import {
  ArrowLeft,
  Calendar,
  Clock,
  Code2,
  FileCode,
  Filter,
  Infinity,
  RefreshCw,
  Timer,
  CheckCircle,
  XCircle,
  MinusCircle,
} from "lucide-react"

interface PolicyDetailProps {
  policy: InsightPolicy
  onBack?: () => void
  onDelete?: (policy: InsightPolicy) => void
  onToggleSuspend?: (policy: InsightPolicy, suspended: boolean) => void
  className?: string
}

function PhaseBadge({ phase }: { phase?: string }) {
  if (!phase) return <Badge variant="outline">Unknown</Badge>

  const variantMap: Record<string, string> = {
    Active: "border-green-400 bg-green-100 text-green-700 dark:border-green-600 dark:bg-green-900/40 dark:text-green-300",
    Suspended: "border-slate-400 bg-slate-100 text-slate-700 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-300",
    Error: "border-red-400 bg-red-100 text-red-700 dark:border-red-600 dark:bg-red-900/40 dark:text-red-300",
  }

  return (
    <Badge variant="outline" className={variantMap[phase] ?? ""}>
      {phase}
    </Badge>
  )
}

function RuleStatusIcon({ error }: { error?: string }) {
  if (error) return <XCircle className="h-4 w-4 text-red-500" />
  return <CheckCircle className="h-4 w-4 text-green-500" />
}

/**
 * PolicyDetail displays the full details of an InsightPolicy.
 * Used in detail pages or slide-over panels.
 */
export function PolicyDetail({
  policy,
  onBack,
  onDelete,
  onToggleSuspend,
  className,
}: PolicyDetailProps) {
  const isSuspended = policy.spec.suspended || false

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          {onBack && (
            <Button variant="ghost" size="sm" onClick={onBack} className="mb-2 -ml-2">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Back to Policies
            </Button>
          )}
          <h1 className="text-2xl font-bold">{policy.metadata.name}</h1>
          <p className="text-sm text-muted-foreground font-mono">
            insights.miloapis.com/v1alpha1/insightpolicies/{policy.metadata.name}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <PhaseBadge phase={policy.status?.phase} />
        </div>
      </div>

      {/* Suspend toggle */}
      {onToggleSuspend && (
        <Card>
          <CardContent className="flex items-center justify-between py-4">
            <div>
              <p className="font-medium">Policy Status</p>
              <p className="text-sm text-muted-foreground">
                {isSuspended
                  ? "Policy is suspended and not evaluating resources"
                  : "Policy is active and evaluating resources"}
              </p>
            </div>
            <div className="flex items-center gap-3">
              <span className="text-sm text-muted-foreground">
                {isSuspended ? "Suspended" : "Active"}
              </span>
              <Switch
                checked={!isSuspended}
                onCheckedChange={(checked) => onToggleSuspend(policy, !checked)}
                aria-label="Toggle policy suspension"
              />
            </div>
          </CardContent>
        </Card>
      )}

      {/* Selector */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <Filter className="h-4 w-4" />
            Resource Selector
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center gap-2">
            <FileCode className="h-4 w-4 text-muted-foreground" />
            <Badge variant="outline" className="font-mono">
              {policy.spec.selector.apiVersion}
            </Badge>
            <Badge variant="secondary" className="font-mono">
              {policy.spec.selector.kind}
            </Badge>
          </div>

          {policy.spec.selector.namespaces && policy.spec.selector.namespaces.length > 0 && (
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">Namespaces:</p>
              <div className="flex flex-wrap gap-1">
                {policy.spec.selector.namespaces.map((ns) => (
                  <Badge key={ns} variant="outline" className="font-mono text-xs">
                    {ns}
                  </Badge>
                ))}
              </div>
            </div>
          )}

          {policy.spec.selector.labelSelector && (
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">Label Selector:</p>
              {policy.spec.selector.labelSelector.matchLabels && (
                <div className="flex flex-wrap gap-1">
                  {Object.entries(policy.spec.selector.labelSelector.matchLabels).map(([key, value]) => (
                    <Badge key={key} variant="outline" className="font-mono text-xs">
                      {key}={value}
                    </Badge>
                  ))}
                </div>
              )}
            </div>
          )}

          {policy.spec.selector.matchExpression && (
            <div className="space-y-1">
              <p className="text-sm text-muted-foreground">Match Expression (CEL):</p>
              <pre className="rounded-md bg-muted p-3 text-xs font-mono overflow-x-auto">
                {policy.spec.selector.matchExpression}
              </pre>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Rules */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="flex items-center gap-2 text-base">
            <Code2 className="h-4 w-4" />
            Rules ({policy.spec.rules.length})
          </CardTitle>
        </CardHeader>
        <CardContent>
          <ScrollArea className="max-h-[500px]">
            <div className="space-y-4">
              {policy.spec.rules.map((rule, index) => {
                const ruleStatus = policy.status?.ruleStatuses?.find((rs) => rs.name === rule.name)

                return (
                  <div key={rule.name} className="rounded-lg border p-4 space-y-3">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-2 flex-wrap">
                        {ruleStatus && <RuleStatusIcon error={ruleStatus.lastError} />}
                        <Badge variant="outline" className="font-mono">
                          {rule.name}
                        </Badge>
                        <SeverityBadge severity={rule.severity} showIcon={false} />
                        <Badge variant="secondary">{rule.category}</Badge>
                      </div>
                      {ruleStatus && (
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                          <span>{ruleStatus.activeInsightCount} active</span>
                          {ruleStatus.lastEvaluationTime && (
                            <>
                              <span>·</span>
                              <span title={formatDate(ruleStatus.lastEvaluationTime)}>
                                {formatRelativeTime(ruleStatus.lastEvaluationTime)}
                              </span>
                            </>
                          )}
                        </div>
                      )}
                    </div>

                    <div className="space-y-2">
                      <div className="space-y-1">
                        <p className="text-xs text-muted-foreground">Condition (CEL):</p>
                        <pre className="rounded-md bg-muted p-2 text-xs font-mono overflow-x-auto whitespace-pre-wrap">
                          {rule.condition}
                        </pre>
                      </div>

                      <div className="space-y-1">
                        <p className="text-xs text-muted-foreground">Message Template:</p>
                        <pre className="rounded-md bg-muted p-2 text-xs font-mono overflow-x-auto whitespace-pre-wrap">
                          {rule.messageTemplate}
                        </pre>
                      </div>

                      {rule.descriptionTemplate && (
                        <div className="space-y-1">
                          <p className="text-xs text-muted-foreground">Description Template:</p>
                          <pre className="rounded-md bg-muted p-2 text-xs font-mono overflow-x-auto whitespace-pre-wrap max-h-32">
                            {rule.descriptionTemplate}
                          </pre>
                        </div>
                      )}

                      {rule.persistent ? (
                        <div className="flex items-center gap-2 text-xs">
                          <Infinity className="h-3 w-3 text-indigo-500" />
                          <span className="text-indigo-600 dark:text-indigo-400 font-medium">
                            Persistent (no auto-expiration)
                          </span>
                        </div>
                      ) : rule.ttlSeconds !== undefined && rule.ttlSeconds > 0 ? (
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <Timer className="h-3 w-3" />
                          TTL: {rule.ttlSeconds} seconds
                        </div>
                      ) : null}
                    </div>

                    {ruleStatus?.lastError && (
                      <div className="rounded-md bg-red-100 dark:bg-red-900/30 p-2 text-xs text-red-800 dark:text-red-300">
                        Error: {ruleStatus.lastError}
                      </div>
                    )}

                    {index < policy.spec.rules.length - 1 && <Separator className="mt-4" />}
                  </div>
                )
              })}
            </div>
          </ScrollArea>
        </CardContent>
      </Card>

      {/* Statistics & Timing */}
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">Statistics</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Matching Resources</span>
              <span className="font-semibold">
                {policy.status?.matchingResourceCount ?? 0}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Skipped Resources</span>
              <span className="font-semibold">
                {policy.status?.skippedResourceCount ?? 0}
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-muted-foreground">Total Insights</span>
              <span className="font-semibold">
                {policy.status?.totalInsightCount ?? 0}
              </span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">Timing</CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-center gap-2">
              <Calendar className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm text-muted-foreground">Created:</span>
              <span className="text-sm">
                {policy.metadata.creationTimestamp
                  ? formatDate(policy.metadata.creationTimestamp)
                  : "Unknown"}
              </span>
            </div>
            <div className="flex items-center gap-2">
              <RefreshCw className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm text-muted-foreground">Resync Period:</span>
              <span className="text-sm">{policy.spec.resyncPeriodSeconds ?? 300}s</span>
            </div>
            {policy.status?.lastEvaluationTime && (
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm text-muted-foreground">Last Evaluated:</span>
                <span className="text-sm">
                  {formatRelativeTime(policy.status.lastEvaluationTime)}
                </span>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Tabs: Conditions + Related Insights */}
      <Tabs defaultValue="insights">
        <TabsList>
          <TabsTrigger value="insights">Related Insights</TabsTrigger>
          <TabsTrigger value="conditions">Conditions</TabsTrigger>
        </TabsList>

        <TabsContent value="insights" className="mt-4">
          <PolicyInsightsTab policyName={policy.metadata.name} />
        </TabsContent>

        <TabsContent value="conditions" className="mt-4">
          {policy.status?.conditions && policy.status.conditions.length > 0 ? (
            <Card>
              <CardContent className="pt-4">
                <div className="space-y-3">
                  {policy.status.conditions.map((condition, index) => (
                    <div key={index} className="flex items-start gap-4 rounded-lg border p-3">
                      <div className="mt-0.5">
                        {condition.status === "True" ? (
                          <CheckCircle className="h-4 w-4 text-green-500" />
                        ) : condition.status === "False" ? (
                          <XCircle className="h-4 w-4 text-red-500" />
                        ) : (
                          <MinusCircle className="h-4 w-4 text-gray-400" />
                        )}
                      </div>
                      <div className="flex-1 space-y-1">
                        <div className="flex items-center justify-between">
                          <span className="text-sm font-medium">{condition.type}</span>
                          <span className="text-xs text-muted-foreground">
                            {formatRelativeTime(condition.lastTransitionTime)}
                          </span>
                        </div>
                        <p className="text-sm font-medium text-muted-foreground">{condition.reason}</p>
                        <p className="text-sm text-muted-foreground">{condition.message}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          ) : (
            <p className="text-sm text-muted-foreground py-4">No conditions reported.</p>
          )}
        </TabsContent>
      </Tabs>

      {/* Delete action */}
      {onDelete && (
        <div className="flex justify-end">
          <Button variant="destructive" onClick={() => onDelete(policy)}>
            Delete Policy
          </Button>
        </div>
      )}
    </div>
  )
}
