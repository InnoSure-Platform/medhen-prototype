import { AppShell } from "@/components/shell/app-shell";

export default function StaffLayout({ children }: { children: React.ReactNode }) {
  return <AppShell persona="staff">{children}</AppShell>;
}
