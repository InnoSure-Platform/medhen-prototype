"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import { inputClass } from "./input";

export interface MoneyInputProps {
  id?: string;
  value: number | "";
  onChange: (value: number | "") => void;
  min?: number;
  max?: number;
  placeholder?: string;
  "aria-invalid"?: boolean;
  "aria-describedby"?: string;
  className?: string;
  disabled?: boolean;
}

/**
 * Validated Birr (ETB) money input. Displays thousands separators, accepts only
 * digits/decimal, and reports a numeric major-unit value (or "" when empty).
 * No floating-point money math happens here — the value is passed as entered.
 */
export function MoneyInput({ id, value, onChange, min, max, placeholder = "0.00", className, disabled, ...aria }: MoneyInputProps) {
  const [text, setText] = React.useState(value === "" ? "" : String(value));

  React.useEffect(() => {
    // Keep display in sync when the value changes externally.
    const asNum = text === "" ? "" : Number(text.replace(/,/g, ""));
    if (asNum !== value) setText(value === "" ? "" : formatDisplay(String(value)));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  function handle(raw: string) {
    const cleaned = raw.replace(/[^0-9.]/g, "");
    setText(formatDisplay(cleaned));
    if (cleaned === "" || cleaned === ".") return onChange("");
    let n = Number(cleaned);
    if (Number.isNaN(n)) return;
    if (min != null && n < min) n = n; // clamp on blur, not while typing
    onChange(n);
  }

  function handleBlur() {
    if (value === "") return;
    let n = value as number;
    if (min != null && n < min) n = min;
    if (max != null && n > max) n = max;
    onChange(n);
    setText(formatDisplay(String(n)));
  }

  return (
    <div className="relative">
      <input
        id={id}
        inputMode="decimal"
        value={text}
        onChange={(e) => handle(e.target.value)}
        onBlur={handleBlur}
        placeholder={placeholder}
        disabled={disabled}
        className={cn(inputClass, "pr-14 text-right font-mono tabular-nums", className)}
        {...aria}
      />
      <span className="pointer-events-none absolute right-3.5 top-1/2 -translate-y-1/2 text-xs font-semibold text-fg-muted">
        ETB
      </span>
    </div>
  );
}

function formatDisplay(cleaned: string): string {
  if (cleaned === "") return "";
  const [int, dec] = cleaned.split(".");
  const grouped = int.replace(/\B(?=(\d{3})+(?!\d))/g, ",");
  return dec != null ? `${grouped}.${dec.slice(0, 2)}` : grouped;
}
