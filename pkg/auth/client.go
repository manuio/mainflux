// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"

	"github.com/MainfluxLabs/mainflux"
	"github.com/go-redis/redis/v8"
)

// Client represents Auth cache.
type Client interface {
	Identify(ctx context.Context, thingKey string) (string, error)
	ConnectionIDS(ctx context.Context, thingKey string) (string, string, error)
}

const (
	chanPrefix = "channel"
	keyPrefix  = "thing_key"
)

type client struct {
	redisClient *redis.Client
	things      mainflux.ThingsServiceClient
}

// New returns redis channel cache implementation.
func New(redisClient *redis.Client, things mainflux.ThingsServiceClient) Client {
	return client{
		redisClient: redisClient,
		things:      things,
	}
}

func (c client) Identify(ctx context.Context, thingKey string) (string, error) {
	tkey := keyPrefix + ":" + thingKey
	thingID, err := c.redisClient.Get(ctx, tkey).Result()
	if err != nil {
		t := &mainflux.Token{
			Value: string(thingKey),
		}

		thid, err := c.things.Identify(context.TODO(), t)
		if err != nil {
			return "", err
		}
		return thid.GetValue(), nil
	}
	return thingID, nil
}

func (c client) ConnectionIDS(ctx context.Context, thingKey string) (string, string, error) {
	req := &mainflux.ConnByKeyReq{
		Key: thingKey,
	}

	conn, err := c.things.GetConnByKey(context.TODO(), req)
	if err != nil {
		return "", "", err
	}

	return conn.ThingID, conn.ChannelID, nil
}
