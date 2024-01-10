package aws

import (
	"fmt"
)

type Interval struct {
	Start        *Event
	Intermediate []*Event
	End          *Event
}

type IntervalMap map[string]map[string][]Interval

func (im IntervalMap) AppendInterval(stackArn string, operationId string, interval Interval) {
	if _, ok := im[stackArn]; !ok {
		im[stackArn] = map[string][]Interval{}
	}
	if _, ok := im[stackArn][operationId]; !ok {
		im[stackArn][operationId] = []Interval{}
	}
	im[stackArn][operationId] = append(im[stackArn][operationId], interval)
}

func (im IntervalMap) GetIntervals(selectedStack, selectedOperation string) []Interval {
	allIntervals := []Interval{}
	for stackArn, operationIntervals := range im {
		if stackArn == selectedStack {
			for operationId, intervals := range operationIntervals {
				if operationId == selectedOperation {
					allIntervals = append(allIntervals, intervals...)
				}
			}
		}
	}
	return allIntervals
}

func (im IntervalMap) String() string {
	str := ""
	for stackArn, operationIntervals := range im {
		for operationId, intervals := range operationIntervals {
			str += fmt.Sprintf("Stack: %-159s, Operation: %s, Count: %d\n", stackArn, operationId, len(intervals))
		}
	}
	return str
}
