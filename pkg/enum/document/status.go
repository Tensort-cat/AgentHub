package document_enum

type DocStatus int8

const (
	Pending DocStatus = iota + 1
	Parsing
	Indexing
	Success
	Error
)
