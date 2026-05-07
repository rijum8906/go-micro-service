package orgmembership_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/apperror"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/services/org_membership/testutil"
	"github.com/stretchr/testify/require"
)

// ######################## GetOrganizationMembership ########################
func Test_GetOrganizationMembership_Validation_Failure(t *testing.T) {
	suite := testutil.NewTestSuite(t)
	service := suite.Service

	// malformed uuid
	_, err := service.GetOrganizationMembership(suite.Ctx, &corev1.IDRequest{
		Id: "malformed-uuid",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), string(apperror.CodeValidation))

	// nil request
	_, err = service.GetOrganizationMembership(suite.Ctx, nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), string(apperror.CodeValidation))
}
