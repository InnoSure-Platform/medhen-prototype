"use client";

import * as React from "react";
import { Menu } from "lucide-react";
import { type Persona } from "@/lib/nav";
import { Sidebar } from "./sidebar";
import { CommandPalette } from "./command-palette";
import { ThemeToggle } from "./theme-toggle";
import { LocaleToggle } from "./locale-toggle";
import { Notifications } from "./notifications";
import { UserMenu } from "./user-menu";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogTitle, DialogTrigger } from "@/components/ui/dialog";

export function AppShell({ persona, children }: { persona: Persona; children: React.ReactNode }) {
  return (
    <div className="flex min-h-dvh bg-canvas">
      {/* Desktop sidebar */}
      <div className="sticky top-0 hidden h-dvh lg:block">
        <Sidebar persona={persona} />
      </div>

      <div className="flex min-w-0 flex-1 flex-col">
        {/* Topbar */}
        <header className="sticky top-0 z-40 flex h-16 items-center gap-3 border-b border-border bg-surface/80 px-4 backdrop-blur-xl sm:px-6">
          {/* Mobile menu */}
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="ghost" size="icon-sm" className="lg:hidden" aria-label="Open menu">
                <Menu className="size-5" />
              </Button>
            </DialogTrigger>
            <DialogContent variant="sheet" hideClose className="left-0 right-auto w-64 p-0">
              <DialogTitle className="sr-only">Navigation</DialogTitle>
              <Sidebar persona={persona} className="w-full border-r-0" />
            </DialogContent>
          </Dialog>

          <div className="flex-1">
            <CommandPalette persona={persona} />
          </div>

          <div className="flex items-center gap-1 sm:gap-2">
            <Notifications />
            <LocaleToggle />
            <ThemeToggle />
            <div className="mx-1 h-6 w-px bg-border" />
            <UserMenu />
          </div>
        </header>

        <main className="flex-1">{children}</main>
      </div>
    </div>
  );
}
