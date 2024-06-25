package MongoDB

import (
	logger "github.com/skwizi4/lib/logs"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDB struct {
	Client         *mongo.Client
	Logger         logger.GoLogger
	DatabaseName   string
	CollectionName string
	Collection     *mongo.Collection
}
