package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	_ "github.com/griffnb/core/lib/config"
	"github.com/griffnb/techboss-ai-go/internal/environment"
)

func main() {
	schema := generateSchema(reflect.TypeOf(environment.Config{}))
	schema["$schema"] = "http://json-schema.org/draft-07/schema#"
	schema["title"] = "Server Configuration"
	schema["description"] = "Configuration schema"

	// Allow $schema property in config files
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		props["$schema"] = map[string]interface{}{
			"type":        "string",
			"description": "JSON Schema reference",
		}
	}

	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling schema: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

func generateSchema(t reflect.Type) map[string]interface{} {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := map[string]interface{}{
		"type":                 "object",
		"properties":           make(map[string]interface{}),
		"additionalProperties": false,
	}

	if t.Kind() != reflect.Struct {
		return schema
	}

	properties := schema["properties"].(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Handle embedded structs (like DefaultConfig)
		if field.Anonymous {
			embeddedSchema := generateSchema(field.Type)
			if embeddedProps, ok := embeddedSchema["properties"].(map[string]interface{}); ok {
				for k, v := range embeddedProps {
					properties[k] = v
				}
			}
			continue
		}

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		fieldSchema := generateFieldSchema(field.Type)
		properties[jsonTag] = fieldSchema
	}

	return schema
}

func generateFieldSchema(t reflect.Type) map[string]interface{} {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return map[string]interface{}{
			"type": "string",
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return map[string]interface{}{
			"type": "integer",
		}
	case reflect.Float32, reflect.Float64:
		return map[string]interface{}{
			"type": "number",
		}
	case reflect.Bool:
		return map[string]interface{}{
			"type": "boolean",
		}
	case reflect.Slice, reflect.Array:
		return map[string]interface{}{
			"type":  "array",
			"items": generateFieldSchema(t.Elem()),
		}
	case reflect.Map:
		valueSchema := generateFieldSchema(t.Elem())
		return map[string]interface{}{
			"type":                 "object",
			"additionalProperties": valueSchema,
		}
	case reflect.Struct:
		structSchema := generateSchema(t)
		// Keep additionalProperties restriction on nested objects
		return structSchema
	default:
		return map[string]interface{}{
			"type": "object",
		}
	}
}
