// Package testutils contains utility functions for testing
package testutils

import (
	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/services/user-service/internal/services/account"
	"github.com/rijum8906/relay/services/user-service/internal/services/auth"
	"github.com/rijum8906/relay/services/user-service/internal/services/profile"
	"github.com/rijum8906/relay/services/user-service/internal/tests/mocks"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

func MockAccountService(mockAccountRepo *mocks.MockAccountRepo) account.AccountService {
	return account.NewAccountService(mockAccountRepo, nil, &utils.UtilsConfig{}, &env.Env{})
}

func MockProfileService(mockProfileRepo *mocks.MockProfileRepo) profile.ProfileService {
	return profile.NewProfileService(mockProfileRepo, nil, &env.Env{})
}

func MockAuthService(mockAuthRepo *mocks.MockAuthRepo, mockAccountRepo *mocks.MockAccountRepo, mockProfileRepo *mocks.MockProfileRepo) auth.AuthService {
	return auth.NewAuthService(&auth.Repo{
		AuthRepo:    mockAuthRepo,
		AccountRepo: mockAccountRepo,
		ProfileRepo: mockProfileRepo,
	}, nil, &utils.UtilsConfig{}, &env.Env{})
}
