package handler

import (
	"github.com/aws/aws-lambda-go/events"
)

type SQSCarrier map[string]events.SQSMessageAttribute

func (c SQSCarrier) Get(key string) string {
	val, ok := c[key]
	if ok && val.StringValue != nil {
		return *val.StringValue
	}
	return ""
}

func (c SQSCarrier) Set(key, value string) {}
func (c SQSCarrier) Keys() []string        { return nil }
