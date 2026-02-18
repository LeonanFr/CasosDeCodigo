package db

import (
	"casos-de-codigo-api/internal/models"
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoManager struct {
	Client          *mongo.Client
	Database        *mongo.Database
	UsersColl       *mongo.Collection
	CasesColl       *mongo.Collection
	ProgressionColl *mongo.Collection
}

func NewMongoManager(uri string, dbName string) (*MongoManager, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)

	manager := &MongoManager{
		Client:          client,
		Database:        db,
		UsersColl:       db.Collection("users"),
		CasesColl:       db.Collection("cases"),
		ProgressionColl: db.Collection("progression"),
	}

	err = manager.createIndexes()
	if err != nil {
		log.Printf("Erro ao criar índices: %v", err)
	}

	return manager, nil
}

func (m *MongoManager) createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	progressionIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "case_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := m.UsersColl.Indexes().CreateMany(ctx, userIndexes)
	if err != nil {
		return err
	}

	_, err = m.ProgressionColl.Indexes().CreateMany(ctx, progressionIndexes)
	return err
}

func (m *MongoManager) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

func (m *MongoManager) CreateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := m.UsersColl.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		user.ID = oid
	}

	return nil
}

func (m *MongoManager) FindUserByUsername(username string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := m.UsersColl.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *MongoManager) FindUserByID(id primitive.ObjectID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := m.UsersColl.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *MongoManager) GetCase(caseID string) (*models.Case, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var caso models.Case
	err := m.CasesColl.FindOne(ctx, bson.M{"_id": caseID}).Decode(&caso)
	if err != nil {
		return nil, err
	}
	return &caso, nil
}

func (m *MongoManager) GetAllCases() ([]models.Case, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cases := make([]models.Case, 0)

	findOptions := options.Find().SetSort(bson.D{{Key: "order", Value: 1}})

	cursor, err := m.CasesColl.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &cases)
	return cases, err
}

func (m *MongoManager) GetProgression(userID primitive.ObjectID, caseID string) (*models.Progression, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var progression models.Progression
	err := m.ProgressionColl.FindOne(ctx, bson.M{
		"user_id": userID,
		"case_id": caseID,
	}).Decode(&progression)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &progression, nil
}

func (m *MongoManager) UpsertProgression(p *models.Progression) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	p.UpdatedAt = time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = p.UpdatedAt
	}

	filter := bson.M{"user_id": p.UserID, "case_id": p.CaseID}

	var existing models.Progression
	err := m.ProgressionColl.FindOne(ctx, filter).Decode(&existing)

	if err == nil && existing.Completed {
		p.Completed = true
	}

	opts := options.Replace().SetUpsert(true)

	_, err = m.ProgressionColl.ReplaceOne(ctx, filter, p, opts)
	return err
}

func (m *MongoManager) ResetProgression(userID primitive.ObjectID, caseID string, startingPuzzle int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.ProgressionColl.UpdateOne(
		ctx,
		bson.M{"user_id": userID, "case_id": caseID},
		bson.M{
			"$set": bson.M{
				"current_puzzle": startingPuzzle,
				"current_focus":  "none",
				"sql_history":    []models.SQLHistoryItem{},
				"updated_at":     time.Now(),
			},
		},
	)
	return err
}

func (m *MongoManager) AddSQLHistory(userID primitive.ObjectID, caseID string, item models.SQLHistoryItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.ProgressionColl.UpdateOne(
		ctx,
		bson.M{
			"user_id": userID,
			"case_id": caseID,
		},
		bson.M{
			"$push": bson.M{"sql_history": item},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (m *MongoManager) GetUserProgressions(userID primitive.ObjectID) ([]models.Progression, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	progressions := make([]models.Progression, 0)
	cursor, err := m.ProgressionColl.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &progressions)
	return progressions, err
}
