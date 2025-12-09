package backup

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ParseSchedule converts schedule string to systemd OnCalendar format
func ParseSchedule(schedule string) (string, error) {
	schedule = strings.TrimSpace(strings.ToLower(schedule))

	// Common schedule patterns
	schedules := map[string]string{
		"hourly":           "*-*-* *:00:00",
		"daily":            "*-*-* 02:00:00",
		"daily 02:00":      "*-*-* 02:00:00",
		"weekly":           "Mon *-*-* 02:00:00",
		"weekly sun 03:00": "Sun *-*-* 03:00:00",
		"monthly":          "*-*-01 02:00:00",
	}

	// Check if it's a predefined pattern
	if cal, ok := schedules[schedule]; ok {
		return cal, nil
	}

	// Parse "daily HH:MM" format
	if after, ok := strings.CutPrefix(schedule, "daily "); ok {
		time := after
		return fmt.Sprintf("*-*-* %s:00", time), nil
	}

	// Parse "weekly DAY HH:MM" format
	if after, ok := strings.CutPrefix(schedule, "weekly "); ok {
		parts := strings.Fields(after)
		if len(parts) == 2 {
			caser := cases.Title(language.English)
			day := caser.String(parts[0])
			return fmt.Sprintf("%s *-*-* %s:00", day, parts[1]), nil
		}
	}

	// If it looks like a systemd calendar format, use it directly
	if strings.Contains(schedule, "*") || strings.Contains(schedule, ":") {
		return schedule, nil
	}

	return "", fmt.Errorf("invalid schedule format: %s", schedule)
}
