import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-xs font-semibold whitespace-nowrap",
  {
    variants: {
      tone: {
        neutral: "border-border bg-subtle text-fg-muted",
        brand: "border-brand-border bg-brand-subtle text-brand-fg",
        success: "border-success/25 bg-success-subtle text-success-fg",
        warning: "border-warning/25 bg-warning-subtle text-warning-fg",
        danger: "border-danger/25 bg-danger-subtle text-danger-fg",
        info: "border-info/25 bg-info-subtle text-info-fg",
        outline: "border-border bg-transparent text-fg",
      },
    },
    defaultVariants: { tone: "neutral" },
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {
  dot?: boolean;
}

export function Badge({ className, tone, dot = false, children, ...props }: BadgeProps) {
  return (
    <span className={cn(badgeVariants({ tone }), className)} {...props}>
      {dot && <span className="size-1.5 rounded-full bg-current" aria-hidden />}
      {children}
    </span>
  );
}

// Maps monolith domain statuses (policies, quotes, claims, invoices, events) to
// a semantic tone so status pills read consistently across the app.
const STATUS_TONE: Record<string, VariantProps<typeof badgeVariants>["tone"]> = {
  QUOTED: "info",
  BOUND: "brand",
  ISSUED: "success",
  ACTIVE: "success",
  PAID: "success",
  SETTLED: "success",
  CANCELLED: "neutral",
  EXPIRED: "neutral",
  OPEN: "warning",
  PARTIALLY_PAID: "warning",
  FILED: "warning",
  REFERRED: "warning",
  PENDING: "warning",
  REJECTED: "danger",
  DECLINED: "danger",
};

export function StatusBadge({ value, className }: { value: string; className?: string }) {
  const key = value?.toUpperCase?.() ?? "";
  const tone = STATUS_TONE[key] ?? "neutral";
  return (
    <Badge tone={tone} dot className={cn("uppercase tracking-wide", className)}>
      {value?.replace(/[._]/g, " ")}
    </Badge>
  );
}
