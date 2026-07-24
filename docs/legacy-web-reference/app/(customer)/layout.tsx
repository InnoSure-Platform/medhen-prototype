import { CustomerLayout } from "@/components/layouts/CustomerLayout";

export default function Layout({ children }: { children: React.ReactNode }) {
  return <CustomerLayout>{children}</CustomerLayout>;
}
