package token

import (
	"context"
	"fmt"

	"github.com/rijum8906/relay/packages/core/apperror"
)

// RevokeAllUserTokens revokes all tokens for a user using the set
func (tm *TokenManager) RevokeAllUserTokens(ctx context.Context, userID string) *apperror.AppError {
	userTokenKey := fmt.Sprintf("user_tokens:%s", userID)

	// Get all token IDs for this user
	tokenIDs, err := tm.redisClient.SMembers(ctx, userTokenKey).Result()
	if err != nil {
		// If set doesn't exist, no tokens to revoke
		return nil
	}

	// Delete all individual tokens
	for _, tokenID := range tokenIDs {
		tm.redisClient.Del(ctx, tokenID)
	}

	// Delete the set itself
	if err := tm.redisClient.Del(ctx, userTokenKey).Err(); err != nil {
		return apperror.ErrInternal.
			WithDetail("internal_message", "failed to delete user token set").
			WithDetail("redis_error", err.Error())
	}

	return nil
}

// RevokeAuthToken revokes a single auth token
func (tm *TokenManager) RevokeAuthToken(ctx context.Context, tokenID, userID string) *apperror.AppError {
	// Delete the token
	if err := tm.redisClient.Del(ctx, tokenID).Err(); err != nil {
		return apperror.ErrInternal.
			WithDetail("internal_message", "failed to delete token").
			WithDetail("redis_error", err.Error())
	}

	// Remove from user's set
	userTokenKey := fmt.Sprintf("user_tokens:%s", userID)
	if err := tm.redisClient.SRem(ctx, userTokenKey, tokenID).Err(); err != nil {
		return apperror.ErrInternal.
			WithDetail("internal_message", "failed to remove token from user token set").
			WithDetail("redis_error", err.Error())
	}

	return nil
}
