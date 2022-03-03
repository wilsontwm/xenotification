package bootstrap

import (
	"context"
	"xenotification/app/repository"

	"github.com/go-redsync/redsync"
	"go.mongodb.org/mongo-driver/mongo"
)

// Bootstrap :
type Bootstrap struct {
	MongoDB    *mongo.Client
	Repository *repository.Repository
	Redsync    *redsync.Redsync
}

// New :
func New() *Bootstrap {

	bs := new(Bootstrap)
	bs.initMongoDB()
	bs.initJaeger()
	bs.initRedsync()
	// go bs.initCron()

	repo := repository.New(context.Background(), bs.MongoDB)

	bs.Repository = repo

	return bs
}
