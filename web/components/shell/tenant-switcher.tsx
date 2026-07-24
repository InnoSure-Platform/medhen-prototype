"use client";

import * as React from "react";
import { Building2, Check, ChevronsUpDown } from "lucide-react";
import { cn } from "@/lib/utils";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

// Tenants the UI offers. The proxy independently re-validates the selection
// against a server allowlist (MEDHEN_ALLOWED_TENANTS) before forwarding.
const TENANTS = (process.env.NEXT_PUBLIC_MEDHEN_TENANTS ?? "eic")
  .split(",")
  .map((s) => s.trim())
  .filter(Boolean);

function readCookie(name: string): string | undefined {
  if (typeof document === "undefined") return undefined;
  return document.cookie
    .split("; ")
    .find((c) => c.startsWith(name + "="))
    ?.split("=")[1];
}

export function TenantSwitcher({ className }: { className?: string }) {
  const [current, setCurrent] = React.useState(TENANTS[0]);

  React.useEffect(() => {
    setCurrent(readCookie("medhen_tenant") ?? TENANTS[0]);
  }, []);

  const select = (tenant: string) => {
    document.cookie = `medhen_tenant=${tenant}; path=/; max-age=${60 * 60 * 24 * 30}; samesite=lax`;
    // Reload so all server-side proxy calls pick up the new tenant.
    window.location.reload();
  };

  const chip = (
    <div className={cn("flex items-center gap-2 rounded-xl border border-border bg-subtle px-3 py-2 text-xs", className)}>
      <Building2 className="size-3.5 text-fg-muted" />
      <span className="font-mono text-fg">tenant · {current}</span>
    </div>
  );

  // Single tenant → static chip, no switcher.
  if (TENANTS.length < 2) return chip;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="w-full outline-none focus-visible:ring-2 focus-visible:ring-ring rounded-xl">
        <div className="flex items-center justify-between gap-2 rounded-xl border border-border bg-subtle px-3 py-2 text-xs">
          <span className="flex items-center gap-2">
            <Building2 className="size-3.5 text-fg-muted" />
            <span className="font-mono text-fg">tenant · {current}</span>
          </span>
          <ChevronsUpDown className="size-3.5 text-fg-muted" />
        </div>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-[var(--radix-dropdown-menu-trigger-width)]">
        <DropdownMenuLabel>Tenant</DropdownMenuLabel>
        {TENANTS.map((t) => (
          <DropdownMenuItem key={t} onClick={() => select(t)} className="justify-between">
            <span className="font-mono">{t}</span>
            {t === current && <Check className="size-4 text-brand-fg" />}
          </DropdownMenuItem>
        ))}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
