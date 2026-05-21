"use client"

import { useState } from "react"
import type { Insight, Condition, ActorReference } from "@/types/insights"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { SeverityBadge } from "./severity-badge"
import { InsightStateBadge } from "./insight-state-badge"
import { TargetRefDisplay } from "./target-ref-display"
import { InsightMutedBanner } from "./insight-muted-banner"
import { InsightLifecycleHistory } from "./insight-lifecycle-history"
import { SnoozeDialog } from "./snooze-dialog"
import { AssignDialog } from "./assign-dialog"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import { Progress } from "@/components/ui/progress"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { cn, formatDate, formatRelativeTime, calculateTTLProgress, formatDuration } from "@/lib/utils"
import {
  isInsightPersistent,
  formatSnoozeRemaining,
  getSnoozeExpiration,
  getAvailableActions,
} from "@/lib/insight-utils"
import type { SnoozeOptions } from "@/lib/api"
import {
  ArrowLeft,
  BellOff,
  BellRing,
  Calendar,
  CheckCircle2,
  ChevronRight,
  Clock,
  Copy,
  FileText,
  FolderOpen,
  Infinity,
  Info,
  Link2,
  AlertTriangle,
  AlertCircle,
  Timer,
  Tag,
  Trash2,
  XCircle,
  CheckCircle,
  MinusCircle,
  UserRound,
} from "lucide-react"

interface InsightDetailProps {
  insight: Insight
  onBack?: () => void
  onDelete?: () => void
  onAcknowledge?: (note?: string) => Promise<void>
  onUnacknowledge?: () => Promise<void>
  onSnooze?: (options: SnoozeOptions) => Promise<void>
  onResolve?: (note?: string) => Promise<void>
  onAssign?: (assignee: Pick<ActorReference, "type" | "name">, note?: string) => Promise<void>
  className?: string
}

function CopyButton({ text, label }: { text: string; label: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            onClick={handleCopy}
            aria-label={copied ? "Copied!" : `Copy ${label}`}
          >
            {copied ? (
              <CheckCircle2 className="h-3.5 w-3.5 text-green-500" />
            ) : (
              <Copy className="h-3.5 w-3.5" />
            )}
          </Button>
        </TooltipTrigger>
        <TooltipContent>
          <p>{copied ? "Copied!" : `Copy ${label}`}</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  )
}

function SeverityIcon({ severity }: { severity: string }) {
  switch (severity) {
    case "critical":
      return <AlertCircle className="h-5 w-5" />
    case "warning":
      return <AlertTriangle className="h-5 w-5" />
    default:
      return <Info className="h-5 w-5" />
  }
}

function ConditionStatusIcon({ status }: { status: string }) {
  switch (status) {
    case "True":
      return <CheckCircle className="h-4 w-4 text-green-500" />
    case "False":
      return <XCircle className="h-4 w-4 text-red-500" />
    default:
      return <MinusCircle className="h-4 w-4 text-gray-400" />
  }
}

function ConditionCard({ condition }: { condition: Condition }) {
  return (
    <div className="flex items-start gap-3 rounded-lg border bg-card p-4 transition-colors hover:bg-muted/50">
      <ConditionStatusIcon status={condition.status} />
      <div className="flex-1 min-w-0 space-y-1">
        <div className="flex items-center justify-between gap-2">
          <span className="font-medium text-sm">{condition.type}</span>
          <span className="text-xs text-muted-foreground">
            {formatRelativeTime(condition.lastTransitionTime)}
          </span>
        </div>
        <p className="text-sm font-medium text-muted-foreground">{condition.reason}</p>
        {condition.message && (
          <p className="text-sm text-muted-foreground break-words">{condition.message}</p>
        )}
      </div>
    </div>
  )
}

