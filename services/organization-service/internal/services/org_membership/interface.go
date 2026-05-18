// Package orgmembership
package orgmembership

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/coreopenfga"
	"github.com/rijum8906/relay/packages/core/hash"
	permissions "github.com/rijum8906/relay/packages/core/permissions/organization"
	"github.com/rijum8906/relay/packages/core/testutils"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
	"github.com/rijum8906/relay/services/organization-service/app"
	"github.com/rijum8906/relay/services/organization-service/app/config"
	"github.com/rijum8906/relay/services/organization-service/internal/db"
	servicetestutils "github.com/rijum8906/relay/services/organization-service/internal/service_test_utils"
	"github.com/rijum8906/relay/services/organization-service/internal/services/helper"
	"go.uber.org/zap"
)

type OrgMembershipService struct {
	DBPool        *pgxpool.Pool
	DBQ           *db.Queries
	UserClient    userv1.UserServiceClient
	TuppleManager coreopenfga.TuppleManager
	Permission    *permissions.PermissionManager
	HashService   hash.HashService
	Logger        *zap.Logger
	Config        *config.Env
	Helper        *helper.ServiceHelper
}

func New() (*apperror.AppError, org_membershipv1.OrganizationMembershipServiceServer) {
	application, appErr := app.GetInstance()
	if appErr != nil {
		return appErr, nil
	}

	q := db.New(application.DB())

	fgaClient, appErr := coreopenfga.NewClient(application.Config().FGAAPIURL)
	if appErr != nil {
		return appErr, nil
	}
	tuppleManager := coreopenfga.NewTupleManager(fgaClient)
	permissionManager := permissions.NewPermissionManager(fgaClient)

	helper, appErr := helper.GetHelper()
	if appErr != nil {
		return appErr, nil
	}

	return nil, &OrgMembershipService{
		DBPool:        application.DB(),
		DBQ:           q,
		UserClient:    application.UserClient(),
		TuppleManager: tuppleManager,
		Permission:    permissionManager,
		HashService: hash.NewHashService(hash.Config{
			BcryptCost: 10,
		}),
		Logger: application.Logger(),
		Config: application.Config(),
		Helper: helper,
	}
}

func NewTestService(fgaClient *coreopenfga.Client) *OrgMembershipService {
	dbPool := testutils.MustConnectDB(testutils.WithDBName(testutils.GetTestDBName("organization-service")))
	dbq := db.New(dbPool)

	testApp := &app.Application{}
	logger := testApp.TestLogger()

	return &OrgMembershipService{
		DBPool:        dbPool,
		DBQ:           dbq,
		UserClient:    servicetestutils.MockUserServiceClient,
		TuppleManager: coreopenfga.NewTupleManager(fgaClient),
		Permission:    permissions.NewPermissionManager(fgaClient),
		HashService: hash.NewHashService(hash.Config{
			BcryptCost: 10,
		}),
		Logger: logger,
		Config: &config.Env{},
	}
}
