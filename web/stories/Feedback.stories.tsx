import type { Meta, StoryObj } from "@storybook/react";
import { toast } from "sonner";
import { Alert } from "@/components/ui/alert";
import { Skeleton } from "@/components/ui/skeleton";
import { Progress } from "@/components/ui/progress";
import { EmptyState, ErrorState } from "@/components/ui/states";
import { Button } from "@/components/ui/button";

const meta: Meta = { title: "Feedback/Gallery", parameters: { layout: "padded" } };
export default meta;
type Story = StoryObj;

export const Alerts: Story = {
  render: () => (
    <div className="grid max-w-lg gap-3">
      <Alert tone="info" title="Heads up">Your session refreshes automatically.</Alert>
      <Alert tone="success" title="Policy issued">EIC/MOT/2026/000001 is now active.</Alert>
      <Alert tone="warning" title="Invoice open">Settle via Telebirr to activate cover.</Alert>
      <Alert tone="danger" title="Settlement failed">The amount exceeds your authority limit.</Alert>
    </div>
  ),
};

export const Toasts: Story = {
  render: () => (
    <div className="flex gap-2">
      <Button onClick={() => toast.success("Policy issued", { description: "EIC/MOT/2026/000001" })}>Success</Button>
      <Button variant="secondary" onClick={() => toast.error("Something went wrong")}>Error</Button>
    </div>
  ),
};

export const Loading: Story = {
  render: () => (
    <div className="max-w-md space-y-3">
      <Skeleton className="h-8 w-1/2" />
      <Skeleton className="h-24 w-full" />
      <Skeleton className="h-5 w-3/4" />
      <Progress value={64} />
    </div>
  ),
};

export const Empty: Story = {
  render: () => <EmptyState title="No policies yet" description="Create your first motor quote to get started." action={{ label: "Start a quote" }} />,
};

export const Error: Story = {
  render: () => <ErrorState title="Couldn’t load KPIs" description="The service is temporarily unavailable." action={{ label: "Try again" }} />,
};
