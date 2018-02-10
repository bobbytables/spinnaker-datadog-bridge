package types

// IncomingWebhook is a structure representing a Spinnaker echo rest Webhook
// You can view an example of the schema here:
// https://www.spinnaker.io/setup/features/notifications/#event-types
type IncomingWebhook struct {
	Details Details `json:"details"`
	Content Content `json:"content"`
}

// Details contains all of the details contained in the webhook
type Details struct {
	Source      string `json:"source"`
	Type        string `json:"type"`
	Application string `json:"application"`
	Created     string `json:"created"`
}

// Content is the main context of the given Webhook. It contains of the execution
// information and stage details as an example
type Content struct {
	ExecutionID string    `json:"executionId"`
	StartTime   Timestamp `json:"startTime"`
	EndTime     Timestamp `json:"endTime"`
}
