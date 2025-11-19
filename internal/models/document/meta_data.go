package document

type MetaData struct {
	WebpageURL     string   `json:"webpage_url,omitempty"`
	FileOpenAIID   string   `json:"file_openai_id,omitempty"`
	VectorOpenAIID string   `json:"vector_openai_id,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}
