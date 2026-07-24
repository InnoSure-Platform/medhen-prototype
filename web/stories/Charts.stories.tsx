import type { Meta, StoryObj } from "@storybook/react";
import { BarChart, DonutChart, TrendAreaChart } from "@/components/ui/charts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";

const meta: Meta = { title: "Charts/Gallery", parameters: { layout: "padded" } };
export default meta;
type Story = StoryObj;

const monthly = [
  { month: "Jan", premium: 4.2, claims: 2.1 },
  { month: "Feb", premium: 5.1, claims: 2.8 },
  { month: "Mar", premium: 4.9, claims: 3.4 },
  { month: "Apr", premium: 6.2, claims: 3.0 },
  { month: "May", premium: 7.0, claims: 4.1 },
  { month: "Jun", premium: 8.4, claims: 4.6 },
];

export const PremiumVsClaims: Story = {
  render: () => (
    <Card className="max-w-2xl">
      <CardHeader>
        <CardTitle>Premium vs. claims (M ETB)</CardTitle>
      </CardHeader>
      <CardContent>
        <TrendAreaChart data={monthly} xKey="month" series={[{ key: "premium", name: "Premium" }, { key: "claims", name: "Claims" }]} />
      </CardContent>
    </Card>
  ),
};

export const PolicyCountByMonth: Story = {
  render: () => (
    <Card className="max-w-2xl">
      <CardHeader>
        <CardTitle>Policies bound</CardTitle>
      </CardHeader>
      <CardContent>
        <BarChart data={monthly} xKey="month" series={[{ key: "premium", name: "Policies" }]} />
      </CardContent>
    </Card>
  ),
};

export const CoverageMix: Story = {
  render: () => (
    <Card className="max-w-sm">
      <CardHeader>
        <CardTitle>Coverage mix</CardTitle>
      </CardHeader>
      <CardContent>
        <DonutChart data={[{ name: "Comprehensive", value: 62 }, { name: "Third-party", value: 38 }]} />
      </CardContent>
    </Card>
  ),
};
