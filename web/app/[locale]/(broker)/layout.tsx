import { AppShell } from "@/components/shell/app-shell";

export default function BrokerLayout({ children }: { children: React.ReactNode }) {
  return <AppShell persona="broker">{children}</AppShell>;
}
