"use client";

import Link from "next/link";
import { useLocale } from "@/components/Shell";
import { t } from "@/lib/i18n";

export default function HomePage() {
  const { locale } = useLocale();
  const am = locale === "am";

  return (
    <div className="overflow-hidden">
      {/* ---- Hero ---------------------------------------------------------- */}
      <section className="mx-auto grid max-w-7xl grid-cols-1 items-center gap-12 px-6 pb-8 pt-16 lg:grid-cols-2 lg:pt-24">
        <div className="animate-rise flex flex-col items-start gap-6">
          <span className="eyebrow">
            <span className="h-1.5 w-1.5 rounded-full bg-gold" />
            {am ? "የኢትዮጵያ ኢንሹራንስ ኮርፖሬሽን" : "Ethiopian Insurance Corporation"}
          </span>
          <h1 className="text-5xl leading-[1.08] text-slate-900 lg:text-6xl">
            {am ? "የመኪና መድን፣" : "Motor insurance,"}
            <br />
            <span className="bg-gradient-to-r from-brand-600 to-brand-500 bg-clip-text text-transparent">
              {am ? "ከቅናሽ እስከ ይገባኛል።" : "quote to claim."}
            </span>
          </h1>
          <p className="max-w-lg text-lg leading-relaxed text-slate-600">{t("tagline", locale)}</p>
          <div className="mt-2 flex flex-wrap items-center gap-3">
            <Link className="btn btn-primary btn-lg" href="/quote">
              {t("ctaQuote", locale)}
              <svg className="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M13 7l5 5-5 5M6 12h12" /></svg>
            </Link>
            <Link className="btn btn-ghost btn-lg" href="/claim">{t("ctaClaim", locale)}</Link>
          </div>
          <div className="mt-4 flex flex-wrap items-center gap-x-6 gap-y-2 text-sm text-slate-500">
            <span className="inline-flex items-center gap-2"><Check /> {am ? "የቴሌብር ክፍያ" : "Telebirr payments"}</span>
            <span className="inline-flex items-center gap-2"><Check /> {am ? "ባለሁለት ቋንቋ ሰነዶች" : "Bilingual documents"}</span>
            <span className="inline-flex items-center gap-2"><Check /> {am ? "ፈጣን ይገባኛል" : "Fast-track claims"}</span>
          </div>
        </div>

        {/* Floating premium quote card */}
        <div className="animate-rise delay-200 relative">
          <div className="absolute -inset-4 -z-10 rounded-[2.5rem] bg-gradient-to-tr from-brand-100 via-white to-amber-50 blur-xl" />
          <div className="card card-pad animate-float">
            <div className="flex items-center justify-between">
              <div>
                <div className="text-xs font-semibold uppercase tracking-wider text-slate-400">{am ? "የዋጋ ቅናሽ" : "Motor quote"}</div>
                <div className="mt-1 font-mono text-sm font-bold text-brand-700">EIC/MOT/2026/000001</div>
              </div>
              <span className="badge badge-success"><span className="badge-dot" />ISSUED</span>
            </div>
            <div className="my-5 h-px bg-slate-100" />
            <dl className="space-y-3 text-sm">
              {[
                [am ? "የራስ ጉዳት (OD)" : "Own damage (OD)", "1,500.00"],
                [am ? "የሶስተኛ ወገን (TPL)" : "Third-party (TPL)", "800.00"],
                [am ? "ተ.እ.ታ 15%" : "VAT 15%", "345.00"],
                [am ? "የቴምብር ቀረጥ" : "Stamp duty", "35.00"],
              ].map(([k, v]) => (
                <div key={k} className="flex items-center justify-between">
                  <dt className="text-slate-500">{k}</dt>
                  <dd className="font-mono font-medium text-slate-700">{v}</dd>
                </div>
              ))}
              <div className="flex items-center justify-between border-t border-slate-100 pt-3">
                <dt className="font-semibold text-slate-900">{am ? "ጠቅላላ (ETB)" : "Gross premium (ETB)"}</dt>
                <dd className="font-mono text-lg font-bold text-brand-700">2,680.00</dd>
              </div>
            </dl>
          </div>
        </div>
      </section>

      {/* ---- Lifecycle ---------------------------------------------------- */}
      <section className="mx-auto max-w-7xl px-6 py-14">
        <div className="card card-pad">
          <div className="mb-6">
            <span className="eyebrow">{am ? "ሙሉ የሥራ ሂደት" : "End-to-end lifecycle"}</span>
          </div>
          <ol className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
            {[
              [am ? "ቅናሽ" : "Quote", "1"],
              [am ? "ማጣራት" : "Underwrite", "2"],
              [am ? "ማሰር" : "Bind", "3"],
              [am ? "ክፍያ" : "Pay", "4"],
              [am ? "ይገባኛል" : "FNOL", "5"],
              [am ? "ማጠናቀቅ" : "Settle", "6"],
            ].map(([label, n]) => (
              <li key={label} className="flex items-center gap-3">
                <span className="grid h-9 w-9 shrink-0 place-items-center rounded-xl bg-brand-600 text-sm font-bold text-white shadow-[var(--shadow-lift)]">{n}</span>
                <span className="text-sm font-semibold text-slate-700">{label}</span>
              </li>
            ))}
          </ol>
        </div>
      </section>

      {/* ---- Features ----------------------------------------------------- */}
      <section className="mx-auto max-w-7xl px-6 pb-16">
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <Feature
            title={am ? "ቀጥታ ማጣራት (STP)" : "Straight-through underwriting"}
            body={am ? "ደንብ-ተኮር ውሳኔ፦ በራስ-ሰር ተቀበል፣ ላክ ወይም ውድቅ አድርግ።" : "Rule-based decisions auto-accept, refer, or decline in milliseconds."}
            icon={<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>}
          />
          <Feature
            title={am ? "ትክክለኛ የገንዘብ ስሌት" : "Exact decimal money"}
            body={am ? "ተ.እ.ታ + የቴምብር ቀረጥ በባንክ ደረጃ ማጠጋጋት — ያለ ተንሳፋፊ ቁጥር።" : "VAT + stamp duty with banker's rounding — never floating-point money."}
            icon={<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path strokeLinecap="round" strokeLinejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8V6m0 12v-2m0-8c1.11 0 2.08.402 2.599 1M12 8V6" /></svg>}
          />
          <Feature
            title={am ? "የማይለወጥ ኦዲት" : "Immutable audit trail"}
            body={am ? "እያንዳንዱ የሁኔታ ለውጥ ይመዘገባል — ለተቆጣጣሪ ዝግጁ።" : "Every state change is recorded — regulator-ready from day one."}
            icon={<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" /></svg>}
          />
          <Feature
            title={am ? "የቴሌብር ክፍያ" : "Telebirr payments"}
            body={am ? "የተረጋገጠ የHMAC ማሳወቂያ — ደረሰኝ በራስ-ሰር ይዘጋል።" : "HMAC-verified callbacks settle invoices automatically and safely."}
            icon={<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path strokeLinecap="round" strokeLinejoin="round" d="M3 10h18M7 15h1m4 0h1m-7 4h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" /></svg>}
          />
          <Feature
            title={am ? "ባለሁለት ቋንቋ" : "Bilingual by design"}
            body={am ? "አማርኛ እና እንግሊዝኛ በሁሉም ስክሪንና ሰነድ ላይ።" : "Amharic and English across every screen, document, and QR sticker."}
            icon={<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path strokeLinecap="round" strokeLinejoin="round" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" /></svg>}
          />
          <Feature
            title={am ? "ቀጥታ KPI" : "Real-time KPIs"}
            body={am ? "የኪሳራ እና ጥምር ሬሾ ከክስተቶች በቀጥታ ይሰላል።" : "Loss & combined ratios projected live from the event stream."}
            icon={<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path strokeLinecap="round" strokeLinejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>}
          />
        </div>
      </section>
    </div>
  );
}

function Feature({ title, body, icon }: { title: string; body: string; icon: React.ReactNode }) {
  return (
    <div className="card-interactive card-pad">
      <div className="stat-icon mb-4 h-11 w-11 [&_svg]:h-5 [&_svg]:w-5">{icon}</div>
      <h3 className="text-lg font-bold text-slate-900">{title}</h3>
      <p className="mt-2 text-sm leading-relaxed text-slate-500">{body}</p>
    </div>
  );
}

function Check() {
  return (
    <svg className="h-4 w-4 text-success-500" fill="none" stroke="currentColor" strokeWidth="2.5" viewBox="0 0 24 24">
      <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
    </svg>
  );
}
