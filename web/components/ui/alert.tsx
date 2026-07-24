import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";
import { AlertTriangle, CheckCircle2, Info, XCircle } from "lucide-react";
import { cn } from "@/lib/utils";

const alertVariants = cva("flex gap-3 rounded-xl border p-4 text-sm", {
  variants: {
    tone: {
      info: "border-info/25 bg-info-subtle text-info-fg",
      success: "border-success/25 bg-success-subtle text-success-fg",
      warning: "border-warning/25 bg-warning-subtle text-warning-fg",
      danger: "border-danger/25 bg-danger-subtle text-danger-fg",
    },
  },
  defaultVariants: { tone: "info" },
});

const ICONS = { info: Info, success: CheckCircle2, warning: AlertTriangle, danger: XCircle };

export interface AlertProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof alertVariants> {
  title?: string;
}

export function Alert({ className, tone = "info", title, children, ...props }: AlertProps) {
  const Icon = ICONS[tone ?? "info"];
  return (
    <div role="alert" className={cn(alertVariants({ tone }), className)} {...props}>
      <Icon className="mt-0.5 size-4 shrink-0" aria-hidden />
      <div className="flex flex-col gap-0.5">
        {title && <p className="font-semibold text-fg">{title}</p>}
        {children && <div className="text-fg-muted [&_a]:font-medium [&_a]:underline">{children}</div>}
      </div>
    </div>
  );
}
