"use client"

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

interface NamespaceSelectorProps {
  namespaces: string[]
  selected?: string
  onSelect: (namespace: string | undefined) => void
  showAllOption?: boolean
  className?: string
}

/**
 * NamespaceSelector allows selecting a Kubernetes namespace to filter resources.
 * Can be embedded in any portal that needs namespace filtering.
 */
export function NamespaceSelector({
  namespaces,
  selected,
  onSelect,
  showAllOption = true,
  className,
}: NamespaceSelectorProps) {
  return (
    <Select
      value={selected || "all"}
      onValueChange={(value) => onSelect(value === "all" ? undefined : value)}
    >
      <SelectTrigger className={className}>
        <SelectValue placeholder="Select namespace" />
      </SelectTrigger>
      <SelectContent>
        {showAllOption && (
          <SelectItem value="all">All Namespaces</SelectItem>
        )}
        {namespaces.map((ns) => (
          <SelectItem key={ns} value={ns}>
            {ns}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}
