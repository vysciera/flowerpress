package turso

import "time"

const timestampLayout = "2006-01-02 15:04:05"

func parseTimestamp(value string) (time.Time, error) {
	return time.Parse(timestampLayout, value)
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(timestampLayout)
}

func formatNullableTimestamp(value *time.Time) any {
	if value == nil {
		return nil
	}

	return formatTimestamp(*value)
}
