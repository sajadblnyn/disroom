package repository_test

import (
	"context"
	"testing"

	"github.com/go-redis/redismock/v8"
	"github.com/sajadblnyn/disroom/config"
	"github.com/sajadblnyn/disroom/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisRepository(t *testing.T) {
	db, mock := redismock.NewClientMock()
	config.RedisClient = db
	ctx := context.Background()

	t.Run("AddUserToRoom", func(t *testing.T) {
		roomID := "test-room"
		userID := "user-123"

		mock.ExpectSAdd("room:test-room:users", userID).SetVal(1)
		err := repository.AddUserToRoom(ctx, userID, roomID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("RemoveUserFromRoom", func(t *testing.T) {
		roomID := "test-room"
		userID := "user-123"

		mock.ExpectSRem("room:test-room:users", userID).SetVal(1)
		err := repository.RemoveUserFromRoom(ctx, userID, roomID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetActiveUsers", func(t *testing.T) {
		roomID := "test-room"
		expectedUsers := []string{"user-123", "user-456"}

		mock.ExpectSMembers("room:test-room:users").SetVal(expectedUsers)
		users, err := repository.GetActiveUsers(ctx, roomID)
		require.NoError(t, err)
		assert.Equal(t, expectedUsers, users)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		roomID := "error-room"
		userID := "user-999"

		t.Run("AddUserError", func(t *testing.T) {
			mock.ExpectSAdd("room:error-room:users", userID).SetErr(assert.AnError)
			err := repository.AddUserToRoom(ctx, userID, roomID)
			assert.Error(t, err)
		})

		t.Run("RemoveUserError", func(t *testing.T) {
			mock.ExpectSRem("room:error-room:users", userID).SetErr(assert.AnError)
			err := repository.RemoveUserFromRoom(ctx, userID, roomID)
			assert.Error(t, err)
		})

		t.Run("GetUsersError", func(t *testing.T) {
			mock.ExpectSMembers("room:error-room:users").SetErr(assert.AnError)
			_, err := repository.GetActiveUsers(ctx, roomID)
			assert.Error(t, err)
		})

	})
}
