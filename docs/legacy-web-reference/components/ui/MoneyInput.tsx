"use client";

import { useId } from "react";

// MoneyInput is a validated Birr (ETB) money field (M7): it replaces ad-hoc
// prompt()/parseInt entry with a numeric input that enforces a minimum, a step,
// and a non-negative value, and carries a proper <label htmlFor> association for
// accessibility. The value is held in major units (Birr); callers convert to
// minor units at the API boundary.
export function MoneyInput({
  label,
  value,
  onChange,
  min = 0,
  step = 100,
  disabled,
  placeholder,
  required,
}: {
  label: string;
  value: number;
  onChange: (birr: number) => void;
  min?: number;
  step?: number;
  disabled?: boolean;
  placeholder?: string;
  required?: boolean;
}) {
  const id = useId();
  const invalid = Number.isNaN(value) || value < min;

  return (
    <div>
      <label htmlFor={id} className="label-text">
        {label}
      </label>
      <div className="relative">
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <span className="text-slate-500 font-semibold text-sm">ETB</span>
        </div>
        <input
          id={id}
          className="input-field pl-12 font-mono font-medium"
          type="number"
          inputMode="decimal"
          min={min}
          step={step}
          value={Number.isNaN(value) ? "" : value}
          disabled={disabled}
          placeholder={placeholder}
          required={required}
          aria-invalid={invalid}
          onChange={(e) => {
            const parsed = e.target.valueAsNumber;
            onChange(Number.isNaN(parsed) ? NaN : parsed);
          }}
        />
      </div>
      {invalid && (
        <p className="mt-1 text-xs font-medium text-red-600">
          Enter an amount of at least {min.toLocaleString()} ETB.
        </p>
      )}
    </div>
  );
}
