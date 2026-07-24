import * as React from "react";
import { cn } from "@/lib/utils";

type Variant = "neutral" | "brand" | "success" | "warning" | "danger";

export function Badge({
  variant = "neutral",
  dot = false,
  className,
  children,
}: {
  variant?: Variant;
  dot?: boolean;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <span className={cn("badge", `badge-${variant}`, className)}>
      {dot && <span className="badge-dot" />}
      {children}
    </span>
  );
}

/** Maps a domain status or event topic to a semantic badge variant + label. */
export function StatusBadge({ value, className }: { value: string; className?: string }) {
  const v = value.toUpperCase();
  let variant: Variant = "neutral";
  if (["ISSUED", "PAID", "SETTLED", "ACTIVE", "AUTO_ACCEPT", "SENT"].some((s) => v.includes(s))) variant = "success";
  else if (["QUOTED", "OPEN", "FILED", "BOUND", "QUEUED"].some((s) => v.includes(s))) variant = "brand";
  else if (["REFER", "PARTIALLY", "PENDING"].some((s) => v.includes(s))) variant = "warning";
  else if (["DECLINE", "CANCELLED", "FAILED", "REJECTED"].some((s) => v.includes(s))) variant = "danger";
  return (
    <Badge variant={variant} dot className={cn("uppercase tracking-wide", className)}>
      {value.replace(/_/g, " ")}
    </Badge>
  );
}
