"use client"

import { useState } from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
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
import { UserRound } from "lucide-react"
import type { ActorReference } from "@/types/insights"

interface AssignDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (assignee: Pick<ActorReference, "type" | "name">, note?: string) => void
  insightName: string
  isLoading?: boolean
  error?: string
}

export function AssignDialog({
  open,
  onOpenChange,
  onConfirm,
  insightName,
  isLoading = false,
  error,
}: AssignDialogProps) {
  const [assigneeName, setAssigneeName] = useState("")
  const [assigneeType, setAssigneeType] = useState<"user" | "serviceaccount">("user")
  const [note, setNote] = useState("")

  const isValid = assigneeName.trim().length > 0

  const handleConfirm = () => {
    if (!isValid) return
    onConfirm(
      { type: assigneeType, name: assigneeName.trim() },
      note.trim() || undefined
    )
  }

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setAssigneeName("")
      setAssigneeType("user")
      setNote("")
    }
    onOpenChange(open)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[440px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <UserRound className="h-5 w-5 text-muted-foreground" />
            Assign Insight
          </DialogTitle>
          <DialogDescription>
            Assign{" "}
            <span className="font-medium text-foreground">{insightName}</span> to
            a user or service account.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div className="space-y-2">
            <Label htmlFor="assignee-type">Assignee Type</Label>
            <Select
              value={assigneeType}
              onValueChange={(value) => setAssigneeType(value as "user" | "serviceaccount")}
              disabled={isLoading}
            >
              <SelectTrigger id="assignee-type">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="user">User</SelectItem>
                <SelectItem value="serviceaccount">Service Account</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="assignee-name">
              Assignee Name <span className="text-destructive">*</span>
            </Label>
            <Input
              id="assignee-name"
              placeholder={assigneeType === "user" ? "alice" : "my-service-account"}
              value={assigneeName}
              onChange={(e) => setAssigneeName(e.target.value)}
              disabled={isLoading}
              aria-required="true"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="assign-note">
              Note{" "}
              <span className="text-muted-foreground text-xs">(optional)</span>
            </Label>
            <Textarea
              id="assign-note"
              placeholder="Add context about why this is being assigned..."
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
          <Button onClick={handleConfirm} disabled={isLoading || !isValid}>
            {isLoading ? "Assigning..." : "Assign"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
