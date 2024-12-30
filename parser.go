package loggeradapter

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const (
	regexpDataDuration = `(?i)^(\d+)(y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second)$`
	regexpDataFileSize = `(?i)^(\d+)(b|byte|kb|kilobyte|mb|megabyte|gb|gigabyte|tb|terabyte)$` // |pb|petabyte|zb|zettabyte

	regexpDuration = `(?i)^(y|year|M|month|mo|mon|w|week|d|day|h|hour|m|minute|min|s|second)$`
	regexpFileSize = `(?i)^(b|byte|kb|kiloByte|mb|megabyte|gb|gigabyte|tb|terabyte)$` // |pb|petabyte|zb|zettabyte

	regexpFixedInterval = `(?i)^(annually|monthly|weekly|daily|hourly|minutely|secondly)$`

	regexpYear    = `(?i)^(y|year)$`
	regexpMonth   = `(?i)^(month|mo|mon)$`
	regexpMonthM  = `^(M)$`
	regexpWeek    = `(?i)^(w|week)$`
	regexpDay     = `(?i)^(d|day)$`
	regexpHour    = `(?i)^(h|hour)$`
	regexpMinute  = `(?i)^(minute|min)$`
	regexpMinuteM = `^(m)$`
	regexpSecond  = `(?i)^(s|second)$`

	regexpByte = `(?i)^(b|byte)$`
	regexpKB   = `(?i)^(kb|kilobyte)$`
	regexpMB   = `(?i)^(mb|megabyte)$`
	regexpGB   = `(?i)^(gb|gigabyte)$`
	regexpTB   = `(?i)^(tb|terabyte)$`

	regexpFixedAnnually = `(?i)^(annually)$`
	regexpFixedMonthly  = `(?i)^(monthly)$`
	regexpFixedWeekly   = `(?i)^(weekly)$`
	regexpFixedDaily    = `(?i)^(daily)$`
	regexpFixedHourly   = `(?i)^(hourly)$`
	regexpFixedMinutely = `(?i)^(minutely)$`
	regexpFixedSecondly = `(?i)^(secondly)$`
)

// ParseExpression match and parse expression
//
// rotation: [y/M/w/d/h/m/s] / [b/kb/mb/gb/tb] / [annually|monthly|weekly|daily|hourly|minutely|secondly]
//
// retain/archive: [y/M/w/d/h/m/s] / [<number>]
func ParseExpression(expression string) (int, string, error) {
	if IsFixedInterval(expression) {
		return parseFixedInterval(expression)
	}

	matches, err := parseExpression(regexpDataDuration, expression)
	if err == nil {
		return parseMatches(matches)
	}

	matches, err = parseExpression(regexpDataFileSize, expression)
	if err == nil {
		return parseMatches(matches)
	}

	var v int
	v, err = strconv.Atoi(expression)
	if err != nil {
		return 0, "", fmt.Errorf("invalid expression: %s", expression)
	}

	return v, "", nil
}

func parseExpression(regexpPattern string, expression string) ([]string, error) {
	regex, err := regexp.Compile(regexpPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regexp: %s", regexpPattern)
	}

	matches := regex.FindStringSubmatch(expression)
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid expression: %s", expression)
	}

	return matches, nil
}

func parseMatches(matches []string) (int, string, error) {
	if len(matches) != 3 {
		return 0, "", errors.New("not enough matches")
	}

	vs := matches[1]
	unit := matches[2]

	v, _err := strconv.Atoi(vs)
	if _err != nil {
		return 0, "", fmt.Errorf("invalid expression value: %s", vs)
	}

	return v, unit, nil
}

func parseFixedInterval(expression string) (int, string, error) {
	if IsFixedAnnually(expression) {
		return 1, "y", nil
	}
	if IsFixedMonthly(expression) {
		return 1, "M", nil
	}
	if IsFixedWeekly(expression) {
		return 1, "w", nil
	}
	if IsFixedDaily(expression) {
		return 1, "d", nil
	}
	if IsFixedHourly(expression) {
		return 1, "h", nil
	}
	if IsFixedMinutely(expression) {
		return 1, "m", nil
	}
	if IsFixedSecondly(expression) {
		return 1, "s", nil
	}
	return 0, "", fmt.Errorf("invalid expression: %s", expression)
}

func Match(pattern string, s string) bool {
	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

func IsFixedInterval(unit string) bool {
	return Match(regexpFixedInterval, unit)
}

func IsDuration(unit string) bool {
	return Match(regexpDuration, unit)
}

func IsFileSize(unit string) bool {
	return Match(regexpFileSize, unit)
}

func IsYear(unit string) bool {
	return Match(regexpYear, unit)
}

func IsMonth(unit string) bool {
	return Match(regexpMonth, unit) || Match(regexpMonthM, unit)
}

func IsWeek(unit string) bool {
	return Match(regexpWeek, unit)
}

func IsDay(unit string) bool {
	return Match(regexpDay, unit)
}

func IsHour(unit string) bool {
	return Match(regexpHour, unit)
}

func IsMinute(unit string) bool {
	return Match(regexpMinute, unit) || Match(regexpMinuteM, unit)
}

func IsSecond(unit string) bool {
	return Match(regexpSecond, unit)
}

func IsByte(unit string) bool {
	return Match(regexpByte, unit)
}

func IsKB(unit string) bool {
	return Match(regexpKB, unit)
}

func IsMB(unit string) bool {
	return Match(regexpMB, unit)
}

func IsGB(unit string) bool {
	return Match(regexpGB, unit)
}

func IsTB(unit string) bool {
	return Match(regexpTB, unit)
}

func IsFixedAnnually(unit string) bool {
	return Match(regexpFixedAnnually, unit)
}

func IsFixedMonthly(unit string) bool {
	return Match(regexpFixedMonthly, unit)
}

func IsFixedWeekly(unit string) bool {
	return Match(regexpFixedWeekly, unit)
}

func IsFixedDaily(unit string) bool {
	return Match(regexpFixedDaily, unit)
}

func IsFixedHourly(unit string) bool {
	return Match(regexpFixedHourly, unit)
}

func IsFixedMinutely(unit string) bool {
	return Match(regexpFixedMinutely, unit)
}

func IsFixedSecondly(unit string) bool {
	return Match(regexpFixedSecondly, unit)
}
