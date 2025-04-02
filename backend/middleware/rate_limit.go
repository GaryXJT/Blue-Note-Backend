package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	ips   map[string][]time.Time
	mutex sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		ips: make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) RateLimit(maxRequests int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		rl.mutex.Lock()
		defer rl.mutex.Unlock()
		
		// 清理过期的请求记录
		now := time.Now()
		cutoff := now.Add(-duration)
		
		if _, exists := rl.ips[ip]; exists {
			var validRequests []time.Time
			for _, t := range rl.ips[ip] {
				if t.After(cutoff) {
					validRequests = append(validRequests, t)
				}
			}
			rl.ips[ip] = validRequests
		}
		
		// 检查请求数量
		if len(rl.ips[ip]) >= maxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    42900,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		
		// 记录本次请求
		rl.ips[ip] = append(rl.ips[ip], now)
		
		c.Next()
	}
} 