package main

import (
	"encoding/json"
	"fmt"
	"github.com/Kong/go-pdk"
)

type cacheData struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers"`
	TTL     string              `json:"ttl"`
	Body    interface{}         `json:"body"`
}

// Access handles cache retrieval for GET requests.
func (c Config) Access(kong *pdk.PDK) {
	if !isGetMethod(kong) {
		return
	}

	if hasNoCacheHeader(kong) {
		return
	}

	requestPath, err := kong.Request.GetPathWithQuery()
	if err != nil {
		kong.Log.Err(fmt.Sprintf("Failed to get request path: %v", err))
		return
	}

	cached, err := retrieveFromCache(c, requestPath)
	if err != nil {
		kong.Log.Err(err.Error())
		return
	}

	if cached == nil {
		kong.Log.Info("No data retrieved from Redis")
		return
	}

	sendCachedResponse(kong, cached)
}

// isGetMethod checks if the request method is GET.
func isGetMethod(kong *pdk.PDK) bool {
	method, err := kong.Request.GetMethod()
	if err != nil {
		kong.Log.Err(fmt.Sprintf("Failed to get request method: %v", err))
		return false
	}
	if method != "GET" {
		return false
	}
	return true
}

// hasNoCacheHeader checks for 'no-cache' header.
func hasNoCacheHeader(kong *pdk.PDK) bool {
	cacheControl, err := kong.Request.GetHeader("Cache-Control")
	return err == nil && cacheControl == "no-cache"
}

// retrieveFromCache handles data retrieval from Redis.
func retrieveFromCache(c Config, requestPath string) (*cacheData, error) {
	hashedPath := hashSHA256(requestPath)
	redisClient, err := NewRedisClient(c)

	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %v", err)
	}

	data, err := redisClient.Retrieve(hashedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve data from Redis: %v", err)
	}

	if data == "" {
		return nil, nil
	}

	var cached cacheData
	if err = json.Unmarshal([]byte(data), &cached); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %v", err)
	}

	return &cached, nil
}

// sendCachedResponse sends the cached response.
func sendCachedResponse(kong *pdk.PDK, cached *cacheData) {
	responseBody, err := json.Marshal(cached.Body)
	if err != nil {
		kong.Log.Err(fmt.Sprintf("Failed to marshal response body: %v", err))
		return
	}

	cached.Headers["X-Cache-Status"] = []string{"HIT"}
	cached.Headers["X-Cache-Until"] = []string{cached.TTL}
	kong.Response.Exit(cached.Status, string(responseBody), cached.Headers)
}
