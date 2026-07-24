"use client";

import * as React from "react";
import { Command } from "cmdk";
import { useTheme } from "next-themes";
import { useTranslations } from "next-intl";
import { Moon, Search, Sun } from "lucide-react";
import { useRouter } from "@/lib/i18n/navigation";
import { NAV, type Persona } from "@/lib/nav";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

export function CommandPalette({ persona }: { persona: Persona }) {
  const [open, setOpen] = React.useState(false);
  const router = useRouter();
  const { setTheme } = useTheme();
  const t = useTranslations("nav");
  const tc = useTranslations("common");
  const items = NAV[persona].items;

  React.useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((o) => !o);
      }
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, []);

  const go = (href: string) => {
    setOpen(false);
    router.push(href);
  };

  return (
    <>
      <Button
        variant="secondary"
        size="sm"
        className="gap-2 text-fg-muted"
        onClick={() => setOpen(true)}
        aria-label={tc("commandHint")}
      >
        <Search className="size-4" />
        <span className="hidden lg:inline">{tc("commandHint")}</span>
        <kbd className="ml-2 hidden rounded border border-border bg-canvas px-1.5 font-mono text-[10px] text-fg-subtle lg:inline">⌘K</kbd>
      </Button>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent hideClose className="max-w-xl overflow-hidden p-0">
          <Command
            className="[&_[cmdk-input-wrapper]]:border-b [&_[cmdk-input-wrapper]]:border-border"
            loop
          >
            <div className="flex items-center gap-2 px-4" cmdk-input-wrapper="">
              <Search className="size-4 text-fg-subtle" />
              <Command.Input
                placeholder={tc("commandHint")}
                className="h-12 flex-1 bg-transparent text-sm text-fg outline-none placeholder:text-fg-subtle"
              />
            </div>
            <Command.List className="max-h-80 overflow-y-auto p-2">
              <Command.Empty className="py-8 text-center text-sm text-fg-muted">{tc("empty")}</Command.Empty>
              <Command.Group heading={t("dashboard")} className="[&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-xs [&_[cmdk-group-heading]]:font-semibold [&_[cmdk-group-heading]]:uppercase [&_[cmdk-group-heading]]:tracking-wider [&_[cmdk-group-heading]]:text-fg-subtle">
                {items.map((item) => {
                  const Icon = item.icon;
                  return (
                    <Command.Item
                      key={item.href}
                      value={t(item.labelKey)}
                      onSelect={() => go(item.href)}
                      className="flex cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-fg data-[selected=true]:bg-brand-subtle data-[selected=true]:text-brand-fg"
                    >
                      <Icon className="size-4 text-fg-muted" />
                      {t(item.labelKey)}
                    </Command.Item>
                  );
                })}
              </Command.Group>
              <Command.Group heading={tc("theme")} className="[&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-xs [&_[cmdk-group-heading]]:font-semibold [&_[cmdk-group-heading]]:uppercase [&_[cmdk-group-heading]]:tracking-wider [&_[cmdk-group-heading]]:text-fg-subtle">
                <Command.Item value="light theme" onSelect={() => { setTheme("light"); setOpen(false); }} className="flex cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-fg data-[selected=true]:bg-brand-subtle">
                  <Sun className="size-4 text-fg-muted" /> Light
                </Command.Item>
                <Command.Item value="dark theme" onSelect={() => { setTheme("dark"); setOpen(false); }} className="flex cursor-pointer items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm text-fg data-[selected=true]:bg-brand-subtle">
                  <Moon className="size-4 text-fg-muted" /> Dark
                </Command.Item>
              </Command.Group>
            </Command.List>
          </Command>
        </DialogContent>
      </Dialog>
    </>
  );
}
