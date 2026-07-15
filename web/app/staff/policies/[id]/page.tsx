"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { useLocale } from "@/components/Shell";
import { api } from "@/lib/api";
import { formatETB } from "@/lib/i18n";

export default function PolicyDetailsPage() {
  const { locale } = useLocale();
  const params = useParams();
  const router = useRouter();
  const policyId = params.id as string;

  const [policy, setPolicy] = useState<any>(null);
  const [err, setErr] = useState("");
  const [loading, setLoading] = useState(false);
  const [showEndorseModal, setShowEndorseModal] = useState(false);

  // Endorse Form State
  const [sumInsured, setSumInsured] = useState<number>(0);

  useEffect(() => {
    fetchPolicy();
  }, [policyId]);

  async function fetchPolicy() {
    try {
      setLoading(true);
      const p = await api.getPolicy(locale, policyId);
      setPolicy(p);
      setSumInsured(p.risk?.sumInsuredMinor ? p.risk.sumInsuredMinor / 100 : 0);
    } catch (e) {
      setErr(String(e));
    } finally {
      setLoading(false);
    }
  }

  async function handleRenew() {
    try {
      setLoading(true);
      const newPol = await api.renewPolicy(locale, policyId);
      alert(`Policy Renewed! New version: ${newPol.version}`);
      router.push(`/staff/policies/${newPol.id}`);
    } catch (e) {
      setErr(String(e));
    } finally {
      setLoading(false);
    }
  }

  async function handleCancel() {
    if (!confirm("Are you sure you want to cancel this policy?")) return;
    try {
      setLoading(true);
      await api.cancelPolicy(locale, policyId);
      alert("Policy Cancelled successfully.");
      await fetchPolicy();
    } catch (e) {
      setErr(String(e));
    } finally {
      setLoading(false);
    }
  }

  async function submitEndorsement(e: React.FormEvent) {
    e.preventDefault();
    try {
      setLoading(true);
      const updatedRisk = { ...policy.risk, sumInsuredMinor: sumInsured * 100 };
      const newPol = await api.endorsePolicy(locale, policyId, updatedRisk);
      setShowEndorseModal(false);
      alert(`Policy Endorsed! Pro-rata diff applied. New version: ${newPol.version}`);
      router.push(`/staff/policies/${newPol.id}`);
    } catch (e) {
      setErr(String(e));
    } finally {
      setLoading(false);
    }
  }

  if (!policy) return <div className="section">Loading...</div>;

  return (
    <div className="section">
      <h1>Policy: {policy.policyNumber} (v{policy.version || 1})</h1>
      {err && <div className="banner-err">{err}</div>}

      <div className="card" style={{ marginTop: "1rem" }}>
        <h3>Details</h3>
        <p><strong>Status:</strong> {policy.status}</p>
        <p><strong>Effective:</strong> {policy.effectiveFrom} to {policy.effectiveTo}</p>
        <p><strong>Premium:</strong> {formatETB(policy.totalMinor, locale)}</p>
        <p><strong>Sum Insured:</strong> {formatETB(policy.risk?.sumInsuredMinor || 0, locale)}</p>
      </div>

      <div style={{ marginTop: "2rem", display: "flex", gap: "1rem" }}>
        <button 
          className="btn" 
          onClick={() => setShowEndorseModal(true)} 
          disabled={loading || policy.status !== "ISSUED"}
        >
          Endorse (Mid-Term)
        </button>
        <button 
          className="btn btn-secondary" 
          onClick={handleRenew} 
          disabled={loading || policy.status !== "ISSUED"}
        >
          Renew
        </button>
        <button 
          className="btn" 
          style={{ background: "#cc0000" }} 
          onClick={handleCancel} 
          disabled={loading || policy.status !== "ISSUED"}
        >
          Cancel
        </button>
      </div>

      {showEndorseModal && (
        <div className="modal-overlay" style={{
          position: "fixed", top: 0, left: 0, width: "100%", height: "100%", 
          background: "rgba(0,0,0,0.5)", display: "flex", alignItems: "center", justifyContent: "center"
        }}>
          <div className="card" style={{ width: "400px", background: "white" }}>
            <h2>Endorse Policy</h2>
            <p>Update risk details to generate a mid-term adjustment.</p>
            <form onSubmit={submitEndorsement} style={{ display: "flex", flexDirection: "column", gap: "1rem", marginTop: "1rem" }}>
              <label>
                <div>Sum Insured (ETB)</div>
                <input 
                  type="number" 
                  className="input" 
                  value={sumInsured} 
                  onChange={(e) => setSumInsured(Number(e.target.value))} 
                  required 
                  min={1000}
                />
              </label>
              <div style={{ display: "flex", gap: "1rem" }}>
                <button type="submit" className="btn" disabled={loading}>Calculate & Endorse</button>
                <button type="button" className="btn btn-secondary" onClick={() => setShowEndorseModal(false)}>Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
