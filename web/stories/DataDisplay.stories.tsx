import type { Meta, StoryObj } from "@storybook/react";
import { ShieldCheck, TrendingUp, Wallet } from "lucide-react";
import { Badge, StatusBadge } from "@/components/ui/badge";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { StatCard } from "@/components/ui/stat-card";
import { Timeline } from "@/components/ui/timeline";
import { Stepper } from "@/components/ui/stepper";

const meta: Meta = { title: "Data Display/Gallery", parameters: { layout: "padded" } };
export default meta;
type Story = StoryObj;

export const Badges: Story = {
  render: () => (
    <div className="flex flex-wrap gap-2">
      <Badge tone="brand">Brand</Badge>
      <Badge tone="success" dot>Active</Badge>
      <Badge tone="warning">Open</Badge>
      <Badge tone="danger">Rejected</Badge>
      <Badge tone="info">Quoted</Badge>
      <StatusBadge value="ISSUED" />
      <StatusBadge value="PARTIALLY_PAID" />
      <StatusBadge value="policy.issued" />
    </div>
  ),
};

export const Stats: Story = {
  render: () => (
    <div className="grid gap-4 sm:grid-cols-3">
      <StatCard label="Policies in force" value="1,284" icon={<ShieldCheck />} delta={{ value: "12%", direction: "up" }} trend={[10, 12, 11, 15, 18, 22, 28]} />
      <StatCard label="Gross written premium" value="8.4M ETB" tone="success" icon={<Wallet />} delta={{ value: "6.2%", direction: "up" }} trend={[4, 5, 6, 6, 7, 8, 8.4]} />
      <StatCard label="Loss ratio" value="72.0%" tone="warning" icon={<TrendingUp />} delta={{ value: "3%", direction: "up", good: false }} hint="vs last month" />
    </div>
  ),
};

export const CardStory: Story = {
  name: "Card",
  render: () => (
    <Card className="max-w-sm" interactive>
      <CardHeader>
        <CardTitle>Motor · Comprehensive</CardTitle>
        <CardDescription>EIC/MOT/2026/000001</CardDescription>
      </CardHeader>
      <CardContent className="flex items-center justify-between">
        <span className="text-fg-muted">Gross premium</span>
        <span className="font-mono text-lg font-bold text-brand-fg">2,680.00 ETB</span>
      </CardContent>
    </Card>
  ),
};

export const LifecycleTimeline: Story = {
  render: () => (
    <Timeline
      className="max-w-md"
      items={[
        { title: "Quote created", description: "STP underwriting passed", timestamp: "09:12", state: "done" },
        { title: "Policy bound & issued", description: "EIC/MOT/2026/000001", timestamp: "09:13", state: "done" },
        { title: "Invoice raised", description: "2,680.00 ETB — OPEN", timestamp: "09:13", state: "current" },
        { title: "Telebirr payment", description: "Awaiting settlement", state: "upcoming" },
      ]}
    />
  ),
};

export const WizardStepper: Story = {
  render: () => (
    <div className="pb-8">
      <Stepper
        current={2}
        steps={[
          { id: "identity", label: "Identity" },
          { id: "vehicle", label: "Vehicle" },
          { id: "quote", label: "Quotation" },
          { id: "done", label: "Complete" },
        ]}
      />
    </div>
  ),
};
