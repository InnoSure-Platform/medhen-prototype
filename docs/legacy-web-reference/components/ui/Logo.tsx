import * as React from "react";

/** Medhen brand mark — a protective shield (Medhen = "protector") with a gold
 *  verification tick. Crisp at any size; used in the topbar and login. */
export function Logo({ size = 34, className = "" }: { size?: number; className?: string }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 40 40"
      fill="none"
      className={className}
      aria-hidden="true"
    >
      <defs>
        <linearGradient id="medhen-shield" x1="8" y1="3" x2="32" y2="37" gradientUnits="userSpaceOnUse">
          <stop stopColor="#3b82f6" />
          <stop offset="1" stopColor="#1d4ed8" />
        </linearGradient>
      </defs>
      <path
        d="M20 3 6 8v10c0 8.5 5.7 15.3 14 18 8.3-2.7 14-9.5 14-18V8L20 3Z"
        fill="url(#medhen-shield)"
      />
      <path
        d="M20 3 6 8v10c0 8.5 5.7 15.3 14 18 8.3-2.7 14-9.5 14-18V8L20 3Z"
        stroke="#0f172a"
        strokeOpacity="0.15"
        strokeWidth="1"
      />
      <path
        d="m14 20 4 4 8-9"
        stroke="#fbbf24"
        strokeWidth="3"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}
