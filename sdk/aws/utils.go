package aws

import (
	"strings"
)

func ExtractStackNameFromArn(arn string) string {
	parts := strings.Split(arn, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}

func GetWindowInterval(intervals *[]Interval) Interval {
	windowInterval := Interval{}
	for _, interval := range *intervals {
		if windowInterval.Start == nil || interval.Start.Timestamp.Before(windowInterval.Start.Timestamp) {
			windowInterval.Start = interval.Start
		}
		if windowInterval.End == nil || interval.End.Timestamp.After(windowInterval.End.Timestamp) {
			windowInterval.End = interval.End
		}
	}
	return windowInterval
}
