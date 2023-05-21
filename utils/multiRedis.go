package utils

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type MultiRedis struct {
	conns           map[string]*redis.Client
	defaultHost     string
	defaultPassword string

	errors []error
}

func NewMultiRedis() *MultiRedis {
	return &MultiRedis{conns: make(map[string]*redis.Client)}
}

func (r *MultiRedis) WithDefault(host string, password string) *MultiRedis {
	r.defaultHost = host
	r.defaultPassword = password
	return r
}

func (r *MultiRedis) Add(alias string, db int) *MultiRedis {
	cli := redis.NewClient(&redis.Options{
		Addr:     r.defaultHost,
		Password: r.defaultPassword,
		DB:       db,
	})
	if err := cli.Ping(context.Background()).Err(); err != nil {
		r.errors = append(r.errors, err)
	} else {
		r.conns[alias] = cli
	}

	return r
}

func (r *MultiRedis) Get(alias string) *redis.Client {
	if cli, ok := r.conns[alias]; ok {
		return cli
	}
	return nil
}

func (r *MultiRedis) Errors() []error {
	return r.errors
}
