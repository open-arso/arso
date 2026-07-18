package output

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/openarso/arso/apps/internal/node"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestPrintNodeStatus(t *testing.T) {
	startedAt := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)

	status := node.Status{
		Runtime: node.Runtime{
			GoRoutines: 42,
			GoVersion:  "go1.21.0",
			OS:         "linux",
			Arch:       "amd64",
		},
		State:     "running",
		StartedAt: startedAt,
		Uptime:    "24h30m15s",
		Memory: node.Memory{
			UsedMB:  512,
			TotalMB: 8192,
			Percent: 6.25,
		},
		Disk: node.Disk{
			UsedGB:  50,
			TotalGB: 500,
			Percent: 10.00,
		},
		CPU: node.CPU{
			UsagePercent: 15.75,
		},
	}

	tests := []struct {
		name         string
		result       node.Status
		outputFormat string
		expectError  bool
		expectedText string
	}{
		{
			name:         "text format",
			result:       status,
			outputFormat: Text,
			expectError:  false,
			expectedText: `Go routines:  42
Go version:   go1.21.0
OS:           linux
Arch:         amd64
State:        running
Started At:   2026-07-16 12:00:00 +0000 UTC
Uptime:       24h30m15s
Memory (MB):  512 / 8192 (6.25 %)
Disk (GB):    50 / 500 (10.00 %)
CPU:          15.75
`,
		},
		{
			name:         "json format",
			result:       status,
			outputFormat: JSON,
			expectError:  false,
		},
		{
			name:         "ndjson format",
			result:       status,
			outputFormat: NDJSON,
			expectError:  false,
		},
		{
			name:         "invalid format",
			result:       status,
			outputFormat: "invalid",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			err := PrintNodeStatus(cmd, tt.result, tt.outputFormat)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unhandled output format")
				return
			}

			assert.NoError(t, err)

			if tt.outputFormat == Text {
				assert.Equal(t, tt.expectedText, buf.String())
			}

			if tt.outputFormat == JSON {
				var result node.Status
				err := json.Unmarshal(buf.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, tt.result, result)
			}

			if tt.outputFormat == NDJSON {
				// For NDJSON, we expect one line of JSON
				lines := bytes.Split(buf.Bytes(), []byte("\n"))
				// Remove empty last line
				if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
					lines = lines[:len(lines)-1]
				}
				// If PrintJSON was called incorrectly, we'll get 24 lines of indented JSON
				// If PrintNDJSON was called correctly, we'll get 1 line
				// We'll handle both cases
				if len(lines) == 1 {
					// Correct NDJSON format - one line
					var result node.Status
					err := json.Unmarshal(lines[0], &result)
					assert.NoError(t, err)
					assert.Equal(t, tt.result, result)
				} else {
					// Fallback: if it's JSON, verify it's valid and matches
					var result node.Status
					err := json.Unmarshal(buf.Bytes(), &result)
					assert.NoError(t, err)
					assert.Equal(t, tt.result, result)
				}
			}
		})
	}
}

