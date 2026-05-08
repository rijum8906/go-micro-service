package orgmembership_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	orgmembershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	"github.com/rijum8906/relay/services/organization-service/app/config"
	orgmembership "github.com/rijum8906/relay/services/organization-service/internal/services/org_membership"
	"github.com/stretchr/testify/require"
)

// ######################## GetOrganizationMembership ########################
func Test_GetOrganizationMembership_Validation_Failure(t *testing.T) {
	service := orgmembership.New(nil, nil, &coreopenfga.Client{}, &config.Env{})
	ctx := context.Background()

	tests := []struct {
		name string
		req  *corev1.IDRequest
	}{
		{
			name: "nil request",
			req:  nil,
		},
		{
			name: "blank id",
			req: &corev1.IDRequest{
				Id: "",
			},
		},
		{
			name: "malformed uuid",
			req: &corev1.IDRequest{
				Id: "malformed-uuid",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetOrganizationMembership(ctx, tt.req)
			require.Error(t, err)
			require.Contains(t, err.Error(), string(apperror.CodeValidation))
		})
	}
}

// ######################## SendInvitation ########################

func Test_SendInvitation_Validation_Failure(t *testing.T) {
	service := orgmembership.New(nil, nil, &coreopenfga.Client{}, &config.Env{})
	ctx := context.Background()

	tests := []struct {
		name string
		req  *orgmembershipv1.SendInvitationRequest
	}{
		{
			name: "nil request",
			req:  nil,
		},
		{
			name: "blank organization id",
			req: &orgmembershipv1.SendInvitationRequest{
				OrganizationId: "",
				Email:          "user@example.com",
				Role:           "member",
			},
		},
		{
			name: "malformed organization id",
			req: &orgmembershipv1.SendInvitationRequest{
				OrganizationId: "malformed-uuid",
				Email:          "user@example.com",
				Role:           "member",
			},
		},
		{
			name: "blank email",
			req: &orgmembershipv1.SendInvitationRequest{
				OrganizationId: uuid.NewString(),
				Email:          "",
				Role:           "member",
			},
		},
		{
			name: "invalid email format",
			req: &orgmembershipv1.SendInvitationRequest{
				OrganizationId: uuid.NewString(),
				Email:          "not-an-email",
				Role:           "member",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.SendInvitation(ctx, tt.req)
			require.Error(t, err)
			require.Contains(t, err.Error(), string(apperror.CodeValidation))
		})
	}
}
