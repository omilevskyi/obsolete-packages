package main

import (
	"cmp"
	"path/filepath"
	"strconv"
	"strings"
)

// VersionType - structure that defines FreeBSD port "versions"
type VersionType struct {
	Path         string
	PortVersion  []string // delimited by "."
	PortRevision int      // _${PORTREVISION} 20250707: 1..102
	PortEpoch    int      // ,${PORTEPOCH} 20250707: 1,2,3,4,6,8
}

// intSuffix extracts an integer suffix from the input string `s`
// following the last occurrence of the separator `sep`. It returns the
// parsed integer (or 0 if parsing fails) and the trimmed string without the suffix.
// Example: intSuffix("pkg-1.2.3_4,5", ',') → (5, "pkg-1.2.3_4")
func intSuffix(s string, sep byte) (int, string) {
	v := 0
	if i := strings.LastIndexByte(s, sep); i >= 0 {
		if i+1 < len(s) {
			if num, err := strconv.Atoi(s[i+1:]); err == nil && num > 0 {
				v = num
			}
		}
		s = s[:i]
	}
	return v, s
}

// keyAndVersion parses a FreeBSD-style port version string into VersionType.
// Example input: "/path/some-pkg-name-1.2.3_4,5.pkg";
// returns: "some-pkg-name", {"/path/some-pkg-name-1.2.3_4,5.pkg", []{"1", "2", "3"}, 4, 5}
func keyAndVersion(path string) (string, *VersionType) {
	s := filepath.Base(path)
	// Remove file extension if exists (e.g., ".pkg")
	for i, isAllDigits := len(s)-1, true; 0 <= i && s[i] != ',' && s[i] != '_'; i-- {
		if s[i] == '.' {
			if !isAllDigits {
				// if i+1 < len(s) { ext = s[i+1:] }
				s = s[:i]
			}
			break
		}
		isAllDigits = isAllDigits && '0' <= s[i] && s[i] <= '9'
	}

	// Extract version part after last dash
	key := ""
	if dash := strings.LastIndexByte(s, '-'); dash >= 0 {
		if dash > 0 {
			key = s[:dash]
		}
		s = s[dash+1:]
	}
	if key == "" {
		return "", nil
	}

	// Extract epoch (after last comma) and revision (after last underscore)
	portEpoch, s := intSuffix(s, ',')
	portRevision, s := intSuffix(s, '_')

	// Count items
	count, slen := 1, len(s)
	if slen == 0 {
		return "", nil
	}
	for i := 0; i < slen; i++ {
		if s[i] == '.' {
			count++
		}
	}

	if p := filepath.Clean(path); p != path {
		path = p
	}

	// Split version into items
	portVersion := make([]string, count)
	count = 0
	for i, start := 0, 0; i <= slen; i++ {
		if i == slen || s[i] == '.' {
			portVersion[count] = s[start:i]
			start, count = i+1, count+1
		}
	}

	return key, &VersionType{
		Path:         path,
		PortVersion:  portVersion,
		PortRevision: portRevision,
		PortEpoch:    portEpoch,
	}
}

// versionsContain checks whether a VersionType slice contains an entry with the specified path.
// Returns true if a match is found, otherwise false.
func versionsContain(s []VersionType, path string) bool {
	for i := 0; i < len(s); i++ {
		if s[i].Path == path {
			return true
		}
	}
	return false
}

// compareVersionDesc compares two VersionType values in descending order.
// It prioritizes PortEpoch, then compares PortVersion segments numerically if possible,
// and finally compares PortRevision. Returns 1 if 'a' is older than 'b', -1 if newer, 0 if equal.
func compareVersionDesc(a, b VersionType) int {
	// Compare epoch
	if a.PortEpoch != b.PortEpoch {
		return cmp.Compare(b.PortEpoch, a.PortEpoch)
	}

	// Compare version segments
	for i := 0; i < max(len(a.PortVersion), len(b.PortVersion)); i++ {
		var va, vb string
		if i < len(a.PortVersion) {
			va = a.PortVersion[i]
		}
		if i < len(b.PortVersion) {
			vb = b.PortVersion[i]
		}

		if va != vb {
			if ia, err := strconv.Atoi(va); err == nil { // Try to compare as integers
				if ib, err := strconv.Atoi(vb); err == nil {
					if ia != ib {
						return cmp.Compare(ib, ia)
					}
					continue
				}
			}
			return cmp.Compare(vb, va)
		}
	}

	return cmp.Compare(b.PortRevision, a.PortRevision) // Compare revision
}
