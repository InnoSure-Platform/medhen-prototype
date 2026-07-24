import * as React from "react";
import { cn } from "@/lib/utils";

export interface PageHeaderProps {
  eyebrow?: React.ReactNode;
  title: string;
  subtitle?: string;
  actions?: React.ReactNode;
  className?: string;
}

export function PageHeader({ eyebrow, title, subtitle, actions, className }: PageHeaderProps) {
  return (
    <div className={cn("flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between", className)}>
      <div className="space-y-2">
        {eyebrow}
        <h1 className="font-display text-2xl font-bold tracking-tight text-fg sm:text-3xl">{title}</h1>
        {subtitle && <p className="max-w-2xl text-fg-muted">{subtitle}</p>}
      </div>
      {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
    </div>
  );
}

export function Eyebrow({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <span
      className={cn(
        "inline-flex items-center gap-2 rounded-full border border-brand-border bg-brand-subtle px-3 py-1.5 text-[11px] font-bold uppercase tracking-widest text-brand-fg [&_svg]:size-3.5",
        className,
      )}
    >
      {children}
    </span>
  );
}
