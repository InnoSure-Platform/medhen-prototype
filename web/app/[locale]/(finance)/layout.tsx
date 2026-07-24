import { AppShell } from "@/components/shell/app-shell";

export default function FinanceLayout({ children }: { children: React.ReactNode }) {
  return <AppShell persona="finance">{children}</AppShell>;
}
