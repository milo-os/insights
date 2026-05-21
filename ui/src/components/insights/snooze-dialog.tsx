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
import { Label } from "@/components/ui/label"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { Input } from "@/components/ui/input"
import { Clock } from "lucide-react"
import { SNOOZE_DURATIONS } from "@/lib/insight-utils"
import type { SnoozeOptions } from "@/lib/api"

const CUSTOM_VALUE = "custom"

interface SnoozeDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (options: SnoozeOptions) => void
  insightName: string
  isLoading?: boolean
  error?: string
}

export function SnoozeDialog({
  open,
  onOpenChange,
  onConfirm,
  insightName,
  isLoading = false,
  error,
}: SnoozeDialogProps) {
  const [selected, setSelected] = useState<string>("4h")
  const [customDatetime, setCustomDatetime] = useState("")

  const isCustom = selected === CUSTOM_VALUE

  // For the custom option, compute a default min value (now)
  const minDatetime = new Date().toISOString().slice(0, 16)

  const isValid = isCustom ? customDatetime.length > 0 : true

  const handleConfirm = () => {
    if (!isValid) return

    if (isCustom) {
      const until = new Date(customDatetime).toISOString()
      onConfirm({ until })
    } else {
      onConfirm({ duration: selected })
    }
  }

  const handleOpenChange = (open: boolean) => {
    if (!open) {
      setSelected("4h")
      setCustomDatetime("")
    }
    onOpenChange(open)
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[420px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Clock className="h-5 w-5 text-muted-foreground" />
            Snooze Insight
          </DialogTitle>
          <DialogDescription>
            Snooze{" "}
            <span className="font-medium text-foreground">{insightName}</span>{" "}
            to temporarily suppress it.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-2">
          <div>
            <Label className="text-sm font-medium mb-3 block">Snooze duration</Label>
            <RadioGroup
              value={selected}
              onValueChange={setSelected}
              className="grid grid-cols-2 gap-3"
              aria-label="Snooze duration"
            >
              {SNOOZE_DURATIONS.map((duration) => (
                <div key={duration.value} className="flex items-center space-x-2">
                  <RadioGroupItem
                    value={duration.value}
                    id={`duration-${duration.value}`}
                    className="peer sr-only"
                    disabled={isLoading}
                  />
                  <Label
                    htmlFor={`duration-${duration.value}`}
                    className="flex-1 cursor-pointer rounded-md border border-input bg-background px-4 py-3 text-center text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary peer-data-[state=checked]:bg-primary/10 peer-data-[state=checked]:text-primary peer-disabled:cursor-not-allowed peer-disabled:opacity-50"
                  >
                    {duration.label}
                  </Label>
                </div>
              ))}

              {/* Custom datetime option */}
              <div className="flex items-center space-x-2 col-span-2">
                <RadioGroupItem
                  value={CUSTOM_VALUE}
                  id="duration-custom"
                  className="peer sr-only"
                  disabled={isLoading}
                />
                <Label
                  htmlFor="duration-custom"
                  className="flex-1 cursor-pointer rounded-md border border-input bg-background px-4 py-3 text-center text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground peer-data-[state=checked]:border-primary peer-data-[state=checked]:bg-primary/10 peer-data-[state=checked]:text-primary peer-disabled:cursor-not-allowed peer-disabled:opacity-50"
                >
                  Custom time
                </Label>
              </div>
            </RadioGroup>
          </div>

          {isCustom && (
            <div className="space-y-2">
              <Label htmlFor="custom-datetime">
                Snooze until <span className="text-destructive">*</span>
              </Label>
              <Input
                id="custom-datetime"
                type="datetime-local"
                value={customDatetime}
                min={minDatetime}
                onChange={(e) => setCustomDatetime(e.target.value)}
                disabled={isLoading}
                aria-required="true"
              />
            </div>
          )}

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
            {isLoading ? "Snoozing..." : "Snooze"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
