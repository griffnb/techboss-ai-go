package document

type DocumentType int

const (
	DOCUMENT_TYPE_WEBPAGE_SUMMARY DocumentType = iota + 1
	DOCUMENT_TYPE_PDF
	DOCUMENT_TYPE_TEXT
	DOCUMENT_TYPE_SITEMAP
)
