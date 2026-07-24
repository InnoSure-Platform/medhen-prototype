import { cn } from "@/lib/utils";

/** Medhen / EIC brand mark — a shield (protection) with a subtle motion arc. */
export function Logo({ size = 32, className }: { size?: number; className?: string }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 40 40"
      fill="none"
      className={cn("shrink-0", className)}
      role="img"
      aria-label="Medhen"
    >
      <path
        d="M20 2.5 5.5 8.2v11.2c0 8.9 6.1 15.6 14.5 18.1 8.4-2.5 14.5-9.2 14.5-18.1V8.2L20 2.5Z"
        fill="var(--brand-default)"
      />
      <path
        d="M20 2.5 5.5 8.2v11.2c0 8.9 6.1 15.6 14.5 18.1V2.5Z"
        fill="var(--brand-hover)"
      />
      <path
        d="M13.5 20.5 18 25l9-10"
        stroke="var(--accent-default)"
        strokeWidth="2.75"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
