package main

type Config struct {
	CachedTtl     int    `json:"cached_ttl"`
	RedisHost     string `json:"redis_host"`
	RedisPort     string `json:"redis_port"`
	RedisPassword string `json:"redis_password"`
}
