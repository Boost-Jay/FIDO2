package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Simple in-memory rate limiter
type memoryStore struct {
	sync.Mutex
	requests map[string][]time.Time
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		requests: make(map[string][]time.Time),
	}
}

// RateLimit middleware limits the number of requests per IP
func RateLimit(limit int, duration time.Duration) gin.HandlerFunc {
	store := newMemoryStore()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		store.Lock()
		defer store.Unlock()

		// Clean up old requests
		now := time.Now()
		if _, exists := store.requests[ip]; exists {
			var validRequests []time.Time
			for _, timestamp := range store.requests[ip] {
				if now.Sub(timestamp) <= duration {
					validRequests = append(validRequests, timestamp)
				}
			}
			store.requests[ip] = validRequests
		}

		// Check if limit exceeded
		if len(store.requests[ip]) >= limit {
			c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(duration).Unix(), 10))
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		// Add current request
		store.requests[ip] = append(store.requests[ip], now)

		// Set headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(limit-len(store.requests[ip])))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(duration).Unix(), 10))

		c.Next()
	}
}

// RedisBased rate limiter (recommended for production)
func RedisRateLimit(redisClient *redis.Client, limit int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		ip := c.ClientIP()
		key := "rate_limit:" + ip

		// Check current count
		count, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		// If first request, set the key with expiry
		if err == redis.Nil {
			redisClient.Set(ctx, key, 1, duration)
			count = 1
		} else if count < limit {
			// Increment the counter
			redisClient.Incr(ctx, key)
			count++
		} else {
			// Rate limit exceeded
			c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Header("X-RateLimit-Remaining", "0")
			remaining, _ := redisClient.TTL(ctx, key).Result()
			c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(remaining).Unix(), 10))
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}

		// Set headers
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(limit-count))
		remaining, _ := redisClient.TTL(ctx, key).Result()
		c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(remaining).Unix(), 10))

		c.Next()
	}
}