function TTLIndicator({ insight }: { insight: Insight }) {
  const persistent = isInsightPersistent(insight)
  const ttl = calculateTTLProgress(
    insight.metadata.creationTimestamp,
    insight.spec.ttlSeconds,
    insight.status?.expiresAt
  )

  if (persistent) {
    return (
      <div className="flex items-center gap-2 text-sm">
        <Infinity className="h-4 w-4 text-indigo-500" />
        <span className="font-medium text-indigo-600 dark:text-indigo-400">Persistent</span>
        <span className="text-muted-foreground">- This insight does not auto-expire</span>
      </div>
    )
  }

  if (!insight.spec.ttlSeconds || insight.spec.ttlSeconds <= 0) {
    return (
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Timer className="h-4 w-4" />
        <span>No expiration set</span>
      </div>
    )
  }

  const progressColor = ttl.isExpired
    ? "bg-gray-400"
    : ttl.percentage < 25
      ? "bg-red-500"
      : ttl.percentage < 50
        ? "bg-amber-500"
        : "bg-green-500"

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between text-sm">
        <div className="flex items-center gap-2 text-muted-foreground">
          <Timer className="h-4 w-4" />
          <span>Time remaining</span>
        </div>
        <span className={cn(
          "font-medium",
          ttl.isExpired && "text-muted-foreground",
          !ttl.isExpired && ttl.percentage < 25 && "text-red-600 dark:text-red-400",
          !ttl.isExpired && ttl.percentage >= 25 && ttl.percentage < 50 && "text-amber-600 dark:text-amber-400"
        )}>
          {ttl.remaining}
        </span>
      </div>
      <Progress value={ttl.percentage} className="h-1.5" indicatorClassName={progressColor} />
      <div className="flex items-center justify-between text-xs text-muted-foreground">
        <span>TTL: {formatDuration(insight.spec.ttlSeconds)}</span>
        {insight.status?.expiresAt && (
          <span>Expires: {formatDate(insight.status.expiresAt)}</span>
        )}
      </div>
    </div>
  )
}

function severityBorderColor(severity: string) {
  switch (severity) {
    case "critical":
      return "border-l-red-500"
    case "warning":
      return "border-l-amber-500"
    default:
      return "border-l-blue-500"
  }
}

function severityBgColor(severity: string) {
  switch (severity) {
    case "critical":
      return "bg-red-50 dark:bg-red-950/20"
    case "warning":
      return "bg-amber-50 dark:bg-amber-950/20"
    default:
      return "bg-blue-50 dark:bg-blue-950/20"
  }
}

/**
 * NoteDialog — reusable confirm dialog with an optional note textarea.
 */
function NoteDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel,
  confirmClassName,
  icon,
  isLoading,
  error,
  onConfirm,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: React.ReactNode
  confirmLabel: string
  confirmClassName?: string
  icon?: React.ReactNode
  isLoading: boolean
  error?: string
  onConfirm: (note?: string) => void
}) {
  const [note, setNote] = useState("")

  const handleConfirm = () => {
    onConfirm(note.trim() || undefined)
  }

  const handleOpenChange = (open: boolean) => {
    if (!open) setNote("")
    onOpenChange(open)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[440px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            {icon}
            {title}
          </DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="space-y-3 py-2">
          <div className="space-y-2">
            <Label htmlFor="action-note">
              Note{" "}
              <span className="text-muted-foreground text-xs">(optional)</span>
            </Label>
            <Textarea
              id="action-note"
              placeholder="Add context..."
              value={note}
              onChange={(e) => setNote(e.target.value)}
              disabled={isLoading}
              maxLength={1024}
              rows={3}
            />
          </div>
          {error && (
            <p className="text-sm text-destructive" role="alert">
              {error}
            </p>
          )}
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={isLoading}
            className={confirmClassName}
          >
            {isLoading ? "Loading..." : confirmLabel}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

/**
 * InsightDetail displays the full details of an Insight.
 * Used in detail pages or slide-over panels.
 */
export function InsightDetail({
  insight,
  onBack,
  onDelete,
  onAcknowledge,
  onUnacknowledge,
  onSnooze,
  onResolve,
  onAssign,
  className,
}: InsightDetailProps) {
  const targetRefString = `${insight.spec.targetRef.kind}/${insight.spec.targetRef.name}${
    insight.spec.targetRef.namespace ? ` (${insight.spec.targetRef.namespace})` : ""
  }`

  const snoozed = insight.status?.state === "Snoozed"
  const acknowledged = insight.status?.state === "Acknowledged"
  const muted = insight.status?.muted === true

  const actions = getAvailableActions(insight)

  // Dialog state
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [acknowledgeDialogOpen, setAcknowledgeDialogOpen] = useState(false)
  const [resolveDialogOpen, setResolveDialogOpen] = useState(false)
  const [snoozeDialogOpen, setSnoozeDialogOpen] = useState(false)
  const [assignDialogOpen, setAssignDialogOpen] = useState(false)

  // Action loading/error states
  const [isAcknowledging, setIsAcknowledging] = useState(false)
  const [acknowledgeError, setAcknowledgeError] = useState<string | undefined>()

  const [isUnacknowledging, setIsUnacknowledging] = useState(false)

  const [isResolving, setIsResolving] = useState(false)
  const [resolveError, setResolveError] = useState<string | undefined>()

  const [isSnoozing, setIsSnoozing] = useState(false)
  const [snoozeError, setSnoozeError] = useState<string | undefined>()

  const [isAssigning, setIsAssigning] = useState(false)
  const [assignError, setAssignError] = useState<string | undefined>()

  const [isDeleting, setIsDeleting] = useState(false)

  const handleAcknowledgeConfirm = async (note?: string) => {
    if (!onAcknowledge) return
    setIsAcknowledging(true)
    setAcknowledgeError(undefined)
    try {
      await onAcknowledge(note)
      setAcknowledgeDialogOpen(false)
    } catch (err) {
      setAcknowledgeError(err instanceof Error ? err.message : "Failed to acknowledge insight")
    } finally {
      setIsAcknowledging(false)
    }
  }

  const handleUnacknowledge = async () => {
    if (!onUnacknowledge) return
    setIsUnacknowledging(true)
    try {
      await onUnacknowledge()
    } finally {
      setIsUnacknowledging(false)
    }
  }

  const handleResolveConfirm = async (note?: string) => {
    if (!onResolve) return
    setIsResolving(true)
    setResolveError(undefined)
    try {
      await onResolve(note)
      setResolveDialogOpen(false)
    } catch (err) {
      setResolveError(err instanceof Error ? err.message : "Failed to resolve insight")
    } finally {
      setIsResolving(false)
    }
  }

  const handleSnoozeConfirm = async (options: SnoozeOptions) => {
    if (!onSnooze) return
    setIsSnoozing(true)
    setSnoozeError(undefined)
    try {
      await onSnooze(options)
      setSnoozeDialogOpen(false)
    } catch (err) {
      setSnoozeError(err instanceof Error ? err.message : "Failed to snooze insight")
    } finally {
      setIsSnoozing(false)
    }
  }

  const handleAssignConfirm = async (
    assignee: Pick<ActorReference, "type" | "name">,
    note?: string
  ) => {
    if (!onAssign) return
    setIsAssigning(true)
    setAssignError(undefined)
    try {
      await onAssign(assignee, note)
      setAssignDialogOpen(false)
    } catch (err) {
      setAssignError(err instanceof Error ? err.message : "Failed to assign insight")
    } finally {
      setIsAssigning(false)
    }
  }

  const handleDeleteConfirm = async () => {
    if (!onDelete) return
    setIsDeleting(true)
    try {
      await onDelete()
    } catch {
      setIsDeleting(false)
      setDeleteDialogOpen(false)
    }
  }

  return (
    <TooltipProvider>
      <div className={cn("space-y-6", className)}>
        {/* Breadcrumb */}
        {onBack && (
          <nav className="flex items-center gap-1 text-sm text-muted-foreground" aria-label="Breadcrumb">
            <Button
              variant="link"
              className="h-auto p-0 text-muted-foreground hover:text-foreground"
              onClick={onBack}
            >
              <ArrowLeft className="mr-1 h-3.5 w-3.5" />
              Insights
            </Button>
            <ChevronRight className="h-4 w-4" />
            <span className="text-foreground font-medium truncate max-w-[200px]">
              {insight.metadata.name}
            </span>
          </nav>
        )}

        {/* Hero Header Card */}
        <Card className={cn("border-l-4 overflow-hidden", severityBorderColor(insight.spec.severity))}>
          <div className={cn("px-6 py-4", severityBgColor(insight.spec.severity))}>
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-start gap-3 min-w-0">
                <div className={cn(
                  "flex-shrink-0 p-2 rounded-lg",
                  insight.spec.severity === "critical" && "bg-red-100 text-red-600 dark:bg-red-900/50 dark:text-red-400",
                  insight.spec.severity === "warning" && "bg-amber-100 text-amber-600 dark:bg-amber-900/50 dark:text-amber-400",
                  insight.spec.severity === "info" && "bg-blue-100 text-blue-600 dark:bg-blue-900/50 dark:text-blue-400"
                )}>
                  <SeverityIcon severity={insight.spec.severity} />
                </div>
                <div className="min-w-0">
                  <div className="flex items-center gap-2 mb-1 flex-wrap">
                    <SeverityBadge severity={insight.spec.severity} />
                    <InsightStateBadge insight={insight} />
                  </div>
                  <h1 className="text-xl font-semibold break-words">{insight.spec.message}</h1>
                </div>
              </div>
            </div>
          </div>
          <CardContent className="pt-4 pb-4">
            <div className="flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-muted-foreground">
              <div className="flex items-center gap-1.5">
                <FolderOpen className="h-4 w-4" />
                <span className="font-mono">{insight.metadata.namespace || "default"}</span>
              </div>
              <Separator orientation="vertical" className="h-4" />
              <div className="flex items-center gap-1.5">
                <Tag className="h-4 w-4" />
                <span>{insight.spec.category}</span>
              </div>
              <Separator orientation="vertical" className="h-4" />
              <div className="flex items-center gap-1.5">
                <Calendar className="h-4 w-4" />
                <span>
                  {insight.metadata.creationTimestamp
                    ? formatRelativeTime(insight.metadata.creationTimestamp)
                    : "Unknown"}
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Muted Banner */}
        {muted && (
          <InsightMutedBanner mutedBy={insight.status?.mutedBy} />
        )}

        {/* Snooze Status Banner */}
        {snoozed && (
          <Card className="border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-900/50">
            <CardContent className="py-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <BellOff className="h-5 w-5 text-slate-600 dark:text-slate-400" />
                  <div>
                    <p className="font-medium text-slate-700 dark:text-slate-300">
                      This insight is snoozed
                    </p>
                    <p className="text-sm text-slate-600 dark:text-slate-400">
                      Will reappear in {formatSnoozeRemaining(insight)} (
                      {formatDate(getSnoozeExpiration(insight)?.toISOString() || "")})
                    </p>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Acknowledged Banner */}
        {acknowledged && (
          <Card className="border-purple-300 dark:border-purple-700 bg-purple-50 dark:bg-purple-900/50">
            <CardContent className="py-3">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <CheckCircle2 className="h-5 w-5 text-purple-600 dark:text-purple-400" />
                  <div>
                    <p className="font-medium text-purple-700 dark:text-purple-300">
                      This insight has been acknowledged
                    </p>
                    {insight.status?.acknowledgement?.by && (
                      <p className="text-sm text-purple-600 dark:text-purple-400">
                        By {insight.status.acknowledgement.by.name}
                      </p>
                    )}
                  </div>
                </div>
                {onUnacknowledge && (
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleUnacknowledge}
                    disabled={isUnacknowledging}
                  >
                    <BellRing className="mr-2 h-4 w-4" />
                    {isUnacknowledging ? "Removing..." : "Unacknowledge"}
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Quick Stats Row */}
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {/* TTL Card */}
          <Card>
            <CardContent className="pt-4">
              <TTLIndicator insight={insight} />
            </CardContent>
          </Card>

          {/* Target Resource Card */}
          <Card className="sm:col-span-1 lg:col-span-2">
            <CardContent className="pt-4">
              <div className="flex items-start justify-between gap-2">
                <div className="flex items-center gap-2 text-sm text-muted-foreground mb-2">
                  <Link2 className="h-4 w-4" />
                  <span>Target Resource</span>
                </div>
                <CopyButton text={targetRefString} label="reference" />
              </div>
              <TargetRefDisplay targetRef={insight.spec.targetRef} />
              {insight.status?.targetExists === false && (
                <div className="mt-2 flex items-center gap-2 text-sm text-amber-600 dark:text-amber-400">
                  <AlertTriangle className="h-4 w-4" />
                  <span>Target resource no longer exists</span>
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        {/* Description */}
        {insight.spec.description && (
          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="flex items-center gap-2 text-base">
                <FileText className="h-4 w-4" />
                Description
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="rounded-lg bg-muted/50 p-4">
                <p className="whitespace-pre-wrap text-sm leading-relaxed">
                  {insight.spec.description}
                </p>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Lifecycle History */}
        {insight.status && (
          <InsightLifecycleHistory status={insight.status} />
        )}

        {/* Source Information */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base">Source Information</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="space-y-3">
                <div>
                  <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    Source Type
                  </span>
                  <div className="mt-1">
                    <Badge variant={insight.spec.source.type === "Policy" ? "secondary" : "outline"}>
                      {insight.spec.source.type}
                    </Badge>
                  </div>
                </div>

                {insight.spec.source.type === "Policy" && insight.spec.source.policyRef && (
                  <>
                    <div>
                      <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                        Policy
                      </span>
                      <div className="mt-1 flex items-center gap-2">
                        <span className="font-mono text-sm">
                          {insight.spec.source.policyRef.name}
                        </span>
                        <CopyButton
                          text={insight.spec.source.policyRef.name}
                          label="policy name"
                        />
                      </div>
                    </div>
                    <div>
                      <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                        Rule Name
                      </span>
                      <div className="mt-1">
                        <Badge variant="secondary" className="font-mono">
                          {insight.spec.source.policyRef.ruleName}
                        </Badge>
                      </div>
                    </div>
                  </>
                )}

                {insight.spec.source.type === "Manual" && insight.spec.source.component && (
                  <div>
                    <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                      Component
                    </span>
                    <p className="mt-1 font-mono text-sm">{insight.spec.source.component}</p>
                  </div>
                )}
              </div>

              <div className="space-y-3">
                <div>
                  <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    Created
                  </span>
                  <p className="mt-1 text-sm">
                    {insight.metadata.creationTimestamp
                      ? formatDate(insight.metadata.creationTimestamp)
                      : "Unknown"}
                  </p>
                </div>
                <div>
                  <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
                    Resource Name
                  </span>
                  <div className="mt-1 flex items-center gap-2">
                    <span className="font-mono text-sm truncate">{insight.metadata.name}</span>
                    <CopyButton text={insight.metadata.name} label="name" />
                  </div>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Conditions */}
        {insight.status?.conditions && insight.status.conditions.length > 0 && (
          <Card>
            <CardHeader className="pb-3">
              <div className="flex items-center justify-between">
                <CardTitle className="text-base">Conditions</CardTitle>
                <Badge variant="outline" className="font-normal">
                  {insight.status.conditions.length} condition{insight.status.conditions.length !== 1 ? "s" : ""}
                </Badge>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {insight.status.conditions.map((condition, index) => (
                  <ConditionCard key={index} condition={condition} />
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Actions Footer */}
        <Card>
          <CardContent className="pt-4">
            <div className="flex flex-wrap items-center justify-between gap-4">
              <div className="text-sm text-muted-foreground">
                <Clock className="inline-block h-4 w-4 mr-1" />
                Last updated {insight.metadata.creationTimestamp
                  ? formatRelativeTime(insight.metadata.creationTimestamp)
                  : "unknown"}
              </div>
              <div className="flex items-center gap-2 flex-wrap">
                {onAcknowledge && actions.acknowledge && (
                  <Button variant="outline" onClick={() => setAcknowledgeDialogOpen(true)}>
                    <CheckCircle2 className="mr-2 h-4 w-4" />
                    Acknowledge
                  </Button>
                )}
                {onSnooze && actions.snooze && (
                  <Button variant="outline" onClick={() => setSnoozeDialogOpen(true)}>
                    <BellOff className="mr-2 h-4 w-4" />
                    Snooze
                  </Button>
                )}
                {onResolve && actions.resolve && (
                  <Button variant="outline" onClick={() => setResolveDialogOpen(true)}>
                    <CheckCircle className="mr-2 h-4 w-4" />
                    Mark as Resolved
                  </Button>
                )}
                {onAssign && actions.assign && (
                  <Button variant="outline" onClick={() => setAssignDialogOpen(true)}>
                    <UserRound className="mr-2 h-4 w-4" />
                    Assign
                  </Button>
                )}
                {onDelete && (
                  <Button variant="destructive" onClick={() => setDeleteDialogOpen(true)}>
                    <Trash2 className="mr-2 h-4 w-4" />
                    Delete
                  </Button>
                )}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Acknowledge Dialog */}
      <NoteDialog
        open={acknowledgeDialogOpen}
        onOpenChange={setAcknowledgeDialogOpen}
        title="Acknowledge Insight"
        description={
          <>
            Acknowledge{" "}
            <span className="font-medium text-foreground">{insight.metadata.name}</span>?
            This indicates the team is aware of this issue.
          </>
        }
        confirmLabel="Acknowledge"
        confirmClassName="bg-purple-600 text-white hover:bg-purple-700"
        icon={<CheckCircle2 className="h-5 w-5 text-purple-600" />}
        isLoading={isAcknowledging}
        error={acknowledgeError}
        onConfirm={handleAcknowledgeConfirm}
      />

      {/* Resolve Dialog */}
      <NoteDialog
        open={resolveDialogOpen}
        onOpenChange={setResolveDialogOpen}
        title="Mark as Resolved"
        description={
          <>
            Mark{" "}
            <span className="font-medium text-foreground">{insight.metadata.name}</span> as
            resolved? This indicates the issue has been addressed.
          </>
        }
        confirmLabel="Mark Resolved"
        confirmClassName="bg-green-600 text-white hover:bg-green-700"
        icon={<CheckCircle2 className="h-5 w-5 text-green-600" />}
        isLoading={isResolving}
        error={resolveError}
        onConfirm={handleResolveConfirm}
      />

      {/* Snooze Dialog */}
      <SnoozeDialog
        open={snoozeDialogOpen}
        onOpenChange={setSnoozeDialogOpen}
        onConfirm={handleSnoozeConfirm}
        insightName={insight.metadata.name}
        isLoading={isSnoozing}
        error={snoozeError}
      />

      {/* Assign Dialog */}
      <AssignDialog
        open={assignDialogOpen}
        onOpenChange={setAssignDialogOpen}
        onConfirm={handleAssignConfirm}
        insightName={insight.metadata.name}
        isLoading={isAssigning}
        error={assignError}
      />

      {/* Delete AlertDialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle className="flex items-center gap-2">
              <Trash2 className="h-5 w-5 text-destructive" />
              Delete Insight
            </AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete{" "}
              <span className="font-medium text-foreground">{insight.metadata.name}</span>?
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteConfirm}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </TooltipProvider>
  )
}
