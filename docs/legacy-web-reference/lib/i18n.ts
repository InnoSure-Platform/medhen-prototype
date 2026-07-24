export type Locale = "en" | "am";

const dict = {
  brand: { en: "Ethiopian Insurance Corporation", am: "የኢትዮጵያ ኢንሹራንስ ኮርፖሬሽን" },
  product: { en: "Medhen", am: "መድህን" },
  tenantTag: { en: "EIC Motor · tenant eic", am: "የኢንሹራንስ መኪና · tenant eic" },
  tagline: {
    en: "Motor cover from quote to claim — bilingual, Telebirr-ready, built for EIC.",
    am: "ከቅናሽ እስከ ይገባኛል — ባለሁለት ቋንቋ፣ ቴሌብር፣ ለኢንሹራንስ የተሰራ።",
  },
  ctaQuote: { en: "Start a motor quote", am: "የመኪና ቅናሽ ጀምር" },
  ctaClaim: { en: "File a claim", am: "ይገባኛል አቅርብ" },
  navQuote: { en: "Quote", am: "ቅናሽ" },
  navClaim: { en: "Claim", am: "ይገባኛል" },
  navStaff: { en: "Staff", am: "ሰራተኛ" },
  heroSide: {
    en: "One shared core. Motor today — Life and Property through configuration.",
    am: "አንድ የጋራ መድረክ። ዛሬ መኪና — ነገ ሕይወትና ንብረት በውቅር።",
  },
  loginTitle: { en: "Sign in to Medhen", am: "ወደ መድህን ግባ" },
  loginSub: { en: "EIC agent portal — Keycloak secured", am: "የኢንሹራንስ ወኪል ፖርታል — Keycloak" },
  loginUser: { en: "Username", am: "የተጠቃሚ ስም" },
  loginPass: { en: "Password", am: "የይለፍ ቃል" },
  loginBtn: { en: "Sign in", am: "ግባ" },
  loginHint: { en: "Demo: demo-agent / medhen-demo", am: "Demo: demo-agent / medhen-demo" },
  loginNav: { en: "Sign in", am: "ግባ" },
  logout: { en: "Sign out", am: "ውጣ" },
  quoteTitle: { en: "Motor quote", am: "የመኪና ቅናሽ" },
  quoteContinue: { en: "Continue", am: "ቀጥል" },
  quoteCalculate: { en: "Calculate premium", am: "ዋጋ አስላ" },
  quotePay: { en: "Pay with Telebirr & issue", am: "በቴሌብር ክፈልና ፖሊሲ አውጣ" },
  quoteIssued: { en: "Policy issued", am: "ፖሊሲ ተሰጥቷል" },
  quoteDocs: { en: "Policy documents", am: "የፖሊሲ ሰነዶች" },
  quoteDownload: { en: "Download PDF", am: "PDF አውርድ" },
  quoteDownloadAll: { en: "Download all documents", am: "ሁሉንም ሰነዶች አውርድ" },
  quoteStp: { en: "STP underwriting", am: "አוטማቲክ UW" },
  quoteTotal: { en: "Total", am: "ጠቅላላ" },
  quoteLine: { en: "Line", am: "ዝርዝር" },
  quoteAmount: { en: "Amount", am: "መጠን" },
  claimTitle: { en: "First notice of loss", am: "የይገባኛል ማሳወቂያ" },
  claimSub: {
    en: "Photo keys, GPS, and estimate — routes to fast-track when under threshold.",
    am: "ፎቶ፣ GPS እና የተገመተ ኪሳራ — በቀላል መስመር ይሄዳል።",
  },
  claimPolicyId: { en: "Policy ID", am: "የፖሊሲ መለያ" },
  claimDescription: { en: "Description", am: "መግለጫ" },
  claimAmount: { en: "Estimated loss (ETB)", am: "የተገመተ ኪሳራ (ETB)" },
  claimSubmit: { en: "Submit FNOL", am: "FNOL አቅርብ" },
  claimSettle: { en: "Settle fast-track", am: "ፈጣን ክፍያ ፍቀድ" },
  claimSettled: { en: "Settled", am: "ተፈትሏል" },
  fullName: { en: "Full name", am: "ሙሉ ስም" },
  fullNameAm: { en: "Name (Amharic)", am: "ስም (አማርኛ)" },
  phone: { en: "Phone (Telebirr)", am: "ስልክ (ቴሌብር)" },
  plate: { en: "Plate number", am: "የሰሌዳ ቁጥር" },
  make: { en: "Make", am: "ማምረት" },
  model: { en: "Model", am: "ሞዴል" },
  year: { en: "Year", am: "ዓ.ም." },
  cover: { en: "Cover type", am: "የሽፋን አይነት" },
  sumInsured: { en: "Sum insured (ETB)", am: "የተገመተ ዋጋ (ETB)" },
  coverComprehensive: { en: "Comprehensive", am: "ሁለንተናዊ" },
  coverThirdParty: { en: "Third party", am: "ሦስተኛ ወገን" },
} as const;

export function t(key: keyof typeof dict, locale: Locale): string {
  return dict[key][locale];
}

// formatETB formats an amount given in minor units (santim), e.g. KPI figures.
export function formatETB(minor: number, locale: Locale): string {
  return formatBirr((minor ?? 0) / 100, locale);
}

// formatBirr formats an amount already in major units (Birr). The monolith's
// domain money marshals as a major-unit JSON number (e.g. 2160.00).
export function formatBirr(birr: number, locale: Locale): string {
  const v = (birr ?? 0).toLocaleString(locale === "am" ? "am-ET" : "en-ET", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
  return `${v} ETB`;
}
