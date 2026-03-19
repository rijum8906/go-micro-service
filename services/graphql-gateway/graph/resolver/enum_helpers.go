package resolver

import (
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/model"
)

func authTypeToProto(authType model.AuthType) user_servicev1.AuthType {
	switch authType {
	case model.AuthTypeAuthTypePassword:
		return user_servicev1.AuthType_AUTH_TYPE_PASSWORD
	default:
		return user_servicev1.AuthType_AUTH_TYPE_UNSPECIFIED
	}
}

func scopedActionToProto(action model.ScopedAction) user_servicev1.ScopedAction {
	switch action {
	case model.ScopedActionScopedActionChangePassword:
		return user_servicev1.ScopedAction_SCOPED_ACTION_CHANGE_PASSWORD
	case model.ScopedActionScopedActionChangeEmail:
		return user_servicev1.ScopedAction_SCOPED_ACTION_CHANGE_EMAIL
	case model.ScopedActionScopedActionDeleteAccount:
		return user_servicev1.ScopedAction_SCOPED_ACTION_DELETE_ACCOUNT
	case model.ScopedActionScopedActionDeleteProfile:
		return user_servicev1.ScopedAction_SCOPED_ACTION_DELETE_PROFILE
	default:
		return user_servicev1.ScopedAction_SCOPED_ACTION_UNSPECIFIED
	}
}
