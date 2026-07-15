// Package pdf generates bilingual EIC motor policy documents with QR codes.
package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
	"time"

	"github.com/InnoSure-Platform/pc-platform/internal/store"
	"github.com/InnoSure-Platform/pc-shared-go/calendar"
	"github.com/InnoSure-Platform/pc-shared-go/i18n"
)

// Generator renders schedule, COI, and windshield sticker PDFs.
type Generator struct {
	OutputDir string
	BaseURL   string // public URL prefix for /files/
}

func NewGenerator(outputDir, baseURL string) *Generator {
	if outputDir == "" {
		outputDir = "./data/docs"
	}
	if baseURL == "" {
		baseURL = "/files"
	}
	_ = os.MkdirAll(outputDir, 0o755)
	return &Generator{OutputDir: outputDir, BaseURL: baseURL}
}

// Pack generates all Phase 0 issuance documents for a policy.
func (g *Generator) Pack(pol *store.Policy, party *store.Party) ([]store.Document, error) {
	issued := time.Now().UTC()
	if pol.IssuedAt != nil {
		issued = *pol.IssuedAt
	}
	eth := calendar.FromGregorian(issued)
	nameEN, nameAM := insuredNames(party)
	var docs []store.Document
	specs := []struct {
		docType, locale, title string
		am                     bool
		sticker                bool
	}{
		{"schedule", "en", i18n.T("doc.schedule", i18n.EN), false, false},
		{"schedule", "am", i18n.T("doc.schedule", i18n.AM), true, false},
		{"coi", "en", i18n.T("doc.coi", i18n.EN), false, false},
		{"coi", "am", i18n.T("doc.coi", i18n.AM), true, false},
		{"sticker", "en", i18n.T("doc.sticker", i18n.EN), false, true},
	}
	for _, sp := range specs {
		d, err := g.render(pol, nameEN, nameAM, eth, sp.docType, sp.locale, sp.title, sp.am, sp.sticker)
		if err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, nil
}

func (g *Generator) render(pol *store.Policy, nameEN, nameAM string, eth calendar.EthiopianDate, docType, locale, title string, am, sticker bool) (store.Document, error) {
	id := storeNewUUID()
	safeNum := sanitize(pol.PolicyNumber)
	fname := fmt.Sprintf("%s-%s-%s.pdf", safeNum, docType, locale)
	path := filepath.Join(g.OutputDir, fname)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(18, 18, 18)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	pdf.Cell(0, 10, "ETHIOPIAN INSURANCE CORPORATION")
	pdf.Ln(8)
	pdf.SetFont("Helvetica", "", 11)
	if am {
		pdf.Cell(0, 6, "የኢትዮጵያ ኢንሹራንስ ኮርፖሬሽን · መድህን")
	} else {
		pdf.Cell(0, 6, "Medhen Platform · Motor Insurance")
	}
	pdf.Ln(10)
	pdf.SetFont("Helvetica", "B", 14)
	pdf.Cell(0, 8, title)
	pdf.Ln(10)
	pdf.SetFont("Helvetica", "", 11)
	lines := []string{
		fmt.Sprintf("Policy: %s", pol.PolicyNumber),
		fmt.Sprintf("Insured: %s", ternary(am, nameAM, nameEN)),
		fmt.Sprintf("Vehicle: %s %s %d · %s", pol.Risk.Make, pol.Risk.Model, pol.Risk.Year, pol.Risk.PlateNumber),
		fmt.Sprintf("Cover: %s", pol.Risk.CoverType),
		fmt.Sprintf("Sum insured: %.2f ETB", float64(pol.Risk.SumInsuredMinor)/100),
		fmt.Sprintf("Premium: %.2f ETB", float64(pol.TotalMinor)/100),
		fmt.Sprintf("Period: %s → %s", pol.EffectiveFrom, pol.EffectiveTo),
		fmt.Sprintf("Ethiopian date: %s", ternary(am, eth.FormatAM(), eth.FormatEN())),
	}
	for _, ln := range lines {
		pdf.Cell(0, 6, ln)
		pdf.Ln(6)
	}
	qrPayload := fmt.Sprintf("medhen://policy/%s", pol.ID)
	png, err := qrcode.Encode(qrPayload, qrcode.Medium, 128)
	if err == nil {
		opt := gofpdf.ImageOptions{ImageType: "PNG"}
		name := fmt.Sprintf("qr-%s", id)
		pdf.RegisterImageOptionsReader(name, opt, bytes.NewReader(png))
		pdf.Ln(4)
		if sticker {
			pdf.SetFont("Helvetica", "B", 12)
			pdf.Cell(0, 8, "NBE Windshield QR Sticker")
			pdf.Ln(8)
		}
		pdf.ImageOptions(name, 18, pdf.GetY(), 32, 32, false, opt, 0, "")
		pdf.SetXY(52, pdf.GetY())
		pdf.SetFont("Helvetica", "", 9)
		pdf.MultiCell(0, 5, qrPayload, "", "", false)
	}
	if err := pdf.OutputFileAndClose(path); err != nil {
		return store.Document{}, err
	}
	return store.Document{
		ID: id, PolicyID: pol.ID, Type: docType, Locale: locale,
		URL: g.BaseURL + "/" + fname, ObjectKey: "medhen-docs/" + fname,
	}, nil
}

func insuredNames(p *store.Party) (en, am string) {
	if p == nil {
		return "", ""
	}
	en = p.FullName
	am = p.FullNameAm
	if am == "" {
		am = en
	}
	return en, am
}

func sanitize(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			out = append(out, c)
		} else if c == '/' {
			out = append(out, '-')
		}
	}
	return string(out)
}

func ternary(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func storeNewUUID() string {
	return newUUID()
}

var newUUID = func() string { return "" }
