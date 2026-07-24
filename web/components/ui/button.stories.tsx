import type { Meta, StoryObj } from "@storybook/react";
import { ArrowRight, Plus } from "lucide-react";
import { Button } from "./button";

const meta: Meta<typeof Button> = {
  title: "Primitives/Button",
  component: Button,
  tags: ["autodocs"],
  args: { children: "Get a quote" },
  argTypes: {
    variant: { control: "select", options: ["primary", "secondary", "outline", "ghost", "gold", "danger", "link"] },
    size: { control: "select", options: ["sm", "md", "lg", "icon", "icon-sm"] },
  },
};
export default meta;
type Story = StoryObj<typeof Button>;

export const Primary: Story = {};

export const Variants: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      <Button variant="primary">Primary</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="gold">Gold</Button>
      <Button variant="danger">Danger</Button>
      <Button variant="link">Link</Button>
    </div>
  ),
};

export const Sizes: Story = {
  render: () => (
    <div className="flex flex-wrap items-center gap-3">
      <Button size="sm">Small</Button>
      <Button size="md">Medium</Button>
      <Button size="lg">
        Large <ArrowRight />
      </Button>
      <Button size="icon" aria-label="Add">
        <Plus />
      </Button>
    </div>
  ),
};

export const Loading: Story = { args: { loading: true, children: "Processing" } };
export const Disabled: Story = { args: { disabled: true } };
