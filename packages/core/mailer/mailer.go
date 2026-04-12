// Package mailer
package mailer

type EmailPriority int

const (
	EmailPriorityLow EmailPriority = iota
	EmailPriorityNormal
	EmailPriorityHigh
)
