package presentation

import (
	"encoding/json"
	"net/http"
	"github.com/inno-sphere/medhen-prototype/services/pc-document-mgmt-svc/internal/application"
)

type DocumentRestHandler struct {
	uploadUseCase *application.UploadDocumentUseCase
}

func NewDocumentRestHandler(uploadUseCase *application.UploadDocumentUseCase) *DocumentRestHandler {
	return &DocumentRestHandler{uploadUseCase: uploadUseCase}
}

func (h *DocumentRestHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit upload size to 100MB
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "Bad Request: File too large or invalid multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		http.Error(w, "Bad Request: Missing 'document' file part", http.StatusBadRequest)
		return
	}
	defer file.Close()

	cmd := application.UploadDocumentCommand{
		TenantID:     r.FormValue("tenant_id"),
		DocumentType: r.FormValue("document_type"),
		EntityType:   r.FormValue("entity_type"),
		EntityID:     r.FormValue("entity_id"),
		MimeType:     header.Header.Get("Content-Type"),
		FileSize:     header.Size,
		Stream:       file, // Stream directly from multipart reader
	}

	docID, err := h.uploadUseCase.Execute(r.Context(), cmd)
	if err != nil {
		http.Error(w, "Internal Server Error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"document_id": docID,
		"status":      "UPLOADED",
	})
}
