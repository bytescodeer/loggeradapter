package loggeradapter

import (
	"testing"
)

func TestParseExpression(t *testing.T) {
	for _, rotation := range testStrings {
		v, unit, err := ParseExpression(rotation)
		if err != nil {
			t.Error("error:", err)
			continue
		}

		t.Logf("匹配成功: %s，值: %d，单位: %s, isDuration: %v, isFileSize: %v\n",
			rotation, v, unit, IsDuration(unit), IsFileSize(unit))
	}
}

var testStrings = []string{
	"1y", "2Y", "3year", "4Year", "5YEAR", // 匹配
	"1M", "2mo", "3mon", "4month", "5Month", "6MONTH", // 匹配
	"1w", "2W", "3week", "4Week", "5WEEK", // 匹配
	"1d", "2D", "3day", "4Day", "5DAY", // 匹配
	"1h", "2H", "3hour", "4Hour", "5HOUR", // 匹配
	"1m", "2minute", "3Minute", "4MINUTE", "5min", "6Min", // 匹配
	"1second", "2Second", "3SECOND", "4s", "5S", // 匹配

	"y", "Y", "year", "Year", "YEAR", // 不匹配
	"M", "mo", "mon", "month", "Month", "MONTH", // 不匹配
	"w", "W", "week", "Week", "WEEK", // 不匹配
	"d", "D", "day", "Day", "DAY", // 不匹配
	"h", "H", "hour", "Hour", "HOUR", // 不匹配
	"m", "minute", "Minute", "MINUTE", "min", "Min", // 不匹配
	"second", "Second", "SECOND", "s", "S", // 不匹配
	"1invalid", "2yday", "4mmonth", // 不匹配

	"1b", "2B", "3byte", "4Byte", "5BYTE", // 匹配
	"1kb", "2KB", "3kilobyte", "4KiloByte", "5KILOBYTE", "6Kb", // 匹配
	"1mb", "2MB", "3Mb", "4megaByte", "5MegaByte", "6MEGABYTE", // 匹配
	"1gb", "2GB", "3Gb", "4gigaByte", "5GigaByte", "6GIGABYTE", // 匹配
	"1tb", "2TB", "3Tb", "4teraByte", "5TeraByte", "6TERABYTE", // 匹配
	"1pb", "2PB", "3Pb", "4petaByte", "5PetaByte", "6PETABYTE", // 匹配
	"1zb", "2ZB", "3Zb", "4zettaByte", "5ZettaByte", "6ZETTABYTE", // 匹配

	"b", "B", "byte", "Byte", "BYTE", // 匹配
	"kb", "KB", "kilobyte", "KiloByte", "KILOBYTE", "Kb", // 不匹配
	"mb", "MB", "Mb", "megaByte", "MegaByte", "MEGABYTE", // 不匹配
	"gb", "GB", "Gb", "gigaByte", "GigaByte", "GIGABYTE", // 不匹配
	"tb", "TB", "Tb", "teraByte", "TeraByte", "TERABYTE", // 不匹配
	"pb", "PB", "Pb", "petaByte", "PetaByte", "PETABYTE", // 不匹配
	"zb", "ZB", "Zb", "zettaByte", "ZettaByte", "ZETTABYTE", // 不匹配

	"1year", "1Month", "1Week", "1Day", "1Hour", "1Minute", "1Second", // 匹配
	"Annually", "Monthly", "Weekly", "Daily", "Hourly", "Minutely", "Secondly", // 匹配
	"year", "Month", "Week", "Day", "Hour", "Minute", "Second", // 不匹配
}
