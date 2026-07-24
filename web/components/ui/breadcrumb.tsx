import * as React from "react";
import { ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

export interface Crumb {
  label: string;
  href?: string;
}

export function Breadcrumb({
  items,
  LinkComponent = "a",
  className,
}: {
  items: Crumb[];
  LinkComponent?: React.ElementType;
  className?: string;
}) {
  return (
    <nav aria-label="Breadcrumb" className={cn("flex items-center gap-1.5 text-sm", className)}>
      {items.map((item, i) => {
        const last = i === items.length - 1;
        return (
          <React.Fragment key={i}>
            {item.href && !last ? (
              <LinkComponent href={item.href} className="text-fg-muted transition-colors hover:text-fg">
                {item.label}
              </LinkComponent>
            ) : (
              <span className={cn(last ? "font-medium text-fg" : "text-fg-muted")} aria-current={last ? "page" : undefined}>
                {item.label}
              </span>
            )}
            {!last && <ChevronRight className="size-4 text-fg-subtle" aria-hidden />}
          </React.Fragment>
        );
      })}
    </nav>
  );
}
