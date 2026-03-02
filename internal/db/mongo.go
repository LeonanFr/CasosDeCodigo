package db

import (
	"casos-de-codigo-api/internal/models"
	"context"
	"errors"
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
	TelemetryColl   *mongo.Collection
	TournamentsColl *mongo.Collection
}

func NewMongoManager(uri, dbName string) (*MongoManager, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(dbName)

	manager := &MongoManager{
		Client:          client,
		Database:        db,
		UsersColl:       db.Collection("users"),
		CasesColl:       db.Collection("cases"),
		ProgressionColl: db.Collection("progression"),
		TelemetryColl:   db.Collection("telemetry"),
		TournamentsColl: db.Collection("tournaments"),
	}

	if err := manager.createIndexes(); err != nil {
		log.Printf("Erro ao criar índices: %v", err)
	}

	return manager, nil
}

func (m *MongoManager) createIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.UsersColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	progressionIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "case_id", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.M{
					"user_id": bson.M{"$exists": true},
				}),
		},
		{
			Keys: bson.D{
				{Key: "team_code", Value: 1},
				{Key: "matricula", Value: 1},
				{Key: "case_id", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetPartialFilterExpression(bson.M{
					"team_code": bson.M{"$exists": true},
					"matricula": bson.M{"$exists": true},
				}),
		},
	}

	if _, err := m.ProgressionColl.Indexes().CreateMany(ctx, progressionIndexes); err != nil {
		return err
	}

	telemetryIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "case_id", Value: 1},
				{Key: "timestamp", Value: 1},
			},
		},
	}

	if _, err := m.TelemetryColl.Indexes().CreateMany(ctx, telemetryIndexes); err != nil {
		return err
	}

	return nil
}

func (m *MongoManager) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

func (m *MongoManager) CreateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	res, err := m.UsersColl.InsertOne(ctx, user)
	if err != nil {
		return err
	}

	if id, ok := res.InsertedID.(primitive.ObjectID); ok {
		user.ID = id
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

func (m *MongoManager) GetCase(caseID string) (*models.Case, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var c models.Case
	err := m.CasesColl.FindOne(ctx, bson.M{"_id": caseID}).Decode(&c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (m *MongoManager) GetAllCases() ([]models.Case, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cases []models.Case

	cursor, err := m.CasesColl.Find(
		ctx,
		bson.M{},
		options.Find().SetSort(bson.D{{Key: "order", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {

		}
	}(cursor, ctx)

	err = cursor.All(ctx, &cases)
	return cases, err
}

func (m *MongoManager) GetProgression(
	caseID string,
	userID *primitive.ObjectID,
	teamCode *string,
	matricula *string,
	sessionID *primitive.ObjectID,
) (*models.Progression, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"case_id": caseID}

	if userID != nil {
		filter["user_id"] = *userID
	}
	if teamCode != nil {
		filter["team_code"] = *teamCode
	}
	if matricula != nil && *matricula != "" {
		filter["matricula"] = *matricula
	}
	if sessionID != nil {
		filter["session_id"] = *sessionID
	}

	var p models.Progression
	err := m.ProgressionColl.FindOne(ctx, filter).Decode(&p)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (m *MongoManager) GetUserProgressions(userID primitive.ObjectID) ([]models.Progression, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := m.ProgressionColl.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
		}
	}(cursor, ctx)

	var progressions []models.Progression
	if err := cursor.All(ctx, &progressions); err != nil {
		return nil, err
	}

	return progressions, nil
}

func (m *MongoManager) GetTournamentProgressions(teamCode string) ([]models.Progression, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := m.ProgressionColl.Find(ctx, bson.M{"team_code": teamCode})
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
		}
	}(cursor, ctx)

	var progressions []models.Progression
	if err := cursor.All(ctx, &progressions); err != nil {
		return nil, err
	}

	return progressions, nil
}

func (m *MongoManager) GetActiveTournament() (*models.Tournament, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var t models.Tournament
	err := m.TournamentsColl.FindOne(ctx, bson.M{"active": true}).Decode(&t)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &t, nil
}

func (m *MongoManager) UpsertProgression(p *models.Progression) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	p.UpdatedAt = now
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}

	filter := bson.M{"case_id": p.CaseID}
	if p.UserID != nil {
		filter["user_id"] = *p.UserID
	}
	if p.TeamCode != nil {
		filter["team_code"] = *p.TeamCode
	}
	if p.Matricula != "" {
		filter["matricula"] = p.Matricula
	}
	if p.SessionID != primitive.NilObjectID {
		filter["session_id"] = p.SessionID
	}

	var existing models.Progression
	err := m.ProgressionColl.FindOne(ctx, filter).Decode(&existing)
	if err == nil && existing.Completed {
		p.Completed = true
	}

	_, err = m.ProgressionColl.ReplaceOne(ctx, filter, p, options.Replace().SetUpsert(true))
	return err
}

func (m *MongoManager) ResetProgression(
	caseID string,
	userID *primitive.ObjectID,
	teamCode *string,
	matricula *string,
	startingPuzzle int,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"case_id": caseID}

	if userID != nil {
		filter["user_id"] = *userID
	}
	if teamCode != nil {
		filter["team_code"] = *teamCode
	}
	if matricula != nil && *matricula != "" {
		filter["matricula"] = *matricula
	}

	_, err := m.ProgressionColl.UpdateOne(
		ctx,
		filter,
		bson.M{
			"$set": bson.M{
				"current_puzzle":     startingPuzzle,
				"current_focus":      "none",
				"sql_history":        []models.SQLHistoryItem{},
				"puzzle_checkpoints": bson.M{},
				"active":             true,
				"completed":          false,
				"updated_at":         time.Now(),
			},
		},
	)
	return err
}

func (m *MongoManager) SaveTelemetry(event *models.TelemetryEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	_, err := m.TelemetryColl.InsertOne(ctx, event)
	return err
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

func (m *MongoManager) CountActiveProgressionsByMatricula(teamCode, matricula string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"team_code": teamCode, "matricula": matricula, "active": true}
	return m.ProgressionColl.CountDocuments(ctx, filter)
}
