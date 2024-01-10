package aws

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type Event struct {
	EventId              string
	StackId              string
	StackName            string
	Timestamp            time.Time
	LogicalResourceId    string
	PhysicalResourceId   string
	ResourceStatus       types.ResourceStatus
	ResourceStatusReason string
	ResourceType         string
}

func (e *Event) String() string {
	if e == nil {
		return "<NULL>"
	}
	return e.EventId
}

func (e *Event) IsOperation() bool {
	if e == nil {
		return false
	}
	return e.ResourceStatusReason == "User Initiated"
}
