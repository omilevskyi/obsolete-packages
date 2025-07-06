package main

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestIntSuffix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		sep      byte
		wantInt  int
		wantRest string
	}{
		{
			name:     "Standard comma suffix",
			input:    "pkg-1.2.3_4,5",
			sep:      ',',
			wantInt:  5,
			wantRest: "pkg-1.2.3_4",
		},
		{
			name:     "Standard underscore suffix",
			input:    "pkg-1.2.3_4",
			sep:      '_',
			wantInt:  4,
			wantRest: "pkg-1.2.3",
		},
		{
			name:     "No separator present",
			input:    "pkg-1.2.3",
			sep:      ',',
			wantInt:  0,
			wantRest: "pkg-1.2.3",
		},
		{
			name:     "Separator at end with no number",
			input:    "pkg-1.2.3_,",
			sep:      ',',
			wantInt:  0,
			wantRest: "pkg-1.2.3_",
		},
		{
			name:     "Non-numeric suffix",
			input:    "pkg-1.2.3_abc",
			sep:      '_',
			wantInt:  0,
			wantRest: "pkg-1.2.3",
		},
		{
			name:     "Multiple separators, only last matters",
			input:    "pkg-1.2.3_4_5",
			sep:      '_',
			wantInt:  5,
			wantRest: "pkg-1.2.3_4",
		},
		{
			name:     "Separator at beginning",
			input:    "_5",
			sep:      '_',
			wantInt:  5,
			wantRest: "",
		},
		{
			name:     "Empty string",
			input:    "",
			sep:      ',',
			wantInt:  0,
			wantRest: "",
		},
		{
			name:     "Only separator",
			input:    ",",
			sep:      ',',
			wantInt:  0,
			wantRest: "",
		},
		{
			name:     "Separator with negative number (invalid)",
			input:    "pkg-1.2.3_-5",
			sep:      '_',
			wantInt:  0, // strconv.Atoi("-5") is valid, but if you want to reject negatives, this would need logic
			wantRest: "pkg-1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInt, gotRest := intSuffix(tt.input, tt.sep)
			if gotInt != tt.wantInt || gotRest != tt.wantRest {
				t.Errorf("intSuffix(%q, %q) = (%d, %q); want (%d, %q)",
					tt.input, tt.sep, gotInt, gotRest, tt.wantInt, tt.wantRest)
			}
		})
	}
}

func TestKeyAndVersion(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedKey string
		expected    *VersionType
	}{
		{
			name:        "Standard .pkg with revision and epoch",
			input:       "/usr/ports/pkg-name-1.2.3_4,5.pkg",
			expectedKey: "pkg-name",
			expected: &VersionType{
				Path:         "/usr/ports/pkg-name-1.2.3_4,5.pkg",
				PortVersion:  []string{"1", "2", "3"},
				PortRevision: 4,
				PortEpoch:    5,
			},
		},
		{
			name:        "No revision, only epoch",
			input:       "pkg-1.2.3_,7.txz",
			expectedKey: "pkg",
			expected: &VersionType{
				Path:         "pkg-1.2.3_,7.txz",
				PortVersion:  []string{"1", "2", "3"},
				PortRevision: 0,
				PortEpoch:    7,
			},
		},
		{
			name:        "No epoch, only revision",
			input:       "pkg-1.2.3_4,.txz",
			expectedKey: "pkg",
			expected: &VersionType{
				Path:         "pkg-1.2.3_4,.txz",
				PortVersion:  []string{"1", "2", "3"},
				PortRevision: 4,
				PortEpoch:    0,
			},
		},
		{
			name:        "No revision and epoch",
			input:       "pkg-1.2.3.txz",
			expectedKey: "pkg",
			expected: &VersionType{
				Path:         "pkg-1.2.3.txz",
				PortVersion:  []string{"1", "2", "3"},
				PortRevision: 0,
				PortEpoch:    0,
			},
		},
		{
			name:        "Long version string",
			input:       "pkg-10.20.30.40_99,88.pkg",
			expectedKey: "pkg",
			expected: &VersionType{
				Path:         "pkg-10.20.30.40_99,88.pkg",
				PortVersion:  []string{"10", "20", "30", "40"},
				PortRevision: 99,
				PortEpoch:    88,
			},
		},
		{
			name:        "Single digit version",
			input:       "pkg-1_1,1.pkg",
			expectedKey: "pkg",
			expected: &VersionType{
				Path:         "pkg-1_1,1.pkg",
				PortVersion:  []string{"1"},
				PortRevision: 1,
				PortEpoch:    1,
			},
		},
		{
			name:        "No dash (no key)",
			input:       "1.2.3_4,5.pkg",
			expectedKey: "",
			expected:    nil,
		},
		{
			name:        "Empty string",
			input:       "",
			expectedKey: "",
			expected:    nil,
		},
		{
			name:        "Only dash",
			input:       "-1.2.3_4,5.pkg",
			expectedKey: "",
			expected:    nil,
		},
		{
			name:        "Malformed version string",
			input:       "pkg-1.2a.3_4,5.pkg",
			expectedKey: "pkg",
			expected: &VersionType{
				Path:         "pkg-1.2a.3_4,5.pkg",
				PortVersion:  []string{"1", "2a", "3"},
				PortRevision: 4,
				PortEpoch:    5,
			},
		},
		{
			name:        "No extension",
			input:       "pkg-1.2.3_4,5",
			expectedKey: "pkg",
			expected: &VersionType{
				Path:         "pkg-1.2.3_4,5",
				PortVersion:  []string{"1", "2", "3"},
				PortRevision: 4,
				PortEpoch:    5,
			},
		},
		{
			name:        "Path with redundant elements",
			input:       "/usr/../usr/ports/pkg-1.2.3_4,5.pkg",
			expectedKey: "pkg",
			expected: &VersionType{
				Path:         filepath.Clean("/usr/../usr/ports/pkg-1.2.3_4,5.pkg"),
				PortVersion:  []string{"1", "2", "3"},
				PortRevision: 4,
				PortEpoch:    5,
			},
		},
		{
			name:        "Dot in extension but all digits",
			input:       "pkg-1.2.3.123",
			expectedKey: "pkg",
			expected: &VersionType{
				Path:         "pkg-1.2.3.123",
				PortVersion:  []string{"1", "2", "3", "123"},
				PortRevision: 0,
				PortEpoch:    0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, result := KeyAndVersion(tt.input)

			if key != tt.expectedKey {
				t.Errorf("KeyAndVersion(%q) key = %q; want %q", tt.input, key, tt.expectedKey)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("KeyAndVersion(%q) = %+v; want %+v", tt.input, result, tt.expected)
			}
		})
	}
}
