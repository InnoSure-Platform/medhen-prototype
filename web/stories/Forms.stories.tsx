import type { Meta, StoryObj } from "@storybook/react";
import * as React from "react";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { MoneyInput } from "@/components/ui/money-input";
import { PhoneInput } from "@/components/ui/phone-input";
import { Checkbox } from "@/components/ui/checkbox";
import { Switch } from "@/components/ui/switch";
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Button } from "@/components/ui/button";

const meta: Meta = { title: "Primitives/Form controls", parameters: { layout: "padded" } };
export default meta;
type Story = StoryObj;

export const Gallery: Story = {
  render: () => {
    const [money, setMoney] = React.useState<number | "">(1500000);
    const [phone, setPhone] = React.useState("+251911234567");
    return (
      <form className="grid max-w-xl gap-5" onSubmit={(e) => e.preventDefault()}>
        <Field label="Full name" htmlFor="f-name" required>
          <Input id="f-name" defaultValue="Abebe Kebede" />
        </Field>
        <Field label="Name (Amharic)" htmlFor="f-name-am" hint="Rendered in Noto Sans Ethiopic">
          <Input id="f-name-am" defaultValue="አበበ ከበደ" className="font-ethiopic" />
        </Field>
        <Field label="Phone (Telebirr)" htmlFor="f-phone">
          <PhoneInput id="f-phone" value={phone} onChange={setPhone} />
        </Field>
        <Field label="Sum insured" htmlFor="f-money" error="Must be at least 1,000 ETB">
          <MoneyInput id="f-money" value={money} onChange={setMoney} min={1000} aria-invalid />
        </Field>
        <Field label="Cover type" htmlFor="f-cover">
          <Select defaultValue="comprehensive">
            <SelectTrigger id="f-cover">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="comprehensive">Comprehensive (OD + TPL)</SelectItem>
              <SelectItem value="tpl">Third-party only (TPL)</SelectItem>
            </SelectContent>
          </Select>
        </Field>
        <Field label="Notes" htmlFor="f-notes">
          <Textarea id="f-notes" placeholder="Any additional details…" />
        </Field>
        <fieldset className="space-y-2">
          <legend className="text-sm font-medium text-fg">Driver age band</legend>
          <RadioGroup defaultValue="adult" className="flex gap-6">
            {["young", "adult", "senior"].map((v) => (
              <div key={v} className="flex items-center gap-2">
                <RadioGroupItem value={v} id={`age-${v}`} />
                <Label htmlFor={`age-${v}`} className="capitalize">{v}</Label>
              </div>
            ))}
          </RadioGroup>
        </fieldset>
        <div className="flex items-center gap-3">
          <Checkbox id="f-terms" defaultChecked />
          <Label htmlFor="f-terms">I accept the policy terms</Label>
        </div>
        <div className="flex items-center gap-3">
          <Switch id="f-paperless" defaultChecked />
          <Label htmlFor="f-paperless">Paperless documents</Label>
        </div>
        <Button type="submit" className="w-fit">Calculate premium</Button>
      </form>
    );
  },
};
