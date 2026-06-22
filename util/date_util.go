package util

import (
	"errors"
	"strings"
	"time"
)

// ParseDateStringToTimeStamp parses a date string and returns a pointer to a time.Time object or an error.
// Returns nil if the input string is empty.
func ParseDateStringToTimeStamp(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	return &parsedDate, nil
}

func NormalizeDate(dateStr string) (time.Time, error) {
	if strings.TrimSpace(dateStr) == "" || strings.ToUpper(dateStr) == "N/A" {
		return time.Time{}, nil
	}

	formats := []string{
		"2 Jan 2006",
		"2006/01/02",
		"2006-01-02",
		"02-01-2006",
		"01-02-2006",
		"2006.01.02",
		"January 2, 2006",
		"2006-01-02T15:04:05Z",
	}
	var parsedTime time.Time
	var err error

	for _, format := range formats {
		parsedTime, err = time.Parse(format, strings.TrimSpace(dateStr))
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, errors.New("invalid date format")
}
