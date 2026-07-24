import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import * as React from "react";
import { MoneyInput } from "./money-input";

function Harness({ onChange }: { onChange: (v: number | "") => void }) {
  const [value, setValue] = React.useState<number | "">("");
  return (
    <MoneyInput
      value={value}
      onChange={(v) => {
        setValue(v);
        onChange(v);
      }}
    />
  );
}

describe("MoneyInput", () => {
  it("groups thousands for display and emits a numeric value", () => {
    const onChange = vi.fn();
    render(<Harness onChange={onChange} />);
    const input = screen.getByRole("textbox") as HTMLInputElement;

    fireEvent.change(input, { target: { value: "1500000" } });

    expect(onChange).toHaveBeenLastCalledWith(1500000);
    expect(input.value).toBe("1,500,000");
  });

  it("rejects non-numeric characters", () => {
    const onChange = vi.fn();
    render(<Harness onChange={onChange} />);
    const input = screen.getByRole("textbox") as HTMLInputElement;

    fireEvent.change(input, { target: { value: "12ab3" } });

    expect(input.value).toBe("123");
    expect(onChange).toHaveBeenLastCalledWith(123);
  });

  it("renders the ETB adornment", () => {
    render(<Harness onChange={() => {}} />);
    expect(screen.getByText("ETB")).toBeInTheDocument();
  });
});
