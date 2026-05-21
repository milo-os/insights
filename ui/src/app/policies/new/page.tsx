"use client"

import { useState } from "react"
import { useCreatePolicy } from "@/hooks/use-insights"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Textarea } from "@/components/ui/textarea"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { Separator } from "@/components/ui/separator"
import { useRouter } from "next/navigation"
import { ArrowLeft, Plus, Trash2 } from "lucide-react"
import type { InsightSeverity, InsightRule } from "@/types/insights"

interface RuleFormData {
  name: string
  condition: string
  severity: InsightSeverity
  category: string
  messageTemplate: string
  descriptionTemplate: string
  ttlSeconds: number
  persistent: boolean
}

const defaultRule: RuleFormData = {
  name: "",
  condition: "",
  severity: "info",
  category: "",
  messageTemplate: "",
  descriptionTemplate: "",
  ttlSeconds: 0,
  persistent: false,
}

export default function NewPolicyPage() {
  const router = useRouter()
  const createPolicy = useCreatePolicy()

  const [name, setName] = useState("")
  const [selectorApiVersion, setSelectorApiVersion] = useState("apps/v1")
  const [selectorKind, setSelectorKind] = useState("Deployment")
  const [selectorNamespaces, setSelectorNamespaces] = useState("")
  const [matchExpression, setMatchExpression] = useState("")
  const [resyncPeriodSeconds, setResyncPeriodSeconds] = useState(300)
  const [suspended, setSuspended] = useState(false)
  const [rules, setRules] = useState<RuleFormData[]>([{ ...defaultRule }])

  const addRule = () => {
    setRules([...rules, { ...defaultRule }])
  }

  const removeRule = (index: number) => {
    if (rules.length > 1) {
      setRules(rules.filter((_, i) => i !== index))
    }
  }

  const updateRule = (index: number, field: keyof RuleFormData, value: string | number | boolean) => {
    const newRules = [...rules]
    newRules[index] = { ...newRules[index], [field]: value }
    setRules(newRules)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const parsedNamespaces = selectorNamespaces
      .split(",")
      .map((ns) => ns.trim())
      .filter(Boolean)

    const policyRules: InsightRule[] = rules.map((rule) => ({
      name: rule.name,
      condition: rule.condition,
      severity: rule.severity,
      category: rule.category,
      messageTemplate: rule.messageTemplate,
      descriptionTemplate: rule.descriptionTemplate || undefined,
      ttlSeconds: rule.persistent ? undefined : (rule.ttlSeconds || undefined),
      persistent: rule.persistent || undefined,
    }))

    try {
      await createPolicy.mutateAsync({
        metadata: {
          name,
        },
        spec: {
          selector: {
            apiVersion: selectorApiVersion,
            kind: selectorKind,
            namespaces: parsedNamespaces.length > 0 ? parsedNamespaces : undefined,
            matchExpression: matchExpression || undefined,
          },
          rules: policyRules,
          resyncPeriodSeconds,
          suspended,
        },
      })

      router.push("/policies")
    } catch (error) {
      console.error("Failed to create policy:", error)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="flex items-center gap-4">
        <Button type="button" variant="ghost" size="sm" onClick={() => router.push("/policies")}>
          <ArrowLeft className="mr-2 h-4 w-4" />
          Back
        </Button>
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Create Policy</h1>
          <p className="text-muted-foreground">
            Define a new InsightPolicy to automatically generate insights.
          </p>
        </div>
      </div>

      {/* Basic Info */}
      <Card>
        <CardHeader>
          <CardTitle>Basic Information</CardTitle>
          <CardDescription>InsightPolicies are cluster-scoped — only a name is required.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Policy Name</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="deployment-best-practices"
              required
              pattern="^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
            />
          </div>
          <div className="flex items-center space-x-2">
            <Switch
              id="suspended"
              checked={suspended}
              onCheckedChange={setSuspended}
            />
            <Label htmlFor="suspended">Start suspended (policy will not evaluate resources)</Label>
          </div>
        </CardContent>
      </Card>

      {/* Selector */}
      <Card>
        <CardHeader>
          <CardTitle>Resource Selector</CardTitle>
          <CardDescription>Define which resources this policy applies to</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="space-y-2">
              <Label htmlFor="apiVersion">API Version</Label>
              <Input
                id="apiVersion"
                value={selectorApiVersion}
                onChange={(e) => setSelectorApiVersion(e.target.value)}
                placeholder="apps/v1"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="kind">Kind</Label>
              <Input
                id="kind"
                value={selectorKind}
                onChange={(e) => setSelectorKind(e.target.value)}
                placeholder="Deployment"
                required
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="namespaces">Namespaces (comma-separated, leave empty for all)</Label>
            <Input
              id="namespaces"
              value={selectorNamespaces}
              onChange={(e) => setSelectorNamespaces(e.target.value)}
              placeholder="default, production, staging"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="matchExpression">Match Expression (CEL, optional)</Label>
            <Textarea
              id="matchExpression"
              value={matchExpression}
              onChange={(e) => setMatchExpression(e.target.value)}
              placeholder='object.metadata.labels.exists(k, k == "app")'
              className="font-mono text-sm"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="resyncPeriod">Resync Period (seconds)</Label>
            <Input
              id="resyncPeriod"
              type="number"
              min={60}
              value={resyncPeriodSeconds}
              onChange={(e) => setResyncPeriodSeconds(parseInt(e.target.value, 10))}
            />
          </div>
        </CardContent>
      </Card>

      {/* Rules */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>Rules</CardTitle>
            <CardDescription>Define rules that generate insights when matched</CardDescription>
          </div>
          <Button type="button" variant="outline" size="sm" onClick={addRule}>
            <Plus className="mr-2 h-4 w-4" />
            Add Rule
          </Button>
        </CardHeader>
        <CardContent className="space-y-6">
          {rules.map((rule, index) => (
            <div key={index} className="space-y-4 rounded-lg border p-4">
              <div className="flex items-center justify-between">
                <h4 className="font-medium">Rule {index + 1}</h4>
                {rules.length > 1 && (
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => removeRule(index)}
                    className="text-destructive hover:text-destructive"
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                )}
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label>Rule Name</Label>
                  <Input
                    value={rule.name}
                    onChange={(e) => updateRule(index, "name", e.target.value)}
                    placeholder="no-resource-limits"
                    required
                    pattern="^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Category</Label>
                  <Input
                    value={rule.category}
                    onChange={(e) => updateRule(index, "category", e.target.value)}
                    placeholder="configuration"
                    required
                  />
                </div>
              </div>

              <div className="grid gap-4 md:grid-cols-2">
                <div className="space-y-2">
                  <Label>Severity</Label>
                  <Select
                    value={rule.severity}
                    onValueChange={(value) => updateRule(index, "severity", value)}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="info">Info</SelectItem>
                      <SelectItem value="warning">Warning</SelectItem>
                      <SelectItem value="critical">Critical</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label className={rule.persistent ? "text-muted-foreground" : ""}>
                      TTL (seconds, 0 = never expires)
                    </Label>
                    <div className="flex items-center space-x-2">
                      <Switch
                        id={`persistent-${index}`}
                        checked={rule.persistent}
                        onCheckedChange={(checked) => updateRule(index, "persistent", checked)}
                      />
                      <Label htmlFor={`persistent-${index}`} className="text-sm font-normal">
                        Persistent
                      </Label>
                    </div>
                  </div>
                  <Input
                    type="number"
                    min={0}
                    value={rule.ttlSeconds}
                    onChange={(e) => updateRule(index, "ttlSeconds", parseInt(e.target.value, 10))}
                    disabled={rule.persistent}
                    className={rule.persistent ? "opacity-50" : ""}
                  />
                  {rule.persistent && (
                    <p className="text-xs text-muted-foreground">
                      Persistent insights do not auto-expire and require manual action.
                    </p>
                  )}
                </div>
              </div>

              <div className="space-y-2">
                <Label>Condition (CEL expression)</Label>
                <Textarea
                  value={rule.condition}
                  onChange={(e) => updateRule(index, "condition", e.target.value)}
                  placeholder="!has(object.spec.template.spec.containers[0].resources.limits)"
                  className="font-mono text-sm"
                  required
                />
              </div>

              <div className="space-y-2">
                <Label>Message Template</Label>
                <Textarea
                  value={rule.messageTemplate}
                  onChange={(e) => updateRule(index, "messageTemplate", e.target.value)}
                  placeholder="Deployment {{ object.metadata.name }} has no resource limits"
                  className="font-mono text-sm"
                  required
                />
              </div>

              <div className="space-y-2">
                <Label>Description Template (optional)</Label>
                <Textarea
                  value={rule.descriptionTemplate}
                  onChange={(e) => updateRule(index, "descriptionTemplate", e.target.value)}
                  placeholder="The deployment {{ object.metadata.name }} does not have resource limits configured..."
                  className="font-mono text-sm"
                  rows={4}
                />
              </div>

              {index < rules.length - 1 && <Separator className="mt-4" />}
            </div>
          ))}
        </CardContent>
      </Card>

      {/* Submit */}
      <div className="flex justify-end gap-4">
        <Button type="button" variant="outline" onClick={() => router.push("/policies")}>
          Cancel
        </Button>
        <Button type="submit" disabled={createPolicy.isPending}>
          {createPolicy.isPending ? "Creating..." : "Create Policy"}
        </Button>
      </div>
    </form>
  )
}
