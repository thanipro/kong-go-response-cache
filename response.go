package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/Kong/go-pdk"
	"net/http"
	"time"
)

// Response handles caching the response.
func (c Config) Response(kong *pdk.PDK) {
	responseStatus, err := getResponseStatus(kong)
	if err != nil {
		logError(kong, "Failed to get response status", err)
		return
	}

	if responseStatus > http.StatusAccepted {
		return
	}

	data, err := getResponseData(kong)
	if err != nil {
		logError(kong, "Failed to get response data", err)
		return
	}

	path, err := kong.Request.GetPathWithQuery()
	if err != nil {
		logError(kong, "Failed to get request path", err)
		return
	}

	redisClient, err := NewRedisClient(c)
	if err != nil {
		logError(kong, "Failed to create Redis client", err)
		return
	}

	hashedPath := hashSHA256(path)
	ttl := time.Duration(c.CachedTtl) * time.Second
	expiresAt := time.Now().Add(ttl).Format(time.RFC1123)

	headers, err := kong.ServiceResponse.GetHeaders(100)
	if err != nil {
		logError(kong, "Failed to get response headers", err)
		return
	}

	cacheData := createCacheData(data, responseStatus, headers, expiresAt)
	cacheDataJSON, err := json.Marshal(cacheData)
	if err != nil {
		logError(kong, "Error marshaling cache data to JSON", err)
		return
	}

	if err := storeInCache(redisClient, hashedPath, string(cacheDataJSON)); err != nil {
		logError(kong, "Error storing value in cache", err)
		return
	}

	addCacheHeaders(kong, hashedPath, expiresAt)
}

func getResponseStatus(kong *pdk.PDK) (int, error) {
	status, err := kong.ServiceResponse.GetStatus()
	if err != nil {
		return 0, err
	}
	return status, nil
}

func getResponseData(kong *pdk.PDK) (interface{}, error) {
	responseStr, err := kong.ServiceResponse.GetRawBody()
	if err != nil {
		return nil, err
	}

	var data interface{}
	if err := json.Unmarshal([]byte(responseStr), &data); err != nil {
		return nil, err
	}

	return data, nil
}

func hashSHA256(path string) string {
	h := sha256.New()
	h.Write([]byte(path))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func createCacheData(data interface{}, status int, headers map[string][]string, expiresAt string) map[string]interface{} {
	return map[string]interface{}{
		"Body":    data,
		"Status":  status,
		"Headers": headers,
		"TTL":     expiresAt,
	}
}

func storeInCache(redisClient *RedisClient, key, value string) error {
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return redisClient.Store(key, value)
}

func addCacheHeaders(kong *pdk.PDK, cacheKey, expiresAt string) {
	headers := map[string]string{
		"X-Cache-Status": "MISS",
		"X-Cache-Key":    cacheKey,
		"X-Cache-Until":  expiresAt,
	}

	for key, value := range headers {
		if err := kong.Response.AddHeader(key, value); err != nil {
			logError(kong, fmt.Sprintf("Error adding header %s", key), err)
		}
	}
}

func logError(kong *pdk.PDK, message string, err error) {
	kong.Log.Err(fmt.Sprintf("%s: %v", message, err))
}
