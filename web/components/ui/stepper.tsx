import * as React from "react";
import { Check } from "lucide-react";
import { cn } from "@/lib/utils";

export interface Step {
  id: string;
  label: string;
}

/** Horizontal progress stepper for multi-step flows (e.g. the quote wizard). */
export function Stepper({ steps, current, className }: { steps: Step[]; current: number; className?: string }) {
  return (
    <nav aria-label="Progress" className={className}>
      <ol className="flex items-center">
        {steps.map((step, i) => {
          const done = i < current;
          const active = i === current;
          const last = i === steps.length - 1;
          return (
            <li key={step.id} className={cn("relative flex items-center", !last && "flex-1")}>
              <div className="flex flex-col items-center gap-2">
                <span
                  aria-current={active ? "step" : undefined}
                  className={cn(
                    "z-10 grid size-9 place-items-center rounded-full border-2 text-sm font-bold transition-colors",
                    done && "border-brand bg-brand text-fg-onbrand",
                    active && "border-brand bg-surface text-brand-fg ring-4 ring-brand-subtle",
                    !done && !active && "border-border-strong bg-surface text-fg-subtle",
                  )}
                >
                  {done ? <Check className="size-4" strokeWidth={3} /> : i + 1}
                </span>
                <span className={cn("absolute top-11 whitespace-nowrap text-xs font-semibold", active ? "text-brand-fg" : "text-fg-muted")}>
                  {step.label}
                </span>
              </div>
              {!last && <span className={cn("mx-2 h-0.5 flex-1 rounded", done ? "bg-brand" : "bg-border")} aria-hidden />}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}
