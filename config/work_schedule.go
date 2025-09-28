package config

import (
	"time"
)

// WorkScheduleConfig defines the working hours and days for the organization
type WorkScheduleConfig struct {
	WorkingDays  []time.Weekday `json:"working_days"`
	StartTime    string         `json:"start_time"`    // "09:00"
	EndTime      string         `json:"end_time"`      // "16:00"
	DailyHours   float64        `json:"daily_hours"`   // 7 hours per day
	WeeklyHours  float64        `json:"weekly_hours"`  // 42 hours per week
	TimeZone     string         `json:"timezone"`      // "Asia/Riyadh"
	LunchBreak   float64        `json:"lunch_break"`   // 1 hour lunch break
	WorkingHours float64        `json:"working_hours"` // 6 effective hours per day
}

// GetDefaultWorkSchedule returns the Arabic work schedule configuration
func GetDefaultWorkSchedule() *WorkScheduleConfig {
	return &WorkScheduleConfig{
		WorkingDays: []time.Weekday{
			time.Saturday,  // السبت
			time.Sunday,    // الأحد
			time.Monday,    // الاثنين
			time.Tuesday,   // الثلاثاء
			time.Wednesday, // الأربعاء
			time.Thursday,  // الخميس
		},
		StartTime:    "09:00",
		EndTime:      "16:00",
		DailyHours:   7.0,  // 9 AM to 4 PM = 7 hours
		LunchBreak:   1.0,  // 1 hour lunch break
		WorkingHours: 6.0,  // 6 effective working hours (7 - 1 lunch break)
		WeeklyHours:  42.0, // 7 hours × 6 days = 42 hours per week
		TimeZone:     "Asia/Riyadh",
	}
}

// IsWorkingDay checks if a given date is a working day
func (w *WorkScheduleConfig) IsWorkingDay(date time.Time) bool {
	// Convert to local timezone
	loc, _ := time.LoadLocation(w.TimeZone)
	localDate := date.In(loc)

	weekday := localDate.Weekday()
	for _, workDay := range w.WorkingDays {
		if weekday == workDay {
			return true
		}
	}
	return false
}

// IsWorkingHour checks if a given time is within working hours
func (w *WorkScheduleConfig) IsWorkingHour(datetime time.Time) bool {
	if !w.IsWorkingDay(datetime) {
		return false
	}

	// Convert to local timezone
	loc, _ := time.LoadLocation(w.TimeZone)
	localTime := datetime.In(loc)

	hour := localTime.Hour()
	minute := localTime.Minute()
	currentMinutes := hour*60 + minute

	// Parse start and end times
	startHour, startMin := 9, 0 // 09:00
	endHour, endMin := 16, 0    // 16:00

	startMinutes := startHour*60 + startMin
	endMinutes := endHour*60 + endMin

	return currentMinutes >= startMinutes && currentMinutes < endMinutes
}

// GetWorkingHoursInPeriod calculates effective working hours in a time period
func (w *WorkScheduleConfig) GetWorkingHoursInPeriod(start, end time.Time) float64 {
	if start.After(end) {
		return 0
	}

	totalHours := 0.0
	current := start

	// Count working days and hours
	for current.Before(end) || current.Equal(end) {
		if w.IsWorkingDay(current) {
			// Calculate overlap with working hours for this day
			dayStart := time.Date(current.Year(), current.Month(), current.Day(), 9, 0, 0, 0, current.Location())
			dayEnd := time.Date(current.Year(), current.Month(), current.Day(), 16, 0, 0, 0, current.Location())

			// Find overlap between [start, end] and [dayStart, dayEnd]
			overlapStart := maxTime(start, dayStart)
			overlapEnd := minTime(end, dayEnd)

			if overlapStart.Before(overlapEnd) {
				duration := overlapEnd.Sub(overlapStart)
				hours := duration.Hours()

				// Subtract lunch break if the overlap includes lunch time (12:00-13:00)
				lunchStart := time.Date(current.Year(), current.Month(), current.Day(), 12, 0, 0, 0, current.Location())
				lunchEnd := time.Date(current.Year(), current.Month(), current.Day(), 13, 0, 0, 0, current.Location())

				if overlapStart.Before(lunchEnd) && overlapEnd.After(lunchStart) {
					lunchOverlapStart := maxTime(overlapStart, lunchStart)
					lunchOverlapEnd := minTime(overlapEnd, lunchEnd)
					lunchDuration := lunchOverlapEnd.Sub(lunchOverlapStart)
					hours -= lunchDuration.Hours()
				}

				totalHours += hours
			}
		}
		current = current.AddDate(0, 0, 1) // Move to next day
	}

	return totalHours
}

// GetNextWorkingDay returns the next working day after the given date
func (w *WorkScheduleConfig) GetNextWorkingDay(date time.Time) time.Time {
	next := date.AddDate(0, 0, 1)
	for !w.IsWorkingDay(next) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

// GetWorkingDaysUntil counts working days between two dates
func (w *WorkScheduleConfig) GetWorkingDaysUntil(start, end time.Time) int {
	if start.After(end) {
		return 0
	}

	count := 0
	current := start

	for current.Before(end) || current.Equal(end) {
		if w.IsWorkingDay(current) {
			count++
		}
		current = current.AddDate(0, 0, 1)
	}

	return count
}

// CalculateDeadlineRisk assesses deadline risk based on working hours remaining
func (w *WorkScheduleConfig) CalculateDeadlineRisk(deadline time.Time, estimatedHours float64) string {
	now := time.Now()

	if deadline.Before(now) {
		return "critical" // Already overdue
	}

	// Calculate working hours until deadline
	hoursUntilDeadline := w.GetWorkingHoursInPeriod(now, deadline)

	// Risk assessment based on ratio of estimated hours to available hours
	ratio := estimatedHours / hoursUntilDeadline

	if ratio >= 1.0 {
		return "critical" // Need more time than available
	} else if ratio >= 0.8 {
		return "high" // Need 80%+ of available time
	} else if ratio >= 0.5 {
		return "medium" // Need 50-80% of available time
	} else {
		return "low" // Need less than 50% of available time
	}
}

// GetOptimalTaskSchedule suggests optimal start time for a task
func (w *WorkScheduleConfig) GetOptimalTaskSchedule(estimatedHours float64, deadline time.Time) time.Time {
	now := time.Now()

	// If we're in working hours, suggest starting now
	if w.IsWorkingHour(now) {
		return now
	}

	// Otherwise, suggest starting at the beginning of the next working day
	nextWorkDay := w.GetNextWorkingDay(now)
	startOfWorkDay := time.Date(nextWorkDay.Year(), nextWorkDay.Month(), nextWorkDay.Day(), 9, 0, 0, 0, nextWorkDay.Location())

	return startOfWorkDay
}

// Helper functions
func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
