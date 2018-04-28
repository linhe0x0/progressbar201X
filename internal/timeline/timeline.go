// Package timeline get the position of the specified time point in a duration.
package timeline

import (
	"errors"
	"time"

	"github.com/apex/log"
	"github.com/sqrthree/toFixed"
)

// New get the position of the specified time point in a duration.
func New(t time.Time, d [2]time.Time) (float64, error) {
	point := t.Unix()
	start, end := d[0].Unix(), d[1].Unix()

	if point < start || point > end {
		return float64(-1), errors.New("out of bounds")
	}

	current := point - start
	total := end - start

	ratio := float64(current) / float64(total)

	ratio = toFixed.ToFixed(ratio, 4)

	return ratio, nil
}

func NewWithYear(t time.Time) (float64, error) {
	year := t.Year()

	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, time.UTC)

	d := [2]time.Time{start, end}

	ratio, err := New(t, d)

	if err != nil {
		return -1, err
	}

	log.Debugf("progress of %v is %v", year, ratio)

	return ratio, nil
}

func NewWithMonth(t time.Time) (float64, error) {
	year := t.Year()
	month := t.Month()

	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)

	d := [2]time.Time{start, end}

	ratio, err := New(t, d)

	if err != nil {
		return -1, err
	}

	log.Debugf("progress of %s is %v", month, ratio)

	return ratio, nil
}
