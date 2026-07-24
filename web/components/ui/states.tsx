import * as React from "react";
import { Inbox, TriangleAlert } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "./button";

interface StateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: { label: string; onClick?: () => void; href?: string };
  className?: string;
}

/** Empty state — no data yet. */
export function EmptyState({ icon, title, description, action, className }: StateProps) {
  return (
    <div className={cn("flex flex-col items-center gap-3 px-6 py-16 text-center", className)}>
      <span className="grid size-12 place-items-center rounded-2xl bg-subtle text-fg-subtle">
        {icon ?? <Inbox className="size-6" />}
      </span>
      <div className="space-y-1">
        <p className="font-semibold text-fg">{title}</p>
        {description && <p className="mx-auto max-w-sm text-sm text-fg-muted">{description}</p>}
      </div>
      {action &&
        (action.href ? (
          <Button asChild size="sm" className="mt-1">
            <a href={action.href}>{action.label}</a>
          </Button>
        ) : (
          <Button size="sm" className="mt-1" onClick={action.onClick}>
            {action.label}
          </Button>
        ))}
    </div>
  );
}

/** Error state — a load failed. */
export function ErrorState({ title, description, action, className }: Omit<StateProps, "icon">) {
  return (
    <div className={cn("flex flex-col items-center gap-3 px-6 py-16 text-center", className)}>
      <span className="grid size-12 place-items-center rounded-2xl bg-danger-subtle text-danger-fg">
        <TriangleAlert className="size-6" />
      </span>
      <div className="space-y-1">
        <p className="font-semibold text-fg">{title}</p>
        {description && <p className="mx-auto max-w-sm text-sm text-fg-muted">{description}</p>}
      </div>
      {action && (
        <Button variant="secondary" size="sm" className="mt-1" onClick={action.onClick}>
          {action.label}
        </Button>
      )}
    </div>
  );
}
