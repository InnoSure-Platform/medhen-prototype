import type { LucideIcon } from "lucide-react";
import {
  Activity,
  BadgePercent,
  BookOpen,
  Calculator,
  ClipboardCheck,
  Coins,
  FileText,
  Files,
  LayoutDashboard,
  Package,
  Receipt,
  ScrollText,
  Settings,
  ShieldCheck,
  Users,
  UsersRound,
} from "lucide-react";

export type Persona = "customer" | "broker" | "staff" | "admin" | "finance";

export interface NavItem {
  href: string;
  /** key into the `nav` i18n namespace */
  labelKey: string;
  icon: LucideIcon;
  exact?: boolean;
}

export interface PersonaNav {
  persona: Persona;
  roleKey: string; // key into `roles` namespace
  items: NavItem[];
}

export const NAV: Record<Persona, PersonaNav> = {
  customer: {
    persona: "customer",
    roleKey: "customer",
    items: [
      { href: "/customer", labelKey: "dashboard", icon: LayoutDashboard, exact: true },
      { href: "/customer/quote", labelKey: "quote", icon: FileText },
      { href: "/customer/policies", labelKey: "policies", icon: ShieldCheck },
      { href: "/customer/claims", labelKey: "claims", icon: ClipboardCheck },
      { href: "/customer/invoices", labelKey: "invoices", icon: Receipt },
      { href: "/customer/documents", labelKey: "documents", icon: Files },
    ],
  },
  broker: {
    persona: "broker",
    roleKey: "broker",
    items: [
      { href: "/broker", labelKey: "book", icon: BookOpen, exact: true },
      { href: "/broker/clients", labelKey: "clients", icon: UsersRound },
      { href: "/broker/new-business", labelKey: "newBusiness", icon: FileText },
      { href: "/broker/commissions", labelKey: "commissions", icon: BadgePercent },
    ],
  },
  staff: {
    persona: "staff",
    roleKey: "staff",
    items: [
      { href: "/staff", labelKey: "dashboard", icon: LayoutDashboard, exact: true },
      { href: "/staff/quotes", labelKey: "quotes", icon: FileText },
      { href: "/staff/policies", labelKey: "policies", icon: ShieldCheck },
      { href: "/staff/claims", labelKey: "claims", icon: ClipboardCheck },
      { href: "/staff/underwriting", labelKey: "underwriting", icon: Activity },
      { href: "/staff/audit", labelKey: "audit", icon: ScrollText },
      { href: "/staff/users", labelKey: "users", icon: Users },
    ],
  },
  admin: {
    persona: "admin",
    roleKey: "admin",
    items: [
      { href: "/admin", labelKey: "dashboard", icon: LayoutDashboard, exact: true },
      { href: "/admin/users", labelKey: "users", icon: Users },
      { href: "/admin/products", labelKey: "products", icon: Package },
      { href: "/admin/settings", labelKey: "settings", icon: Settings },
    ],
  },
  finance: {
    persona: "finance",
    roleKey: "finance",
    items: [
      { href: "/finance", labelKey: "dashboard", icon: LayoutDashboard, exact: true },
      { href: "/finance/reconciliation", labelKey: "reconciliation", icon: Calculator },
      { href: "/finance/commissions", labelKey: "commissions", icon: Coins },
    ],
  },
};
