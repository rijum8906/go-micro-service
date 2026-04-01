// Package dto
package dto

import "strings"

// ============================================
// Subject Naming Convention
// ============================================
// NOTE: Format: {domain}.{entity}.{action}[.{subtype}]
// Examples:
//   - user.user.created (domain=user, entity=user, action=created)
//   - user.profile.updated (domain=user, entity=profile, action=updated)
//   - task.task.assigned (domain=task, entity=task, action=assigned)
//   - email.verification.send (domain=email, entity=verification, action=send)
//
// NOTE: Benefits:
// - Wildcard subscriptions: user.* (all user domain events)
// - Entity-level wildcards: user.user.* (all user entity events)
// - Hierarchical organization
// - Easy to filter and route
// - Scalable naming

// ============================================
// Domain Types
// ============================================

type Domain string

const (
	DomainUser         Domain = "user"
	DomainTask         Domain = "task"
	DomainProject      Domain = "project"
	DomainOrganization Domain = "organization"
	DomainSession      Domain = "session"
	DomainEmail        Domain = "email"
	DomainNotification Domain = "notification"
	DomainCleanup      Domain = "cleanup"
)

// ============================================
// Event Subjects (follow {domain}.{entity}.{action})
// ============================================

type EventSubject string

func (e EventSubject) String() string {
	return string(e)
}

func (e EventSubject) Domain() string {
	parts := strings.Split(string(e), ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func (e EventSubject) Entity() string {
	parts := strings.Split(string(e), ".")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func (e EventSubject) Action() string {
	parts := strings.Split(string(e), ".")
	if len(parts) > 2 {
		return parts[2]
	}
	return ""
}

// User Domain Events - domain=user
const (
	// User entity events (entity=user)
	EventUserCreated  EventSubject = "user.user.created"
	EventUserUpdated  EventSubject = "user.user.updated"
	EventUserDeleted  EventSubject = "user.user.deleted"
	EventUserArchived EventSubject = "user.user.archived"
	EventUserRestored EventSubject = "user.user.restored"

	// User authentication events (entity=auth)
	EventUserLoggedIn        EventSubject = "user.auth.logged_in"
	EventUserLoggedOut       EventSubject = "user.auth.logged_out"
	EventUserPasswordChanged EventSubject = "user.auth.password_changed"
	EventUserPasswordReset   EventSubject = "user.auth.password_reset"

	// User profile events (entity=profile)
	EventUserProfileUpdated EventSubject = "user.profile.updated"
	EventUserAvatarUpdated  EventSubject = "user.profile.avatar_updated"
	EventUserBioUpdated     EventSubject = "user.profile.bio_updated"

	// User verification events (entity=verification)
	EventUserEmailVerified EventSubject = "user.verification.email_verified"
	EventUserEmailChanged  EventSubject = "user.verification.email_changed"
	EventUserPhoneVerified EventSubject = "user.verification.phone_verified"

	// User settings events (entity=settings)
	EventUserPreferencesUpdated   EventSubject = "user.settings.preferences_updated"
	EventUserNotificationsUpdated EventSubject = "user.settings.notifications_updated"
)

// Task Domain Events - domain=task
const (
	// Task entity events (entity=task)
	EventTaskCreated  EventSubject = "task.task.created"
	EventTaskUpdated  EventSubject = "task.task.updated"
	EventTaskDeleted  EventSubject = "task.task.deleted"
	EventTaskArchived EventSubject = "task.task.archived"

	// Task assignment events (entity=assignment)
	EventTaskAssigned   EventSubject = "task.assignment.assigned"
	EventTaskUnassigned EventSubject = "task.assignment.unassigned"
	EventTaskReassigned EventSubject = "task.assignment.reassigned"

	// Task workflow events (entity=workflow)
	EventTaskStarted   EventSubject = "task.workflow.started"
	EventTaskCompleted EventSubject = "task.workflow.completed"
	EventTaskBlocked   EventSubject = "task.workflow.blocked"
	EventTaskUnblocked EventSubject = "task.workflow.unblocked"

	// Task comments (entity=comment)
	EventTaskCommented      EventSubject = "task.comment.created"
	EventTaskCommentUpdated EventSubject = "task.comment.updated"
	EventTaskCommentDeleted EventSubject = "task.comment.deleted"
)

// Project Domain Events - domain=project
const (
	// Project entity events
	EventProjectCreated  EventSubject = "project.project.created"
	EventProjectUpdated  EventSubject = "project.project.updated"
	EventProjectDeleted  EventSubject = "project.project.deleted"
	EventProjectArchived EventSubject = "project.project.archived"

	// Project member events (entity=member)
	EventProjectMemberAdded       EventSubject = "project.member.added"
	EventProjectMemberRemoved     EventSubject = "project.member.removed"
	EventProjectMemberRoleChanged EventSubject = "project.member.role_changed"
)

// GetAllEventSubjects returns all event subjects
func GetAllEventSubjects() []EventSubject {
	return []EventSubject{
		// User events
		EventUserCreated,
		EventUserUpdated,
		EventUserDeleted,
		EventUserLoggedIn,
		EventUserLoggedOut,
		EventUserPasswordChanged,
		EventUserEmailVerified,
		EventUserPreferencesUpdated,

		// Task events
		EventTaskCreated,
		EventTaskUpdated,
		EventTaskDeleted,
		EventTaskAssigned,
		EventTaskCompleted,
		EventTaskCommented,

		// Project events
		EventProjectCreated,
		EventProjectUpdated,
		EventProjectDeleted,
	}
}

// GetUserEventSubjects returns user-related event subjects
func GetUserEventSubjects() []EventSubject {
	return []EventSubject{
		EventUserCreated,
		EventUserUpdated,
		EventUserDeleted,
		EventUserArchived,
		EventUserRestored,
		EventUserLoggedIn,
		EventUserLoggedOut,
		EventUserPasswordChanged,
		EventUserPasswordReset,
		EventUserEmailVerified,
		EventUserEmailChanged,
		EventUserPhoneVerified,
		EventUserPreferencesUpdated,
		EventUserNotificationsUpdated,
	}
}

// GetTaskEventSubjects returns task-related event subjects
func GetTaskEventSubjects() []EventSubject {
	return []EventSubject{
		EventTaskCreated,
		EventTaskUpdated,
		EventTaskDeleted,
		EventTaskAssigned,
		EventTaskCompleted,
		EventTaskCommented,
	}
}

// IsValidEventSubject checks if subject follows event pattern
func IsValidEventSubject(subject string) bool {
	parts := strings.Split(subject, ".")
	if len(parts) != 3 {
		return false
	}

	// Check domain is valid
	domain := Domain(parts[0])
	switch domain {
	case DomainUser, DomainTask, DomainProject, DomainOrganization, DomainSession:
		return true
	}
	return false
}

// FormatEventSubject creates a formatted event subject
func FormatEventSubject(domain Domain, entity, action string) EventSubject {
	return EventSubject(string(domain) + "." + entity + "." + action)
}
