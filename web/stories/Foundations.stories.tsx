import type { Meta, StoryObj } from "@storybook/react";

const meta: Meta = {
  title: "Foundations/Overview",
  parameters: { layout: "fullscreen" },
};
export default meta;
type Story = StoryObj;

function Swatch({ token, label }: { token: string; label: string }) {
  return (
    <div className="flex flex-col gap-1.5">
      <div
        className="h-16 w-full rounded-xl border border-border shadow-[var(--shadow-soft)]"
        style={{ background: `var(--${token})` }}
      />
      <span className="text-xs font-semibold text-fg">{label}</span>
      <span className="font-mono text-[10px] text-fg-muted">--{token}</span>
    </div>
  );
}

const semantic: [string, string][] = [
  ["bg-canvas", "Canvas"],
  ["bg-surface", "Surface"],
  ["bg-subtle", "Subtle"],
  ["brand-default", "Brand"],
  ["brand-hover", "Brand hover"],
  ["accent-default", "Accent (gold)"],
  ["success-default", "Success"],
  ["warning-default", "Warning"],
  ["danger-default", "Danger"],
  ["info-default", "Info"],
  ["fg-default", "Foreground"],
  ["fg-muted", "Foreground muted"],
];

const chart: [string, string][] = [
  ["chart-1", "Chart 1"],
  ["chart-2", "Chart 2"],
  ["chart-3", "Chart 3"],
  ["chart-4", "Chart 4"],
  ["chart-5", "Chart 5"],
  ["chart-6", "Chart 6"],
];

export const Colors: Story = {
  render: () => (
    <div className="min-h-dvh bg-canvas p-10 text-fg">
      <h1 className="text-3xl">Semantic color tokens</h1>
      <p className="mt-1 text-fg-muted">
        Flip the theme toolbar to see light/dark. Every token is generated from{" "}
        <code className="font-mono text-brand-fg">tokens/src/*.json</code>.
      </p>
      <div className="mt-8 grid grid-cols-2 gap-5 sm:grid-cols-3 lg:grid-cols-6">
        {semantic.map(([token, label]) => (
          <Swatch key={token} token={token} label={label} />
        ))}
      </div>
      <h2 className="mt-12 text-2xl">Categorical chart palette</h2>
      <div className="mt-6 grid grid-cols-2 gap-5 sm:grid-cols-3 lg:grid-cols-6">
        {chart.map(([token, label]) => (
          <Swatch key={token} token={token} label={label} />
        ))}
      </div>
    </div>
  ),
};

export const Typography: Story = {
  render: () => (
    <div className="min-h-dvh bg-canvas p-10 text-fg">
      <h1 className="text-3xl">Typography</h1>
      <div className="mt-8 space-y-6">
        <div>
          <span className="text-xs uppercase tracking-widest text-fg-muted">Display · Plus Jakarta Sans</span>
          <p className="font-display text-5xl font-bold tracking-tight">Motor, quote to claim</p>
        </div>
        <div>
          <span className="text-xs uppercase tracking-widest text-fg-muted">Body · Inter</span>
          <p className="text-lg">Enterprise-grade insurance operations for the Ethiopian Insurance Corporation.</p>
        </div>
        <div>
          <span className="text-xs uppercase tracking-widest text-fg-muted">Amharic · Noto Sans Ethiopic</span>
          <p className="font-ethiopic text-2xl">የመኪና መድን፣ ከቅናሽ እስከ ይገባኛል።</p>
        </div>
        <div>
          <span className="text-xs uppercase tracking-widest text-fg-muted">Mono · IDs & money</span>
          <p className="font-mono text-lg text-brand-fg">EIC/MOT/2026/000001 · 2,680.00 ETB</p>
        </div>
      </div>
    </div>
  ),
};

export const Elevation: Story = {
  render: () => (
    <div className="min-h-dvh bg-canvas p-10 text-fg">
      <h1 className="text-3xl">Elevation</h1>
      <div className="mt-8 grid grid-cols-1 gap-8 sm:grid-cols-3">
        {[
          ["soft", "Soft"],
          ["card", "Card"],
          ["lift", "Lift"],
        ].map(([token, label]) => (
          <div key={token} className="flex flex-col items-center gap-3">
            <div
              className="grid h-28 w-full place-items-center rounded-2xl border border-border bg-surface"
              style={{ boxShadow: `var(--shadow-${token})` }}
            >
              <span className="text-sm font-semibold">{label}</span>
            </div>
            <code className="font-mono text-xs text-fg-muted">--shadow-{token}</code>
          </div>
        ))}
      </div>
    </div>
  ),
};
