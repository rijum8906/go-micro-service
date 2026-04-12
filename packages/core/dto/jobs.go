package dto

import "strings"

// ============================================
// NOTE: Job Subjects (follow {domain}.{entity}.{action})
// ============================================

type JobSubject string

func (j JobSubject) String() string {
	return string(j)
}

func (j JobSubject) Domain() string {
	parts := strings.Split(string(j), ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func (j JobSubject) Entity() string {
	parts := strings.Split(string(j), ".")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func (j JobSubject) Action() string {
	parts := strings.Split(string(j), ".")
	if len(parts) > 2 {
		return parts[2]
	}
	return ""
}

// Email Jobs - domain=email
const (
	JobEmailVerification  JobSubject = "email.verification.send"
	JobEmailPasswordReset JobSubject = "email.password.reset.send"
	JobEmailWelcome       JobSubject = "email.welcome.send"
	JobEmailNotification  JobSubject = "email.notification.send"
	JobEmailTaskAssigned  JobSubject = "email.task.assigned.send"
	JobEmailDailyDigest   JobSubject = "email.daily.digest.send"
)

// Cleanup Jobs - domain=cleanup
const (
	JobCleanupLogs      JobSubject = "cleanup.logs.delete"
	JobCleanupSessions  JobSubject = "cleanup.sessions.delete"
	JobCleanupTasks     JobSubject = "cleanup.tasks.archive"
	JobCleanupArchived  JobSubject = "cleanup.archived.purge"
	JobCleanupTempFiles JobSubject = "cleanup.temp.delete"
)

// Notification Jobs - domain=notification
const (
	JobNotificationPush    JobSubject = "notification.push.send"
	JobNotificationSMS     JobSubject = "notification.sms.send"
	JobNotificationWebhook JobSubject = "notification.webhook.trigger"
)

// ============================================
// Helper Functions
// ============================================

// GetDomainWildcard returns wildcard for entire domain
func GetDomainWildcard(domain Domain) string {
	return string(domain) + ".*"
}

// GetEntityWildcard returns wildcard for specific entity within domain
func GetEntityWildcard(domain Domain, entity string) string {
	return string(domain) + "." + entity + ".*"
}

// GetActionWildcard returns wildcard for specific action
func GetActionWildcard(domain Domain, entity, action string) string {
	return string(domain) + "." + entity + "." + action
}

// GetAllJobSubjects returns all job subjects
func GetAllJobSubjects() []JobSubject {
	return []JobSubject{
		JobEmailVerification,
		JobEmailPasswordReset,
		JobEmailWelcome,
		JobEmailNotification,
		JobEmailTaskAssigned,
		JobEmailDailyDigest,
		JobCleanupLogs,
		JobCleanupSessions,
		JobCleanupTasks,
		JobCleanupArchived,
		JobNotificationPush,
		JobNotificationSMS,
		JobNotificationWebhook,
	}
}

// IsValidJobSubject checks if subject follows job pattern
func IsValidJobSubject(subject string) bool {
	parts := strings.Split(subject, ".")
	if len(parts) != 3 && len(parts) != 4 {
		return false
	}

	domain := Domain(parts[0])
	switch domain {
	case DomainEmail, DomainNotification, DomainCleanup:
		return true
	}
	return false
}

// FormatJobSubject creates a formatted job subject
func FormatJobSubject(domain Domain, entity, action string) JobSubject {
	return JobSubject(string(domain) + "." + entity + "." + action)
}
