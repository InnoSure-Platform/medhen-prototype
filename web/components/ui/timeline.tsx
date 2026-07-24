import * as React from "react";
import { Check } from "lucide-react";
import { cn } from "@/lib/utils";

export interface TimelineItem {
  title: string;
  description?: string;
  timestamp?: string;
  icon?: React.ReactNode;
  state?: "done" | "current" | "upcoming";
}

const dotState = {
  done: "border-success bg-success text-white",
  current: "border-brand bg-brand text-fg-onbrand ring-4 ring-brand-subtle",
  upcoming: "border-border-strong bg-surface text-fg-subtle",
};

export function Timeline({ items, className }: { items: TimelineItem[]; className?: string }) {
  return (
    <ol className={cn("relative flex flex-col", className)}>
      {items.map((item, i) => {
        const state = item.state ?? "done";
        const last = i === items.length - 1;
        return (
          <li key={i} className="relative flex gap-4 pb-6 last:pb-0">
            {!last && <span className="absolute left-[15px] top-8 h-[calc(100%-1rem)] w-px bg-border" aria-hidden />}
            <span className={cn("z-10 grid size-8 shrink-0 place-items-center rounded-full border-2 [&_svg]:size-4", dotState[state])}>
              {item.icon ?? (state === "done" ? <Check className="size-4" strokeWidth={3} /> : <span className="size-2 rounded-full bg-current" />)}
            </span>
            <div className="pt-1">
              <p className="font-semibold text-fg">{item.title}</p>
              {item.description && <p className="mt-0.5 text-sm text-fg-muted">{item.description}</p>}
              {item.timestamp && <p className="mt-1 font-mono text-xs text-fg-subtle">{item.timestamp}</p>}
            </div>
          </li>
        );
      })}
    </ol>
  );
}
