import * as React from "react";
import Link from "next/link";
import { cn } from "@/lib/utils";

export function AdminLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen flex-col bg-slate-50 dark:bg-slate-900 md:flex-row">
      {/* Sidebar Navigation */}
      <aside className="w-full shrink-0 border-b border-slate-200 bg-brand-blue-900 text-white md:w-64 md:border-b-0 md:border-r md:border-slate-800 dark:bg-slate-950">
        <div className="flex h-16 items-center px-6">
          <Link href="/admin" className="font-display text-xl font-bold tracking-tight text-white flex items-center gap-2">
            <span className="text-brand-gold">EIC</span> Admin
          </Link>
        </div>
        <nav className="flex flex-col gap-1 p-4">
          <NavLink href="/admin">Dashboard</NavLink>
          <NavLink href="/admin/underwriting">Underwriting</NavLink>
          <NavLink href="/admin/policies">Policies</NavLink>
          <NavLink href="/admin/claims">Claims Management</NavLink>
          <NavLink href="/admin/reports">Reports</NavLink>
        </nav>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 flex flex-col min-h-[calc(100vh-4rem)] md:min-h-screen">
        <header className="flex h-16 items-center justify-between border-b border-slate-200 bg-white/70 px-6 backdrop-blur-md dark:border-slate-800 dark:bg-slate-900/70">
          <h2 className="font-display text-lg font-semibold text-slate-800 dark:text-slate-100">Management Portal</h2>
          <div className="flex items-center gap-4">
             {/* User Menu Placeholder */}
             <div className="h-8 w-8 rounded-full bg-brand-blue-600 flex items-center justify-center text-sm font-medium text-white shadow-sm">
               A
             </div>
          </div>
        </header>
        <div className="flex-1 p-6 lg:p-10 max-w-7xl mx-auto w-full">
          {children}
        </div>
      </main>
    </div>
  );
}

function NavLink({ href, children }: { href: string; children: React.ReactNode }) {
  return (
    <Link
      href={href}
      className={cn(
        "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
        "text-slate-300 hover:bg-white/10 hover:text-white"
      )}
    >
      {children}
    </Link>
  );
}
