package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (m *MongoManager) CleanupOldSessions() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	twoWeeksAgo := time.Now().Add(-14 * 24 * time.Hour)

	_, err := m.ProgressionColl.DeleteMany(ctx, bson.M{
		"updated_at": bson.M{"$lt": twoWeeksAgo},
	})
	return err
}

func (m *MongoManager) UpdateUserProgress(userID primitive.ObjectID, caseID string, puzzle int, focus string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.ProgressionColl.UpdateOne(
		ctx,
		bson.M{
			"user_id": userID,
			"case_id": caseID,
		},
		bson.M{
			"$set": bson.M{
				"current_puzzle": puzzle,
				"current_focus":  focus,
				"updated_at":     time.Now(),
			},
		},
	)
	return err
}
