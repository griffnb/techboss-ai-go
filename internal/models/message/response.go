package message

// ToJSONDoc Returns JSONDOC map
func (this *Message) ToJSONDoc() map[string]interface{} {
	jsonDoc := map[string]interface{}{
		"id":         this.Key,
		"type":       "message",
		"attributes": this,
	}

	return jsonDoc
}
