package db

import (
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
)

type CollectionProvider struct {
	mu         sync.RWMutex
	collection *mongo.Collection
}

func (p *CollectionProvider) Set(collection *mongo.Collection) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.collection = collection
}

func (p *CollectionProvider) Get() *mongo.Collection {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.collection
}

func (p *CollectionProvider) Ready() bool {
	return p.Get() != nil
}
