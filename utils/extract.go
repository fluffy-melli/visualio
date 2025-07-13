package utils

import (
	"regexp"
	"strconv"
)

type NumberUnit struct {
	Number string
	Unit   string
	Value  float64
}

func ExtractNumber(input string) (NumberUnit, bool) {
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)(%|px)`)
	match := re.FindStringSubmatch(input)

	if len(match) > 2 {
		number := match[1]
		unit := match[2]

		value, err := strconv.ParseFloat(number, 64)
		if err != nil {
			return NumberUnit{}, false
		}

		nu := NumberUnit{
			Number: number,
			Unit:   unit,
			Value:  value,
		}

		return nu, true
	}

	return NumberUnit{}, false
}
