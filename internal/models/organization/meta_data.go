package organization

import (
	"encoding/json"
	"fmt"
)

type MetaData struct {
	VectorStoreIDs map[string]string `public:"view" json:"vector_store_id,omitempty"`
	OnboardAnswers map[string]any    `public:"view" json:"onboard_answers,omitempty"`
}

func (this *MetaData) GetOnboardAnswersString() string {
	if this.OnboardAnswers == nil {
		return ""
	}

	b, err := json.MarshalIndent(this.OnboardAnswers, "", "  ")
	if err != nil {
		return fmt.Sprintf("error marshaling onboard answers: %v", err)
	}
	return string(b)
}
