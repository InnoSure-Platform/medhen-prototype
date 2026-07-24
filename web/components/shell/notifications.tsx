"use client";

import * as React from "react";
import { Bell } from "lucide-react";
import { useLocale, useTranslations } from "next-intl";
import { useAuth } from "@/components/providers";
import { useAudit } from "@/lib/api/hooks";
import { relativeTime } from "@/lib/format";
import type { Locale } from "@/lib/i18n/routing";
import { Button } from "@/components/ui/button";
import { StatusBadge } from "@/components/ui/badge";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";

const SEEN_KEY = "medhen.notifications.seen";

/**
 * In-app notification center backed by the immutable audit event stream. Shows
 * an unread count for events newer than the last time the panel was opened.
 * Only rendered for roles that can read audit (staff/claims/admin) — the proxy
 * enforces the same boundary.
 */
export function Notifications() {
  const { can } = useAuth();
  const t = useTranslations();
  const locale = useLocale() as Locale;
  const [open, setOpen] = React.useState(false);
  const [lastSeen, setLastSeen] = React.useState(0);

  const enabled = can("audit:read");
  const audit = useAudit(15);

  React.useEffect(() => {
    setLastSeen(Number(localStorage.getItem(SEEN_KEY) ?? 0));
  }, []);

  const events = enabled ? (audit.data ?? []) : [];
  const unread = events.filter((e) => new Date(e.recorded_at).getTime() > lastSeen).length;

  const markSeen = () => {
    const now = Date.now();
    localStorage.setItem(SEEN_KEY, String(now));
    setLastSeen(now);
  };

  if (!enabled) return null;

  return (
    <Popover
      open={open}
      onOpenChange={(o) => {
        setOpen(o);
        if (o) markSeen();
      }}
    >
      <PopoverTrigger asChild>
        <Button variant="ghost" size="icon-sm" aria-label="Notifications" className="relative">
          <Bell className="size-4" />
          {unread > 0 && (
            <span className="absolute right-1 top-1 grid min-w-4 place-items-center rounded-full bg-danger px-1 text-[10px] font-bold leading-4 text-white">
              {unread > 9 ? "9+" : unread}
            </span>
          )}
        </Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="w-80 p-0">
        <div className="flex items-center justify-between border-b border-border px-4 py-3">
          <span className="font-semibold text-fg">{t("staff.auditTitle")}</span>
          <span className="text-xs text-fg-muted">{events.length}</span>
        </div>
        <ul className="max-h-96 divide-y divide-border-subtle overflow-y-auto">
          {events.length === 0 ? (
            <li className="px-4 py-8 text-center text-sm text-fg-muted">{t("common.empty")}</li>
          ) : (
            events.map((e) => (
              <li key={e.id} className="flex items-center justify-between gap-3 px-4 py-3">
                <StatusBadge value={e.topic} />
                <span className="whitespace-nowrap font-mono text-xs text-fg-subtle">{relativeTime(e.recorded_at, locale)}</span>
              </li>
            ))
          )}
        </ul>
      </PopoverContent>
    </Popover>
  );
}
