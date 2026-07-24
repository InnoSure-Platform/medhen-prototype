import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { ArrowDownRight, ArrowUpRight } from "lucide-react";
import { cn } from "@/lib/utils";
import { Sparkline } from "./charts";

const iconTone = cva("grid size-10 shrink-0 place-items-center rounded-xl border [&_svg]:size-5", {
  variants: {
    tone: {
      brand: "border-brand-border bg-brand-subtle text-brand-fg",
      success: "border-success/20 bg-success-subtle text-success-fg",
      warning: "border-warning/20 bg-warning-subtle text-warning-fg",
      danger: "border-danger/20 bg-danger-subtle text-danger-fg",
      info: "border-info/20 bg-info-subtle text-info-fg",
    },
  },
  defaultVariants: { tone: "brand" },
});

const sparkColor: Record<string, string> = {
  brand: "var(--chart-1)",
  success: "var(--success-default)",
  warning: "var(--warning-default)",
  danger: "var(--danger-default)",
  info: "var(--info-default)",
};

export interface StatCardProps extends VariantProps<typeof iconTone> {
  label: string;
  value: string;
  icon?: React.ReactNode;
  hint?: string;
  delta?: { value: string; direction: "up" | "down"; good?: boolean };
  trend?: number[];
  className?: string;
}

export function StatCard({ label, value, icon, hint, delta, trend, tone = "brand", className }: StatCardProps) {
  const up = delta?.direction === "up";
  const positive = delta?.good ?? up;
  return (
    <div className={cn("relative overflow-hidden rounded-2xl border border-border bg-surface p-5 shadow-[var(--shadow-card)] transition-all duration-300 hover:-translate-y-1 hover:shadow-[var(--shadow-lift)]", className)}>
      <div className="flex items-start justify-between gap-3">
        <span className="text-xs font-semibold uppercase tracking-wider text-fg-muted">{label}</span>
        {icon && <span className={iconTone({ tone })}>{icon}</span>}
      </div>
      <div className="mt-3 flex items-end justify-between gap-3">
        <div>
          <p className="font-display text-2xl font-bold tracking-tight text-fg sm:text-3xl">{value}</p>
          <div className="mt-1 flex items-center gap-2 text-xs">
            {delta && (
              <span className={cn("inline-flex items-center gap-0.5 font-semibold", positive ? "text-success-fg" : "text-danger-fg")}>
                {up ? <ArrowUpRight className="size-3.5" /> : <ArrowDownRight className="size-3.5" />}
                {delta.value}
              </span>
            )}
            {hint && <span className="text-fg-subtle">{hint}</span>}
          </div>
        </div>
        {trend && trend.length > 1 && (
          <div className="h-10 w-24 shrink-0">
            <Sparkline data={trend} color={sparkColor[tone ?? "brand"]} />
          </div>
        )}
      </div>
    </div>
  );
}
