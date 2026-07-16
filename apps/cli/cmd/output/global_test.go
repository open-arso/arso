package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintJSON(t *testing.T) {
	tests := []struct {
		name          string
		value         any
		expectedJSON  string
		expectError   bool
	}{
		{
			name: "simple struct",
			value: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			expectedJSON: `{
  "name": "John",
  "age": 30
}
`,
			expectError: false,
		},
		{
			name: "map",
			value: map[string]interface{}{
				"key":   "value",
				"count": 42,
				"active": true,
			},
			expectedJSON: `{
  "active": true,
  "count": 42,
  "key": "value"
}
`,
			expectError: false,
		},
		{
			name: "slice",
			value: []string{"apple", "banana", "cherry"},
			expectedJSON: `[
  "apple",
  "banana",
  "cherry"
]
`,
			expectError: false,
		},
		{
			name: "nested struct",
			value: struct {
				ID     int `json:"id"`
				Person struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				} `json:"person"`
			}{
				ID: 1,
				Person: struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				}{
					Name:  "Alice",
					Email: "alice@example.com",
				},
			},
			expectedJSON: `{
  "id": 1,
  "person": {
    "name": "Alice",
    "email": "alice@example.com"
  }
}
`,
			expectError: false,
		},
		{
			name: "pointer to struct",
			value: &struct {
				Value string `json:"value"`
			}{
				Value: "test",
			},
			expectedJSON: `{
  "value": "test"
}
`,
			expectError: false,
		},
		{
			name: "nil value",
			value: nil,
			expectedJSON: `null
`,
			expectError: false,
		},
		{
			name: "channel - should fail",
			value: make(chan int),
			expectError: true,
		},
		{
			name: "complex number - should fail",
			value: complex(1, 2),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			err := PrintJSON(cmd, tt.value)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedJSON, buf.String())

			// Verify the output is valid JSON
			var result interface{}
			err = json.Unmarshal(buf.Bytes(), &result)
			assert.NoError(t, err, "Output should be valid JSON")
		})
	}
}

func TestPrintNDJSON(t *testing.T) {
	tests := []struct {
		name          string
		values        interface{} // Use interface{} to handle different types
		expectedLines []string
		expectError   bool
	}{
		{
			name: "simple structs",
			values: []struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				{Name: "John", Age: 30},
				{Name: "Jane", Age: 25},
				{Name: "Bob", Age: 35},
			},
			expectedLines: []string{
				`{"name":"John","age":30}`,
				`{"name":"Jane","age":25}`,
				`{"name":"Bob","age":35}`,
			},
			expectError: false,
		},
		{
			name: "strings",
			values: []string{"apple", "banana", "cherry"},
			expectedLines: []string{
				`"apple"`,
				`"banana"`,
				`"cherry"`,
			},
			expectError: false,
		},
		{
			name: "integers",
			values: []int{1, 2, 3, 4, 5},
			expectedLines: []string{
				"1",
				"2",
				"3",
				"4",
				"5",
			},
			expectError: false,
		},
		{
			name: "maps",
			values: []map[string]interface{}{
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
			},
			expectedLines: []string{
				`{"id":1,"name":"Alice"}`,
				`{"id":2,"name":"Bob"}`,
			},
			expectError: false,
		},
		{
			name: "nested structs",
			values: []struct {
				ID     int `json:"id"`
				Person struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				} `json:"person"`
			}{
				{
					ID: 1,
					Person: struct {
						Name  string `json:"name"`
						Email string `json:"email"`
					}{
						Name:  "Alice",
						Email: "alice@example.com",
					},
				},
				{
					ID: 2,
					Person: struct {
						Name  string `json:"name"`
						Email string `json:"email"`
					}{
						Name:  "Bob",
						Email: "bob@example.com",
					},
				},
			},
			expectedLines: []string{
				`{"id":1,"person":{"name":"Alice","email":"alice@example.com"}}`,
				`{"id":2,"person":{"name":"Bob","email":"bob@example.com"}}`,
			},
			expectError: false,
		},
		{
			name: "empty slice",
			values: []string{},
			expectedLines: []string{},
			expectError: false,
		},
		{
			name: "nil slice",
			values: []string(nil),
			expectedLines: []string{},
			expectError: false,
		},
		{
			name: "single element",
			values: []int{42},
			expectedLines: []string{"42"},
			expectError: false,
		},
		{
			name: "custom type",
			values: []struct {
				ID      int     `json:"id"`
				Price   float64 `json:"price"`
				InStock bool    `json:"in_stock"`
			}{
				{ID: 1, Price: 19.99, InStock: true},
				{ID: 2, Price: 29.99, InStock: false},
			},
			expectedLines: []string{
				`{"id":1,"price":19.99,"in_stock":true}`,
				`{"id":2,"price":29.99,"in_stock":false}`,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			// Use type assertion to call PrintNDJSON with the correct type
			var err error
			switch v := tt.values.(type) {
			case []struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}:
				err = PrintNDJSON(cmd, v)
			case []string:
				err = PrintNDJSON(cmd, v)
			case []int:
				err = PrintNDJSON(cmd, v)
			case []map[string]interface{}:
				err = PrintNDJSON(cmd, v)
			case []struct {
				ID     int `json:"id"`
				Person struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				} `json:"person"`
			}:
				err = PrintNDJSON(cmd, v)
			case []struct {
				ID      int     `json:"id"`
				Price   float64 `json:"price"`
				InStock bool    `json:"in_stock"`
			}:
				err = PrintNDJSON(cmd, v)
			default:
				t.Fatalf("Unsupported type in test: %T", tt.values)
			}

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Split output into lines
			lines := bytes.Split(buf.Bytes(), []byte("\n"))

			// Remove empty last line if present
			if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
				lines = lines[:len(lines)-1]
			}

			// Check number of lines
			assert.Equal(t, len(tt.expectedLines), len(lines), "Number of lines doesn't match")

			// Check each line
			for i, line := range lines {
				assert.Equal(t, tt.expectedLines[i], string(line), "Line %d doesn't match", i)
			}

			// Verify each line is valid JSON
			for i, line := range lines {
				var result interface{}
				err := json.Unmarshal(line, &result)
				assert.NoError(t, err, "Line %d is not valid JSON: %s", i, string(line))
			}
		})
	}
}

