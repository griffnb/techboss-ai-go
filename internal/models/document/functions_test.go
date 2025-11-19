package document

import (
	"testing"
)

func TestDocument_GetFilePath(t *testing.T) {
	obj := New()
	obj.RawS3URL.Set("https://dev-bb-pub-assets.s3.us-east-1.amazonaws.com/stbl/core.webp")

	path := obj.GetFilePath("raw_s3_url")
	if path != "stbl/core.webp" {
		t.Errorf("Expected stbl/core.webp, got %s", path)
	}

}
