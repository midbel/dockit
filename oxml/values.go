package oxml

import (
	"time"
)

var supportedDateFormats = []string{
	"2006-01-02",
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05-07:00",
}

func ParseDate(str string) (time.Time, error) {
	var (
		when time.Time
		err  error
	)
	for _, f := range supportedDateFormats {
		when, err = time.Parse(f, str)
		if err == nil {
			break
		}
	}
	return when, err
}
