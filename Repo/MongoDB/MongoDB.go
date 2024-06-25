package MongoDB

import (
	"context"
	logger "github.com/skwizi4/lib/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"main.go/internal/domain"
)

// CRUD methods - Create, Read, Update, Delete

func InitMongo(uri, databaseName, collectionName string) (*MongoDB, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, err
	}
	collection := client.Database(databaseName).Collection(collectionName)

	return &MongoDB{
		Client:         client,
		DatabaseName:   databaseName,
		CollectionName: collectionName,
		Logger:         logger.InitLogger(),
		Collection:     collection,
	}, nil
}
func (m MongoDB) Create(User domain.User) error {
	_, err := m.Collection.InsertOne(context.Background(), User)
	if err != nil {
		return err
	}
	return nil
}
func (m MongoDB) Delete(id int64) error {
	_, err := m.Collection.DeleteOne(context.Background(), bson.D{{Key: "_id", Value: id}})
	if err != nil {
		return err
	}
	return nil

}

func (m MongoDB) Update(id int64, updatedUser domain.User) error {
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{
		{"$set", bson.D{
			{"name", updatedUser.Name},
			{"password", updatedUser.Password},
			{"spread_sheet_id", updatedUser.SpreadSheetID},
			{"frequency_of_notifications", updatedUser.FrequencyOfNotifications},
			{"is_send_notification", updatedUser.IsSendNotification},
			{"chat_id", updatedUser.ChatID},
		}},
	}
	_, err := m.Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (m MongoDB) Get(id int64) (*domain.User, error) {
	var user domain.User
	filter := bson.D{{Key: "_id", Value: id}}
	if err := m.Collection.FindOne(context.Background(), filter).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}
