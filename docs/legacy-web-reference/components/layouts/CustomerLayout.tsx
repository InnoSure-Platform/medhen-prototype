import * as React from "react";
import Link from "next/link";
import { cn } from "@/lib/utils";

export function CustomerLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col bg-slate-50 dark:bg-slate-900">
      {/* Top Navigation */}
      <header className="sticky top-0 z-50 w-full border-b border-slate-200 bg-white/80 backdrop-blur-md dark:border-slate-800 dark:bg-slate-950/80 shadow-sm">
        <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-6">
          <Link href="/customer" className="font-display text-xl font-bold tracking-tight text-brand-blue-900 dark:text-slate-50 flex items-center gap-2">
            <span className="text-brand-gold">EIC</span> Customer Portal
          </Link>
          
          <nav className="hidden md:flex items-center gap-6">
            <NavLink href="/customer">Dashboard</NavLink>
            <NavLink href="/customer/policies">My Policies</NavLink>
            <NavLink href="/customer/claims">Claims</NavLink>
            <NavLink href="/customer/payments">Payments</NavLink>
          </nav>
          
          <div className="flex items-center gap-4">
             {/* User Menu Placeholder */}
             <div className="h-8 w-8 rounded-full bg-brand-blue-600 flex items-center justify-center text-sm font-medium text-white shadow-sm ring-2 ring-white dark:ring-slate-900">
               C
             </div>
          </div>
        </div>
      </header>

      {/* Main Content Area */}
      <main className="flex-1 w-full max-w-7xl mx-auto p-6 md:p-10">
        {children}
      </main>

      <footer className="border-t border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-950 py-8">
        <div className="mx-auto max-w-7xl px-6 text-center text-sm text-slate-500">
          &copy; {new Date().getFullYear()} Ethiopian Insurance Corporation. All rights reserved.
        </div>
      </footer>
    </div>
  );
}

function NavLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      className={cn(
        "text-sm font-medium transition-colors hover:text-brand-blue-600 dark:hover:text-brand-gold",
        "text-slate-600 dark:text-slate-300"
      )}
    >
      {children}
    </Link>
  );
}
