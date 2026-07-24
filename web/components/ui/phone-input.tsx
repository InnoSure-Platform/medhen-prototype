"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import { inputClass } from "./input";

export interface PhoneInputProps {
  id?: string;
  value: string;
  onChange: (value: string) => void;
  className?: string;
  disabled?: boolean;
  "aria-invalid"?: boolean;
  "aria-describedby"?: string;
}

/**
 * Ethiopian phone input that normalizes to E.164 (+251…). Accepts local formats
 * (09…, 9…) and produces the canonical +2519XXXXXXXX the backend expects.
 */
export function PhoneInput({ id, value, onChange, className, disabled, ...aria }: PhoneInputProps) {
  return (
    <div className="flex">
      <span className="inline-flex items-center rounded-l-xl border border-r-0 border-border bg-subtle px-3 font-mono text-sm text-fg-muted">
        +251
      </span>
      <input
        id={id}
        inputMode="tel"
        value={value.replace(/^\+251/, "")}
        onChange={(e) => onChange(normalize(e.target.value))}
        placeholder="911234567"
        disabled={disabled}
        className={cn(inputClass, "rounded-l-none font-mono", className)}
        {...aria}
      />
    </div>
  );
}

function normalize(local: string): string {
  const digits = local.replace(/\D/g, "").replace(/^251/, "").replace(/^0/, "");
  return `+251${digits}`.slice(0, 13);
}
