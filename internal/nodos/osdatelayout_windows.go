package nodos

import (
	"regexp"
	"strings"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func international(key string) (string, error) {
	k, err := registry.OpenKey(
		registry.CURRENT_USER,
		`Control Panel\International`,
		registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()

	val, _, err := k.GetStringValue(key)
	return val, err
}

var rxHasSingleD = regexp.MustCompile(`\bd\b`)

func osDateLayout() (string, error) {
	shortDate, err := international("sShortDate")
	if err != nil {
		return "", err
	}

	// When the layout has a single 'd',
	// on the codepage 932, the weekday is appended at the tail.
	// on the codepage 437, the weekday is inserted at the head.
	// The source of the information is
	// https://kurasaba.hatenablog.com/entries/2006/01/31

	layout := table.Replace(shortDate)
	if rxHasSingleD.MatchString(shortDate) {
		if windows.GetACP() == 932 {
			layout = layout + " Mon"
		} else {
			layout = "Mon " + layout
		}
	}
	return layout, nil
}

var table = strings.NewReplacer(
	"yyyy", "2006",
	"dddd", "Mon",
	"MM", "01",
	"dd", "02",
	"d", "02",
	"M", "01",
	"H", "15",
	"mm", "04",
	"ss", "05",
)

var weekDayReplacer = strings.NewReplacer(
	"Mon", "月",
	"Tue", "火",
	"Wed", "水",
	"Thu", "木",
	"Fri", "金",
	"Sat", "土",
	"Sun", "日",
)

func timeFormatOsLayout(t time.Time) (string, error) {
	layout, err := osDateLayout()
	if err != nil {
		return "", err
	}
	s := t.Format(layout)
	if windows.GetACP() == 932 {
		s = weekDayReplacer.Replace(s)
	}
	return s, nil
}
