package main

import (
	"context"
	"testing"
)

func TestBuild(t *testing.T) {
	ctx := context.Background()
	cache, err := NewDataCache(ctx, "holy-diver-297719")
	if err != nil {
		t.Fatal(err)
	}
	err = cache.Warmup(ctx)
	if err != nil {
		t.Fatal(err)
	}
	data := cache.Data()
	t.Log(len(data))
	t.Log(data[0])
}