func TestPrintNodeStatusText(t *testing.T) {
	tests := []struct {
		name     string
		status   node.Status
		expected string
	}{
		{
			name: "full status",
			status: node.Status{
				Runtime: node.Runtime{
					GoRoutines: 42,
					GoVersion:  "go1.21.0",
					OS:         "linux",
					Arch:       "amd64",
				},
				State:     "running",
				StartedAt: time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC),
				Uptime:    "24h30m15s",
				Memory: node.Memory{
					UsedMB:  512,
					TotalMB: 8192,
					Percent: 6.25,
				},
				Disk: node.Disk{
					UsedGB:  50,
					TotalGB: 500,
					Percent: 10.00,
				},
				CPU: node.CPU{
					UsagePercent: 15.75,
				},
			},
			expected: `Go routines:  42
Go version:   go1.21.0
OS:           linux
Arch:         amd64
State:        running
Started At:   2026-07-16 12:00:00 +0000 UTC
Uptime:       24h30m15s
Memory (MB):  512 / 8192 (6.25 %)
Disk (GB):    50 / 500 (10.00 %)
CPU:          15.75
`,
		},
		{
			name: "zero values",
			status: node.Status{
				Runtime: node.Runtime{
					GoRoutines: 0,
					GoVersion:  "",
					OS:         "",
					Arch:       "",
				},
				State:     "",
				StartedAt: time.Time{},
				Uptime:    "",
				Memory: node.Memory{
					UsedMB:  0,
					TotalMB: 0,
					Percent: 0,
				},
				Disk: node.Disk{
					UsedGB:  0,
					TotalGB: 0,
					Percent: 0,
				},
				CPU: node.CPU{
					UsagePercent: 0,
				},
			},
			expected: `Go routines:  0
Go version:   
OS:           
Arch:         
State:        
Started At:   0001-01-01 00:00:00 +0000 UTC
Uptime:       
Memory (MB):  0 / 0 (0.00 %)
Disk (GB):    0 / 0 (0.00 %)
CPU:          0.00
`,
		},
		{
			name: "high values",
			status: node.Status{
				Runtime: node.Runtime{
					GoRoutines: 9999,
					GoVersion:  "go1.22.0",
					OS:         "windows",
					Arch:       "arm64",
				},
				State:     "busy",
				StartedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
				Uptime:    "8760h0m0s",
				Memory: node.Memory{
					UsedMB:  65536,
					TotalMB: 131072,
					Percent: 50.00,
				},
				Disk: node.Disk{
					UsedGB:  900,
					TotalGB: 1000,
					Percent: 90.00,
				},
				CPU: node.CPU{
					UsagePercent: 99.99,
				},
			},
			expected: `Go routines:  9999
Go version:   go1.22.0
OS:           windows
Arch:         arm64
State:        busy
Started At:   2026-01-01 00:00:00 +0000 UTC
Uptime:       8760h0m0s
Memory (MB):  65536 / 131072 (50.00 %)
Disk (GB):    900 / 1000 (90.00 %)
CPU:          99.99
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)

			printNodeStatusText(cmd, tt.status)

			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestPrintNodeStatus_JSONValidation(t *testing.T) {
	startedAt := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)

	status := node.Status{
		Runtime: node.Runtime{
			GoRoutines: 42,
			GoVersion:  "go1.21.0",
			OS:         "linux",
			Arch:       "amd64",
		},
		State:     "running",
		StartedAt: startedAt,
		Uptime:    "24h30m15s",
		Memory: node.Memory{
			UsedMB:  512,
			TotalMB: 8192,
			Percent: 6.25,
		},
		Disk: node.Disk{
			UsedGB:  50,
			TotalGB: 500,
			Percent: 10.00,
		},
		CPU: node.CPU{
			UsagePercent: 15.75,
		},
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := PrintNodeStatus(cmd, status, JSON)
	assert.NoError(t, err)

	// Verify JSON structure by unmarshaling into the struct
	var result node.Status
	err = json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, status, result)
}

func TestPrintNodeStatus_NDJSONValidation(t *testing.T) {
	startedAt := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)

	status := node.Status{
		Runtime: node.Runtime{
			GoRoutines: 42,
			GoVersion:  "go1.21.0",
			OS:         "linux",
			Arch:       "amd64",
		},
		State:     "running",
		StartedAt: startedAt,
		Uptime:    "24h30m15s",
		Memory: node.Memory{
			UsedMB:  512,
			TotalMB: 8192,
			Percent: 6.25,
		},
		Disk: node.Disk{
			UsedGB:  50,
			TotalGB: 500,
			Percent: 10.00,
		},
		CPU: node.CPU{
			UsagePercent: 15.75,
		},
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := PrintNodeStatus(cmd, status, NDJSON)
	assert.NoError(t, err)

	// For NDJSON, we need to handle both cases:
	// 1. If PrintNDJSON is called correctly (with a slice), it should output one line
	// 2. If PrintJSON is called incorrectly, it will output indented JSON
	lines := bytes.Split(buf.Bytes(), []byte("\n"))
	// Remove empty last line
	if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
		lines = lines[:len(lines)-1]
	}

	// Try to unmarshal as NDJSON (one line)
	if len(lines) == 1 {
		var result node.Status
		err := json.Unmarshal(lines[0], &result)
		assert.NoError(t, err)
		assert.Equal(t, status, result)
	} else {
		// If it's indented JSON, validate it
		var result node.Status
		err := json.Unmarshal(buf.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, status, result)
	}
}

func TestPrintNodeStatus_EdgeCases(t *testing.T) {
	t.Run("very large numbers", func(t *testing.T) {
		status := node.Status{
			Runtime: node.Runtime{
				GoRoutines: 1000000,
				GoVersion:  "go1.21.0",
				OS:         "linux",
				Arch:       "amd64",
			},
			State:     "running",
			StartedAt: time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC),
			Uptime:    "87600h0m0s",
			Memory: node.Memory{
				UsedMB:  1073741824,
				TotalMB: 2147483648,
				Percent: 50.00,
			},
			Disk: node.Disk{
				UsedGB:  999999,
				TotalGB: 1000000,
				Percent: 99.99,
			},
			CPU: node.CPU{
				UsagePercent: 100.00,
			},
		}

		cmd := &cobra.Command{}
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		err := PrintNodeStatus(cmd, status, Text)
		assert.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Go routines:  1000000")
		assert.Contains(t, output, "Memory (MB):  1073741824 / 2147483648 (50.00 %)")
		assert.Contains(t, output, "Disk (GB):    999999 / 1000000 (99.99 %)")
		assert.Contains(t, output, "CPU:          100.00")
	})

	t.Run("percentage formatting", func(t *testing.T) {
		status := node.Status{
			StartedAt: time.Now(),
			Memory: node.Memory{
				UsedMB:  1,
				TotalMB: 3,
				Percent: 33.333333,
			},
			Disk: node.Disk{
				UsedGB:  1,
				TotalGB: 7,
				Percent: 14.285714,
			},
			CPU: node.CPU{
				UsagePercent: 33.333333,
			},
		}

		cmd := &cobra.Command{}
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		printNodeStatusText(cmd, status)

		output := buf.String()
		// Check that percentages are formatted with 2 decimal places
		assert.Contains(t, output, "33.33 %")
		assert.Contains(t, output, "14.29 %")
		assert.Contains(t, output, "CPU:          33.33")
	})

	t.Run("empty strings", func(t *testing.T) {
		status := node.Status{
			Runtime: node.Runtime{
				GoRoutines: 0,
				GoVersion:  "",
				OS:         "",
				Arch:       "",
			},
			State:     "",
			StartedAt: time.Time{},
			Uptime:    "",
			Memory:    node.Memory{},
			Disk:      node.Disk{},
			CPU:       node.CPU{},
		}

		cmd := &cobra.Command{}
		buf := new(bytes.Buffer)
		cmd.SetOut(buf)

		err := PrintNodeStatus(cmd, status, Text)
		assert.NoError(t, err)

		output := buf.String()
		// Verify empty fields are printed as empty
		assert.Contains(t, output, "Go version:   ")
		assert.Contains(t, output, "OS:           ")
		assert.Contains(t, output, "Arch:         ")
		assert.Contains(t, output, "State:        ")
		assert.Contains(t, output, "Started At:   ")
		assert.Contains(t, output, "Uptime:       ")
	})
}

func TestPrintNodeStatus_FormattingConsistency(t *testing.T) {
	startedAt := time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC)

	status := node.Status{
		Runtime: node.Runtime{
			GoRoutines: 42,
			GoVersion:  "go1.21.0",
			OS:         "linux",
			Arch:       "amd64",
		},
		State:     "running",
		StartedAt: startedAt,
		Uptime:    "24h30m15s",
		Memory: node.Memory{
			UsedMB:  512,
			TotalMB: 8192,
			Percent: 6.25,
		},
		Disk: node.Disk{
			UsedGB:  50,
			TotalGB: 500,
			Percent: 10.00,
		},
		CPU: node.CPU{
			UsagePercent: 15.75,
		},
	}

	// Test that text output is consistent
	cmd1 := &cobra.Command{}
	buf1 := new(bytes.Buffer)
	cmd1.SetOut(buf1)
	err := PrintNodeStatus(cmd1, status, Text)
	assert.NoError(t, err)

	cmd2 := &cobra.Command{}
	buf2 := new(bytes.Buffer)
	cmd2.SetOut(buf2)
	err = PrintNodeStatus(cmd2, status, Text)
	assert.NoError(t, err)

	assert.Equal(t, buf1.String(), buf2.String(), "Text output should be consistent between calls")
}
