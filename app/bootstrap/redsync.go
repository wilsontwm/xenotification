package bootstrap

import (
	"time"

	"xenotification/app/env"

	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
)

func (bs *Bootstrap) initRedsync() *Bootstrap {

	bs.Redsync = redsync.New([]redsync.Pool{
		getRedisPool(),
	})

	return bs
}

func getRedisPool() *redis.Pool {
	url := env.Config.Redis.Host
	password := env.Config.Redis.Password
	database := 0

	redisPool := &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", url, redis.DialPassword(password), redis.DialDatabase(database))
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	if err := ping(redisPool.Get()); err != nil {
		panic(err)
	}

	return redisPool
}

func ping(c redis.Conn) error {
	pong, err := c.Do("PING")
	if err != nil {
		return err
	}

	_, err = redis.String(pong, err)
	if err != nil {
		return err
	}

	return nil
}
