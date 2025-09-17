package helpers

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/CrowdShield/go-core/lib/sanitize"
)

type MockPublicJSONData struct {
	PrivateData map[string]interface{} `json:"private"`
	PublicData  map[string]interface{} `json:"public"`
}

func (m MockPublicJSONData) ToPublicJSON() any {
	return m.PublicData
}

func TestToPublicJSONDataResponseMap(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name: "Struct with fields implementing PublicJSONData",
			input: struct {
				Model MockPublicJSONData `json:"model"`
				Raw   string             `json:"raw"`
			}{
				Model: MockPublicJSONData{
					PrivateData: map[string]interface{}{"privateKey": "privateValue"},
					PublicData:  map[string]interface{}{"key": "value"},
				},
				Raw: "unchanged",
			},
			expected: map[string]interface{}{
				"model": map[string]interface{}{"key": "value"},
				"raw":   "unchanged",
			},
		},
		{
			name: "Single model implementing PublicJSONData",
			input: MockPublicJSONData{
				PrivateData: map[string]interface{}{"privateKey": "privateValue"},
				PublicData:  map[string]interface{}{"key": "value"},
			},
			expected: map[string]interface{}{"key": "value"},
		},
		{
			name: "Slice of models implementing PublicJSONData",
			input: []MockPublicJSONData{
				{
					PrivateData: map[string]interface{}{"privateKey1": "privateValue1"},
					PublicData:  map[string]interface{}{"key1": "value1"},
				},
				{
					PrivateData: map[string]interface{}{"privateKey2": "privateValue2"},
					PublicData:  map[string]interface{}{"key2": "value2"},
				},
			},
			expected: []interface{}{
				map[string]interface{}{"key1": "value1"},
				map[string]interface{}{"key2": "value2"},
			},
		},
		{
			name: "Map with models implementing PublicJSONData",
			input: map[string]interface{}{
				"model1": MockPublicJSONData{
					PrivateData: map[string]interface{}{"privateKey": "privateValue"},
					PublicData:  map[string]interface{}{"key": "value"},
				},
				"raw": "unchanged",
			},
			expected: map[string]interface{}{
				"model1": map[string]interface{}{"key": "value"},
				"raw":    "unchanged",
			},
		},
		{
			name: "Nested complex structure",
			input: []interface{}{
				MockPublicJSONData{
					PrivateData: map[string]interface{}{"privateKey1": "privateValue1"},
					PublicData:  map[string]interface{}{"key1": "value1"},
				},
				map[string]interface{}{
					"model": MockPublicJSONData{
						PrivateData: map[string]interface{}{"privateKey2": "privateValue2"},
						PublicData:  map[string]interface{}{"key2": "value2"},
					},
				},
			},
			expected: []interface{}{
				map[string]interface{}{"key1": "value1"},
				map[string]interface{}{
					"model": map[string]interface{}{"key2": "value2"},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := sanitize.PublicSanitizeResponse(test.input)
			if !reflect.DeepEqual(result, test.expected) {
				t.Errorf("expected %+v, got %+v", test.expected, result)
			}
		})
	}
}

func TestGenerateKey(_ *testing.T) {
	fmt.Println(GenerateKey(32))
}
