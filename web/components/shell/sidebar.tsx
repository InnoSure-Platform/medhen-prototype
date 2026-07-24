"use client";

import { useTranslations } from "next-intl";
import { Link, usePathname } from "@/lib/i18n/navigation";
import { NAV, type Persona } from "@/lib/nav";
import { cn } from "@/lib/utils";
import { Logo } from "./logo";
import { TenantSwitcher } from "./tenant-switcher";
import { Badge } from "@/components/ui/badge";

export function Sidebar({ persona, className }: { persona: Persona; className?: string }) {
  const t = useTranslations("nav");
  const tRoles = useTranslations("roles");
  const pathname = usePathname();
  const nav = NAV[persona];

  const isActive = (href: string, exact?: boolean) =>
    exact ? pathname === href : pathname === href || pathname.startsWith(href + "/");

  return (
    <aside className={cn("flex h-dvh w-64 shrink-0 flex-col border-r border-border bg-surface", className)}>
      <div className="flex h-16 items-center gap-2.5 border-b border-border px-5">
        <Logo size={30} />
        <div className="leading-none">
          <p className="font-display text-[15px] font-extrabold tracking-tight text-fg">
            Medhen <span className="text-brand-fg">·</span> EIC
          </p>
          <p className="mt-1 text-[11px] text-fg-muted">Motor insurance</p>
        </div>
      </div>

      <div className="px-4 pt-4">
        <Badge tone="brand" className="w-full justify-center py-1.5">{tRoles(nav.roleKey)}</Badge>
      </div>

      <nav className="flex-1 space-y-1 overflow-y-auto p-4" aria-label="Primary">
        {nav.items.map((item) => {
          const active = isActive(item.href, item.exact);
          const Icon = item.icon;
          return (
            <Link
              key={item.href}
              href={item.href}
              aria-current={active ? "page" : undefined}
              className={cn(
                "flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors",
                active
                  ? "bg-brand-subtle text-brand-fg"
                  : "text-fg-muted hover:bg-subtle hover:text-fg",
              )}
            >
              <Icon className={cn("size-[18px]", active ? "text-brand-fg" : "text-fg-subtle")} />
              {t(item.labelKey)}
            </Link>
          );
        })}
      </nav>

      <div className="border-t border-border p-4">
        <TenantSwitcher />
      </div>
    </aside>
  );
}
