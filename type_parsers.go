package config

import (
	"fmt"
	"net/url"
	"time"
)

///////////////////////////////////////////////////////////////////////////////
// Time, supports the following formats

const TimeFormat20 = "2006-01-02T15:04:05Z"
const TimeFormat19 = "2006-01-02T15:04:05"
const TimeFormat10 = "2006-01-02"

func selectTimeFormatFor(s string) (string, error) {
	if l := len(s); l == 20 {
		return TimeFormat20, nil
	} else if l == 19 {
		return TimeFormat19, nil
	} else if l == 10 {
		return TimeFormat10, nil
	} else {
		return "", fmt.Errorf("unknown time format, expected %s or %s or %s",
			TimeFormat20, TimeFormat19, TimeFormat10)
	}
}

func ParseTime(s string) (time.Time, error) {
	format, err := selectTimeFormatFor(s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(format, s)
}

///////////////////////////////////////////////////////////////////////////////
// URL support

func ParseUrl(s string) (*url.URL, error) {
	return url.Parse(s)
}
