import * as React from "react";
import { cn } from "@/lib/utils";

/** A KPI tile with an icon, label, value and an optional trend/footnote. */
export function StatCard({
  label,
  value,
  icon,
  hint,
  tone = "brand",
  className,
}: {
  label: string;
  value: React.ReactNode;
  icon?: React.ReactNode;
  hint?: React.ReactNode;
  tone?: "brand" | "success" | "warning" | "danger";
  className?: string;
}) {
  const toneMap = {
    brand: "bg-brand-50 text-brand-600 border-brand-100",
    success: "bg-success-50 text-success-600 border-emerald-100",
    warning: "bg-warning-50 text-warning-700 border-amber-100",
    danger: "bg-danger-50 text-danger-600 border-rose-100",
  } as const;

  return (
    <div className={cn("stat-card", className)}>
      {/* decorative wash */}
      <div className="pointer-events-none absolute -right-6 -top-6 h-24 w-24 rounded-full bg-slate-50" />
      <div className="relative flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="stat-label">{label}</div>
          <div className="stat-value truncate">{value}</div>
          {hint && <div className="mt-1 text-xs font-medium text-slate-400">{hint}</div>}
        </div>
        {icon && <div className={cn("stat-icon shrink-0", toneMap[tone])}>{icon}</div>}
      </div>
    </div>
  );
}
