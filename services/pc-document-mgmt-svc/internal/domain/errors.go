package domain

import "errors"

var (
	ErrTemplateNotFound     = errors.New("template not found")
	ErrSchemaMismatch       = errors.New("payload does not match template schema")
	ErrMalwareDetected      = errors.New("malware detected during ICAP scan")
	ErrVirusScanPending     = errors.New("virus scan pending, action not allowed")
	ErrInvalidState         = errors.New("invalid state transition")
	ErrDocumentNotFound     = errors.New("document not found")
	ErrSignatureNotPending  = errors.New("signature request is not pending")
	ErrLegalHoldActive      = errors.New("document is under legal hold and cannot be modified or archived")
)
