"use client";

import * as React from "react";
import { type ColumnDef } from "@tanstack/react-table";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { UserPlus } from "lucide-react";
import { useCreateUser, useUsers } from "@/lib/api/hooks";
import { errorMessage } from "@/lib/api/client";
import type { IamUser } from "@/lib/api/types";
import { PageHeader, Eyebrow } from "@/components/patterns/page-header";
import { DataTable } from "@/components/ui/data-table";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { ErrorState } from "@/components/ui/states";
import { Button } from "@/components/ui/button";
import { Field } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";

const ALL_ROLES = ["customer", "agent", "staff", "claims", "admin"];

export default function AdminUsersPage() {
  const t = useTranslations();
  const tErr = useTranslations("errors");
  const users = useUsers();
  const createUser = useCreateUser();

  const [open, setOpen] = React.useState(false);
  const [form, setForm] = React.useState({ subject: "", email: "", full_name: "", roles: new Set<string>(["customer"]) });

  const toggleRole = (r: string) =>
    setForm((f) => {
      const roles = new Set(f.roles);
      if (roles.has(r)) roles.delete(r);
      else roles.add(r);
      return { ...f, roles };
    });

  async function submit() {
    try {
      await createUser.mutateAsync({
        subject: form.subject.trim(),
        email: form.email.trim() || undefined,
        full_name: form.full_name.trim() || undefined,
        roles: [...form.roles],
      });
      toast.success(t("admin.createUser"));
      setOpen(false);
      setForm({ subject: "", email: "", full_name: "", roles: new Set(["customer"]) });
    } catch (e) {
      toast.error(errorMessage(e, tErr));
    }
  }

  const columns: ColumnDef<IamUser>[] = React.useMemo(
    () => [
      { accessorKey: "full_name", header: t("quote.fullName"), cell: ({ row }) => <span className="font-medium text-fg">{row.original.full_name || row.original.subject}</span> },
      { accessorKey: "email", header: "Email", cell: ({ row }) => <span className="text-fg-muted">{row.original.email || "—"}</span> },
      {
        id: "roles",
        header: t("admin.roles"),
        cell: ({ row }) => (
          <div className="flex flex-wrap gap-1">{(row.original.roles ?? []).map((r) => <Badge key={r} tone="brand">{r}</Badge>)}</div>
        ),
      },
    ],
    [t],
  );

  return (
    <div className="mx-auto max-w-6xl space-y-8 px-6 py-8">
      <PageHeader
        eyebrow={<Eyebrow>{t("admin.eyebrow")}</Eyebrow>}
        title={t("admin.usersTitle")}
        actions={
          <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger asChild>
              <Button><UserPlus /> {t("admin.addUser")}</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader><DialogTitle>{t("admin.addUser")}</DialogTitle></DialogHeader>
              <div className="space-y-4">
                <Field label={t("admin.subject")} htmlFor="subject" required>
                  <Input id="subject" value={form.subject} onChange={(e) => setForm({ ...form, subject: e.target.value })} placeholder="keycloak subject / username" />
                </Field>
                <Field label="Email" htmlFor="email">
                  <Input id="email" type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
                </Field>
                <Field label={t("quote.fullName")} htmlFor="full_name">
                  <Input id="full_name" value={form.full_name} onChange={(e) => setForm({ ...form, full_name: e.target.value })} />
                </Field>
                <div className="space-y-2">
                  <Label>{t("admin.roles")}</Label>
                  <div className="grid grid-cols-2 gap-2">
                    {ALL_ROLES.map((r) => (
                      <label key={r} className="flex items-center gap-2 text-sm text-fg">
                        <Checkbox checked={form.roles.has(r)} onCheckedChange={() => toggleRole(r)} /> {r}
                      </label>
                    ))}
                  </div>
                </div>
              </div>
              <DialogFooter>
                <Button variant="secondary" onClick={() => setOpen(false)}>{t("common.cancel")}</Button>
                <Button onClick={submit} loading={createUser.isPending} disabled={!form.subject.trim() || form.roles.size === 0}>
                  {t("admin.createUser")}
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        }
      />
      {users.isError ? (
        <Card><ErrorState title={t("errors.boundaryTitle")} description={errorMessage(users.error, tErr)} action={{ label: t("common.retry"), onClick: () => users.refetch() }} /></Card>
      ) : (
        <DataTable columns={columns} data={users.data ?? []} loading={users.isLoading} filterColumn="full_name" filterPlaceholder={t("common.search")} emptyTitle={t("common.empty")} />
      )}
    </div>
  );
}
