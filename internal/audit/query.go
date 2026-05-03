package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"time"
)

// QueryFilter filters audit events. Zero value matches all.
type QueryFilter struct {
	Severity string    // "info", "notice", "warning", "alert" — empty = all
	Category string    // empty = all
	Since    time.Time // zero = all
	Limit    int       // 0 = default 200
}

// QueryJSONL scans the JSONL audit log (canonical, plaintext) and returns events matching f.
// Events are returned newest-first (file is scanned in reverse via buffered pass).
func QueryJSONL(path string, f QueryFilter) ([]*AuditEvent, error) {
	if f.Limit <= 0 {
		f.Limit = 200
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	// Forward scan into a buffer, then reverse — log files are small enough.
	var lines [][]byte
	sc := bufio.NewScanner(file)
	sc.Buffer(make([]byte, 1<<20), 1<<20) // 1MB per line max
	for sc.Scan() {
		line := make([]byte, len(sc.Bytes()))
		copy(line, sc.Bytes())
		lines = append(lines, line)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	var out []*AuditEvent
	for i := len(lines) - 1; i >= 0 && len(out) < f.Limit; i-- {
		var ev AuditEvent
		if err := json.Unmarshal(lines[i], &ev); err != nil {
			continue
		}
		if f.Severity != "" && string(ev.Severity) != f.Severity {
			continue
		}
		if f.Category != "" && ev.Category != f.Category {
			continue
		}
		if !f.Since.IsZero() && ev.Timestamp.Before(f.Since) {
			continue
		}
		out = append(out, &ev)
	}
	return out, nil
}
