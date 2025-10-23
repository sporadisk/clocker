package summary

import (
	"fmt"
	"strings"
	"time"
)

const (
	Uncategorized = "uncategorized"
)

type Summary struct {
	Valid         bool
	ValidationMsg string
	TimeWorked    time.Duration
	TimeLeft      *time.Duration
	Surplus       *time.Duration
	FullDayAt     *time.Time
	Categories    []ResultCategory
	Date          *Date
	Warnings      []string
}

type Date struct {
	DayName string
	Day     int
	Month   int
	Year    int
}

func (sd *Date) String() string {
	return fmt.Sprintf("%04d-%02d-%02d", sd.Year, sd.Month, sd.Day)
}

type ResultCategory struct {
	Name       string
	TimeWorked time.Duration
}

func (cr *Summary) AddCategory(name string, dur time.Duration) {
	for i, c := range cr.Categories {
		if c.MatchName(name) {
			cr.Categories[i].TimeWorked += dur
			return
		}
	}

	cr.Categories = append(cr.Categories, ResultCategory{
		Name:       name,
		TimeWorked: dur,
	})
}

func (rc *ResultCategory) MatchName(name string) bool {
	return strings.EqualFold(name, rc.Name)
}

type Output interface {
	OutputSummary(summary Summary) error
}
