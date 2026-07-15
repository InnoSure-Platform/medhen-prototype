// Package i18n provides bilingual (en/am) message catalogs for Phase 0.
package i18n

type Locale string

const (
	EN Locale = "en"
	AM Locale = "am"
)

func ParseLocale(s string) Locale {
	if s == "am" || s == "am-ET" {
		return AM
	}
	return EN
}

var messages = map[string]map[Locale]string{
	"premium.base": {
		EN: "Base premium",
		AM: "መሰረታዊ ፕሪሚየም",
	},
	"premium.factor.age": {
		EN: "Vehicle age factor",
		AM: "የተሽከርካሪ ዕድሜ ክፍያ",
	},
	"premium.factor.usage": {
		EN: "Usage factor",
		AM: "የአጠቃቀም ክፍያ",
	},
	"premium.vat": {
		EN: "VAT (15%)",
		AM: "ተ.አ.ታ (15%)",
	},
	"premium.stamp": {
		EN: "Stamp duty",
		AM: "የማህተም ቀረጥ",
	},
	"policy.issued": {
		EN: "Your EIC motor policy has been issued.",
		AM: "የኢንሹራንስ ፖሊሲዎ ተሰጥቷል።",
	},
	"claim.settled": {
		EN: "Your claim has been settled. Funds will arrive via Telebirr.",
		AM: "የይገባኛል ጥያቄዎ ተፈትሏል። ክፍያ በቴሌብር ይደርሳል።",
	},
	"doc.schedule": {
		EN: "Policy Schedule",
		AM: "የፖሊሲ መርሃ ግብር",
	},
	"doc.coi": {
		EN: "Certificate of Insurance",
		AM: "የኢንሹራንስ የምስክር ወረቀት",
	},
	"doc.sticker": {
		EN: "Windshield QR Sticker",
		AM: "የመኪና መስታወት QR ስቲከር",
	},
}

func T(key string, loc Locale) string {
	if m, ok := messages[key]; ok {
		if v, ok := m[loc]; ok {
			return v
		}
		return m[EN]
	}
	return key
}
