package product

// Product is a configuration-driven insurance product (BC-MDH-02).
type Product struct {
	Code       string   `json:"code"`
	Name       string   `json:"name"`
	NameAm     string   `json:"nameAm"`
	LOB        string   `json:"lob"`
	Version    string   `json:"version"`
	CoverTypes []string `json:"coverTypes"`
}

func SeedMotor() Product {
	return Product{
		Code:       "MOTOR-PRIVATE-COMP",
		Name:       "EIC Motor Private Comprehensive",
		NameAm:     "የኢንሹራንስ የግል መኪና ሁለንተናዊ",
		LOB:        "motor",
		Version:    "2026.1",
		CoverTypes: []string{"comprehensive", "third_party"},
	}
}

func RiskSchema() map[string]any {
	return map[string]any{
		"productCode": "MOTOR-PRIVATE-COMP",
		"lob":         "motor",
		"note":        "Shared-core extension: only this risk schema changes for a new LOB variant — core services stay the same.",
		"fields": []map[string]any{
			{"name": "plateNumber", "type": "string", "required": true, "label": "Plate number", "labelAm": "የሰሌዳ ቁጥር"},
			{"name": "chassisNumber", "type": "string", "required": false, "label": "Chassis", "labelAm": "ቻሲ"},
			{"name": "make", "type": "string", "required": true},
			{"name": "model", "type": "string", "required": true},
			{"name": "year", "type": "integer", "required": true, "min": 1990, "max": 2027},
			{"name": "usage", "type": "enum", "values": []string{"private", "commercial"}},
			{"name": "coverType", "type": "enum", "values": []string{"comprehensive", "third_party"}},
			{"name": "sumInsuredMinor", "type": "integer", "required": true, "currency": "ETB"},
		},
	}
}
