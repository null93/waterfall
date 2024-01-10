package aws

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"golang.org/x/exp/slices"
)

type DataSet struct {
	cfnClient        *cloudformation.Client
	loading          bool
	stacks           []string
	operations       []Event
	stackEvents      map[string][]Event
	OriginalStackArn string
	StackIntervals   IntervalMap
}

func NewDataSet(cfg aws.Config, arn string) *DataSet {
	ds := &DataSet{
		cfnClient:        cloudformation.NewFromConfig(cfg),
		loading:          false,
		stacks:           []string{arn},
		operations:       []Event{},
		stackEvents:      map[string][]Event{},
		OriginalStackArn: arn,
		StackIntervals:   IntervalMap{},
	}
	ds.AddStackArn(arn)
	return ds
}

func (ds *DataSet) GetStackEvents(stackArn string) []Event {
	return ds.stackEvents[stackArn]
}

func (ds *DataSet) GetAllStackEvents() []Event {
	allEvents := []Event{}
	for _, events := range ds.stackEvents {
		allEvents = append(allEvents, events...)
	}
	return allEvents
}

func (ds *DataSet) GetLatestOperation(selectedStack string, allStacks bool) string {
	latestEventId := ""
	latestTimestamp := time.Unix(0, 0)
	for _, operation := range ds.GetOperations(selectedStack, allStacks) {
		if operation.Timestamp.After(latestTimestamp) {
			latestEventId = operation.EventId
			latestTimestamp = operation.Timestamp
		}
	}
	return latestEventId
}

func (ds *DataSet) GetStackArns() []string {
	return ds.stacks
}

func (ds *DataSet) AddStackArn(stackArn string) {
	if _, ok := ds.stackEvents[stackArn]; !ok {
		ds.stackEvents[stackArn] = []Event{}
	}
	for _, arn := range ds.stacks {
		if arn == stackArn {
			return
		}
	}
	ds.stacks = append(ds.stacks, stackArn)
}

func (ds *DataSet) GetOperations(stackArn string, allStacks bool) []Event {
	ops := []Event{}
	for _, event := range ds.operations {
		if event.StackId == stackArn || allStacks {
			ops = append(ops, event)
		}
	}
	return ops
}

func (ds *DataSet) Refresh() error {
	ds.loading = true
	if err := ds.refreshEvents(); err != nil {
		ds.loading = false
		return err
	}
	ds.loading = false
	ds.refreshIntervals()
	return nil
}

func (ds *DataSet) refreshEvents() error {
	operations := []Event{}
	for _, stackArn := range ds.stacks {
		events := []Event{}
		params := cloudformation.DescribeStackEventsInput{StackName: aws.String(stackArn)}
		paginator := cloudformation.NewDescribeStackEventsPaginator(ds.cfnClient, &params)
		for paginator.HasMorePages() {
			output, err := paginator.NextPage(context.TODO())
			if err != nil {
				return err
			}
			for _, event := range output.StackEvents {
				item := Event{
					EventId:              aws.ToString(event.EventId),
					StackId:              aws.ToString(event.StackId),
					StackName:            aws.ToString(event.StackName),
					Timestamp:            aws.ToTime(event.Timestamp),
					LogicalResourceId:    aws.ToString(event.LogicalResourceId),
					PhysicalResourceId:   aws.ToString(event.PhysicalResourceId),
					ResourceStatus:       event.ResourceStatus,
					ResourceStatusReason: aws.ToString(event.ResourceStatusReason),
					ResourceType:         aws.ToString(event.ResourceType),
				}
				if item.ResourceStatusReason == "User Initiated" {
					operations = append(operations, item)
				}
				events = append(events, item)
			}
		}
		newEvents := []Event{}
		for i := 0; i < len(events); i++ {
			newEvents = append(newEvents, events[i])
		}
		ds.stackEvents[stackArn] = newEvents
	}
	slices.SortFunc(operations, func(a, b Event) int { return b.Timestamp.Compare(a.Timestamp) })
	ds.operations = operations
	return nil
}

func (ds *DataSet) refreshIntervals() {
	events := ds.GetAllStackEvents()
	stackIntervals := IntervalMap{}
	temp := Interval{}
	seen := map[string]bool{}
	lastOperationStack := ""
	lastOperationEventId := ""
	for i := len(events) - 1; i >= 0; i-- {
		event := events[i]
		if event.IsOperation() {
			lastOperationStack = event.StackId
			lastOperationEventId = event.EventId
		}
		if lastOperationStack == "" || lastOperationEventId == "" {
			panic("last operation stack and event id cannot be empty")
		}
		if _, ok := seen[event.EventId]; !ok {
			isDuplicateCompleteEvent := strings.HasSuffix(string(event.ResourceStatus), "_COMPLETE")
			if !isDuplicateCompleteEvent {
				if temp.Start == nil {
					temp.Start = &event
					for j := i - 1; j >= 0; j-- {
						target := events[j]
						if _, ok := seen[target.EventId]; !ok && target.LogicalResourceId == event.LogicalResourceId && target.StackId == event.StackId {
							hasValidEnd := strings.HasSuffix(string(target.ResourceStatus), "_FAILED") || strings.HasSuffix(string(target.ResourceStatus), "_COMPLETE")
							if event.ResourceStatus == target.ResourceStatus || !hasValidEnd {
								seen[target.EventId] = true
								temp.Intermediate = append(temp.Intermediate, &target)
							} else {
								seen[target.EventId] = true
								temp.End = &target
								stackIntervals.AppendInterval(lastOperationStack, lastOperationEventId, temp)
								temp = Interval{}
								break
							}
						}
					}
					if temp.Start != nil && temp.End == nil {
						temp.End = &Event{Timestamp: time.Now(), ResourceStatus: temp.Start.ResourceStatus}
						stackIntervals.AppendInterval(lastOperationStack, lastOperationEventId, temp)
						temp = Interval{}
					}
				}
			} else {
				seen[event.EventId] = true
				temp.Intermediate = append(temp.Intermediate, &event)
			}
		}
		seen[event.EventId] = true
	}
	for _, operationIntervals := range stackIntervals {
		for _, intervals := range operationIntervals {
			slices.Reverse(intervals)
		}
	}
	ds.StackIntervals = stackIntervals
}

func (ds *DataSet) AddNestedStacks() error {
	ds.loading = true
	params := cloudformation.DescribeStackResourcesInput{StackName: aws.String(ds.OriginalStackArn)}
	response, err := ds.cfnClient.DescribeStackResources(context.TODO(), &params)
	if err != nil {
		ds.loading = false
		return err
	}
	for _, resource := range response.StackResources {
		if aws.ToString(resource.ResourceType) == "AWS::CloudFormation::Stack" {
			stackArn := aws.ToString(resource.PhysicalResourceId)
			ds.AddStackArn(stackArn)
		}
	}
	ds.loading = false
	return nil
}

func (ds *DataSet) IsLoading() bool {
	return ds.loading
}

func (ds *DataSet) GetSortedIntervals(selectedStack, selectedOperation string, allStacks, allOperations bool) []Interval {
	allIntervals := []Interval{}
	for _, operation := range ds.operations {
		if allStacks || operation.StackId == selectedStack {
			if allOperations || operation.EventId == selectedOperation {
				intervals := ds.StackIntervals.GetIntervals(operation.StackId, operation.EventId)
				allIntervals = append(allIntervals, intervals...)
			}
		}
	}
	return allIntervals
}
