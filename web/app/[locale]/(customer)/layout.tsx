import { AppShell } from "@/components/shell/app-shell";

export default function CustomerLayout({ children }: { children: React.ReactNode }) {
  return <AppShell persona="customer">{children}</AppShell>;
}
