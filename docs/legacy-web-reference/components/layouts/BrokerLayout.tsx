import * as React from "react";
import Link from "next/link";
import { cn } from "@/lib/utils";

export function BrokerLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col bg-slate-50 dark:bg-slate-900">
      {/* Top Navigation */}
      <header className="sticky top-0 z-50 w-full border-b border-slate-200 bg-brand-blue-900 shadow-sm dark:border-slate-800 dark:bg-slate-950">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
          <Link href="/broker" className="font-display text-xl font-bold tracking-tight text-white flex items-center gap-2">
            <span className="text-brand-gold">EIC</span> Broker Portal
          </Link>
          
          <nav className="hidden md:flex items-center gap-6">
            <NavLink href="/broker">Dashboard</NavLink>
            <NavLink href="/broker/clients">Clients</NavLink>
            <NavLink href="/broker/quotes">Quotes</NavLink>
            <NavLink href="/broker/commissions">Commissions</NavLink>
          </nav>
          
          <div className="flex items-center gap-4">
             {/* User Menu Placeholder */}
             <div className="h-8 w-8 rounded-full bg-white flex items-center justify-center text-sm font-bold text-brand-blue-900 shadow-sm">
               B
             </div>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 w-full max-w-7xl mx-auto p-6 md:p-10">
        {children}
      </main>
    </div>
  );
}

function NavLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      className={cn(
        "text-sm font-medium transition-colors hover:text-brand-gold",
        "text-slate-300"
      )}
    >
      {children}
    </Link>
  );
}
