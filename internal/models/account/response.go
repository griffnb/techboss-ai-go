package account

import "github.com/griffnb/core/lib/sanitize"

// ToPublicJSON converts the model to a sanitized JSON representation for public consumption
func (this *Account) ToPublicJSON() any {
	return sanitize.SanitizeModel(this, &Structure{})
}

func (this *Account) ToJSONDoc() map[string]interface{} {
	jsonSafeData := this.GetDataCopy()
	delete(jsonSafeData, "hashed_password")
	return jsonSafeData
}
