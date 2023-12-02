package repository

import (
	"RIpPeakBack/internal/app/ds"
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"time"
)

type Redis struct {
	client *redis.Client
}

func NewRedis() Redis {
	r := Connect()

	return Redis{
		client: r,
	}
}

func Connect() *redis.Client {
	r := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	_, err := r.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}

	return r
}

func (s *Redis) Add(session ds.Session) error {
	if session.Token == "" {
		return ds.ErrInvalidToken
	}

	jsonData, err := json.Marshal(ds.SessionContext{
		UserID: session.UserID,
		Role:   session.Role,
	})
	if err != nil {
		return ds.ErrInvalidToken
	}

	duration := session.ExpiresAt.Sub(time.Now())
	err = s.client.Set(context.TODO(), session.Token, jsonData, duration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *Redis) SessionExists(token string) (ds.SessionContext, error) {
	if token == "" {
		return ds.SessionContext{}, ds.ErrInvalidToken
	}

	r, err := s.client.Get(context.Background(), token).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ds.SessionContext{}, ds.ErrNotFound
		}
		return ds.SessionContext{}, err
	}

	var sc ds.SessionContext
	err = json.Unmarshal([]byte(r), &sc)
	if err != nil {
		return ds.SessionContext{}, err
	}

	return sc, nil
}

func (s *Redis) DeleteByToken(token string) error {
	if token == "" {
		return ds.ErrInvalidToken
	}
	err := s.client.Del(context.Background(), token).Err()
	if err != nil {
		return err
	}
	return nil
}
