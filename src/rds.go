package main

import (
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

var initPublisherOnce sync.Once
var initSubscriberOnce sync.Once
var initCacheOnce sync.Once
var getOptionsOnce sync.Once

var publisher *redis.Client
var subscriber *redis.Client
var cache *redis.Client
var options *redis.Options

func getOptions() *redis.Options {
	getOptionsOnce.Do(func() {
		options = &redis.Options{
			Addr:     os.Getenv("REDIS_ENDPOINT"),
			Password: os.Getenv("REDIS_SECRET"),
		}
	})
	return options
}

func publish(roomName *string, message []byte) {
	initPublisherOnce.Do(func() {
		publisher = redis.NewClient(getOptions())
	})
	log.Debugf("publishing to room %s", *roomName)
	cmd := publisher.Publish(ctx, *roomName, message)
	if _, err := cmd.Result(); err != nil {
		log.Errorf("Error while publishing %s", err)
	}
}

func subscribe(roomName *string) *redis.PubSub {
	initSubscriberOnce.Do(func() {
		subscriber = redis.NewClient(getOptions())
	})
	log.Debugf("subscribing to room %s", *roomName)
	return subscriber.Subscribe(ctx, *roomName)
}

func setCache(key string, path string, value interface{}) {
	initCacheOnce.Do(func() {
		cache = redis.NewClient(getOptions())
	})
	cmd := cache.JSONSet(ctx, key, path, value)
	val, err := cmd.Result()
	if err != nil {
		log.Errorf("Error while setting cache %s", err)
	}
	log.Debugf("Set cache for %s on %s with %s", key, path, val)
}

func removeCache(key string, path string) {
	initCacheOnce.Do(func() {
		cache = redis.NewClient(getOptions())
	})
	cmd := cache.JSONDel(ctx, key, path)
	val, err := cmd.Result()
	if err != nil {
		log.Errorf("Error while removing cache %s", err)
	}
	log.Debugf("Removed cache for %s on %s with %v", key, path, val)
}

func getCache(key string, path string) *string {
	initCacheOnce.Do(func() {
		cache = redis.NewClient(getOptions())
	})
	cmd := cache.JSONGet(ctx, key, path)
	val, err := cmd.Result()
	if err != nil {
		log.Errorf("Error while getting cache %s", err)
	}
	log.Debugf("Got cache for %s on %s with %v", key, path, val)
	return &val
}
