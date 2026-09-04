package document_enum

type DocType string

const (
	Markdown DocType = ".md"
	Text     DocType = ".txt"
)
