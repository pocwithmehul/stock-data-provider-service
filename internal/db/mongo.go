package db

import (
	"context"

	"github.com/pocwithmehul/common-go-lib/pkg/mongo"
)

func ConnectMongo(ctx context.Context, uri, database string) (*mongo.Client, error) {
	return mongo.NewClient(ctx, mongo.Config{
		URI:      uri,
		Database: database,
	})
}
