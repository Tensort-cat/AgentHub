package document_enum

type DocStatus int8

const (
	Uploading DocStatus = iota + 1
	Success
	Error
)
