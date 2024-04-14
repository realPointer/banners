package redis

import "time"

type Option func(*Redis)

func DialTimeout(timeout time.Duration) Option {
	return func(r *Redis) {
		r.Client.Options().DialTimeout = timeout
	}
}
