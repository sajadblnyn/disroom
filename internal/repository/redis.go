package repository

import (
	"context"
	"fmt"

	"github.com/sajadblnyn/disroom/config"
)

func AddUserToRoom(ctx context.Context, userID, roomID string) error {
	return config.RedisClient.SAdd(ctx, fmt.Sprintf("room:%s:users", roomID), userID).Err()
}

func RemoveUserFromRoom(ctx context.Context, userID, roomID string) error {
	return config.RedisClient.SRem(ctx, fmt.Sprintf("room:%s:users", roomID), userID).Err()
}

func GetActiveUsers(ctx context.Context, roomID string) ([]string, error) {
	return config.RedisClient.SMembers(ctx, fmt.Sprintf("room:%s:users", roomID)).Result()
}
