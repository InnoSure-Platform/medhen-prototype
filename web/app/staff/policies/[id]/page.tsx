"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { useLocale } from "@/components/Shell";
import { api, errorMessage, type Policy } from "@/lib/api";
import { formatBirr } from "@/lib/i18n";

export default function PolicyDetailsPage() {
  const { locale } = useLocale();
  const params = useParams();
  const policyId = params.id as string;

  const [policy, setPolicy] = useState<Policy | null>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        setLoading(true);
        setPolicy(await api.getPolicy(locale, policyId));
      } catch (e) {
        setErr(errorMessage(e));
      } finally {
        setLoading(false);
      }
    })();
  }, [policyId, locale]);

  return (
    <div className="max-w-3xl mx-auto px-6 py-12 animate-in fade-in slide-in-from-bottom-4 duration-700">
      {err && (
        <div className="mb-6 p-4 rounded-xl bg-red-50 border border-red-200 text-red-700 text-sm font-medium" role="alert">
          {err}
        </div>
      )}

      {loading && !policy ? (
        <div className="text-slate-500">Loading…</div>
      ) : policy ? (
        <>
          <h1 className="text-3xl font-display font-bold text-slate-900 tracking-tight mb-1">
            Policy {policy.PolicyNumber}
          </h1>
          <p className="text-slate-500 mb-8">Version {policy.Version}</p>

          <div className="bg-white border border-slate-200 shadow-sm rounded-2xl p-8 grid grid-cols-1 sm:grid-cols-2 gap-6">
            <Detail label="Status" value={policy.Status} />
            <Detail label="Premium" value={formatBirr(policy.GrossPremium, locale)} />
            <Detail label="Effective from" value={new Date(policy.EffectiveFrom).toLocaleDateString()} />
            <Detail label="Effective to" value={new Date(policy.EffectiveTo).toLocaleDateString()} />
            <Detail label="Issued at" value={new Date(policy.IssuedAt).toLocaleString()} />
            <Detail label="Product" value={policy.ProductCode} />
          </div>
        </>
      ) : (
        !err && <div className="text-slate-500">Policy not found.</div>
      )}
    </div>
  );
}

function Detail({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-xs font-bold uppercase tracking-wider text-slate-500 mb-1">{label}</div>
      <div className="text-lg font-mono font-semibold text-slate-900">{value}</div>
    </div>
  );
}