func TestPrintNDJSON_ErrorHandling(t *testing.T) {
	// Test with values that can't be marshaled
	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	// Create a slice with an unmarshalable type (channel)
	values := []struct {
		Ch chan int `json:"ch"`
	}{
		{Ch: make(chan int)},
	}

	err := PrintNDJSON(cmd, values)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "json: unsupported type")
}

func TestPrintJSON_WithCustomMarshaler(t *testing.T) {
	type CustomTime struct {
		Time string `json:"time"`
	}

	type CustomStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	value := CustomStruct{
		ID:   1,
		Name: "test",
	}

	err := PrintJSON(cmd, value)
	assert.NoError(t, err)

	expected := `{
  "id": 1,
  "name": "test"
}
`
	assert.Equal(t, expected, buf.String())
}

func TestPrintNDJSON_MixedTypes(t *testing.T) {
	// Test with a slice of interfaces
	type Item struct {
		Type  string      `json:"type"`
		Value interface{} `json:"value"`
	}

	values := []Item{
		{Type: "string", Value: "hello"},
		{Type: "number", Value: 42},
		{Type: "boolean", Value: true},
		{Type: "object", Value: map[string]string{"key": "value"}},
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := PrintNDJSON(cmd, values)
	assert.NoError(t, err)

	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	// Remove empty last line
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	assert.Equal(t, 4, len(lines))

	// Verify each line is valid JSON and contains expected fields
	for _, line := range lines {
		var result map[string]interface{}
		err := json.Unmarshal(line, &result)
		assert.NoError(t, err)
		assert.Contains(t, result, "type")
		assert.Contains(t, result, "value")
	}
}

func TestPrintJSON_Formatting(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		contains []string
	}{
		{
			name: "indentation",
			value: struct {
				Field1 string `json:"field1"`
				Field2 int    `json:"field2"`
				Nested  struct {
					Inner string `json:"inner"`
				} `json:"nested"`
			}{
				Field1: "value1",
				Field2: 42,
				Nested: struct {
					Inner string `json:"inner"`
				}{
					Inner: "inner value",
				},
			},
			contains: []string{
				"{",
				"  \"field1\": \"value1\",",
				"  \"field2\": 42,",
				"  \"nested\": {",
				"    \"inner\": \"inner value\"",
				"  }",
				"}",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			err := PrintJSON(cmd, tt.value)
			assert.NoError(t, err)

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

// Benchmark tests
func BenchmarkPrintJSON(b *testing.B) {
	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	data := struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = PrintJSON(cmd, data)
	}
}

func BenchmarkPrintNDJSON(b *testing.B) {
	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	data := []struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}{
		{ID: 1, Name: "John", Email: "john@example.com"},
		{ID: 2, Name: "Jane", Email: "jane@example.com"},
		{ID: 3, Name: "Bob", Email: "bob@example.com"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		_ = PrintNDJSON(cmd, data)
	}
}

// Helper function to verify JSON structure
func verifyJSONStructure(t *testing.T, jsonStr string, expectedKeys ...string) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	require.NoError(t, err, "Failed to parse JSON: %s", jsonStr)

	for _, key := range expectedKeys {
		assert.Contains(t, result, key, "Missing expected key: %s", key)
	}
}

func TestPrintJSON_WithOmitEmpty(t *testing.T) {
	type TestStruct struct {
		ID      int    `json:"id"`
		Name    string `json:"name,omitempty"`
		Email   string `json:"email,omitempty"`
		Age     *int   `json:"age,omitempty"`
		Address string `json:"address"`
	}

	tests := []struct {
		name     string
		value    TestStruct
		expected string
	}{
		{
			name: "all fields populated",
			value: TestStruct{
				ID:      1,
				Name:    "John",
				Email:   "john@example.com",
				Age:     intPtr(30),
				Address: "123 Main St",
			},
			expected: `{
  "id": 1,
  "name": "John",
  "email": "john@example.com",
  "age": 30,
  "address": "123 Main St"
}
`,
		},
		{
			name: "with omitted fields",
			value: TestStruct{
				ID:      1,
				Address: "123 Main St",
			},
			expected: `{
  "id": 1,
  "address": "123 Main St"
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			err := PrintJSON(cmd, tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func intPtr(i int) *int {
	return &i
}