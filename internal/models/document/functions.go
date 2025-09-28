package document

import (
	"context"
	"fmt"
	"strings"

	"github.com/griffnb/techboss-ai-go/internal/environment"
)

func (this *Document) GetFilePath(field string) string {
	fullURL := this.GetString(field)
	fullURL = strings.Replace(fullURL, "https://", "", -1)
	urlParts := strings.Split(fullURL, "/")
	return strings.Join(urlParts[1:], "/")
}

func (this *Document) GetExtension(field string) string {
	filePath := this.GetFilePath(field)
	parts := strings.Split(filePath, ".")
	return parts[len(parts)-1]
}

func (this *Document) GenerateFileName(field string) string {
	return fmt.Sprintf("%s.%s", this.GetString("key"), this.GetExtension(field))
}

func BuildS3URL(documentGroupKey string, documentKey, extension string) string {
	return fmt.Sprintf("documents/%s/%s.%s", documentGroupKey, documentKey, extension)
}

func (this *Document) GetS3Data(field string) ([]byte, error) {
	return environment.GetS3().DownloadToBuffer(context.Background(), environment.GetConfig().S3Config.Buckets["documents"], this.GetFilePath(field))
}
