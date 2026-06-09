package db

import (
	"sync"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func TestCollectionProvider_InitiallyNotReady(t *testing.T) {
	p := &CollectionProvider{}

	if p.Ready() {
		t.Error("expected not ready initially")
	}
	if p.Get() != nil {
		t.Error("expected Get to return nil initially")
	}
}

func TestCollectionProvider_SetAndGet(t *testing.T) {
	p := &CollectionProvider{}
	col := &mongo.Collection{}

	p.Set(col)

	if !p.Ready() {
		t.Error("expected ready after Set")
	}
	if p.Get() != col {
		t.Error("expected Get to return the collection that was set")
	}
}

func TestCollectionProvider_SetNilMakesNotReady(t *testing.T) {
	p := &CollectionProvider{}
	p.Set(&mongo.Collection{})
	p.Set(nil)

	if p.Ready() {
		t.Error("expected not ready after setting nil")
	}
}

func TestCollectionProvider_ConcurrentAccess(t *testing.T) {
	p := &CollectionProvider{}
	col := &mongo.Collection{}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			p.Set(col)
		}()
		go func() {
			defer wg.Done()
			_ = p.Get()
			_ = p.Ready()
		}()
	}
	wg.Wait()
}
