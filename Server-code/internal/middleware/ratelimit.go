package middleware

import (
	"sync"
	"time"

	"labelpro-server/internal/config"
	"labelpro-server/internal/utils"

	"github.com/gin-gonic/gin"
)

// rateLimiter 按 ClientIP 做固定窗口限流。
// 注意：必须使用"窗口起始时间"而非"最近请求时间"来判定窗口过期。
// 若以最近请求时间刷新过期点，持续访问会导致计数永不重置，
// 累计超过阈值后形成"限流死锁"（429 的重试请求同样刷新时间戳，永远无法自愈）。
//
// 限流参数每次请求均从 config.GetActive() 动态读取：
// 管理员在系统设置（PUT /api/v1/system/config）中修改后无需重启即可生效。
type rateLimiter struct {
	mu           sync.Mutex
	visitors     map[string]*visitor
	lastSweep    time.Time
}

type visitor struct {
	windowStart time.Time // 当前固定窗口起点，到期即重置计数
	apiCount    int       // 普通API窗口内请求数
	loginCount  int       // 登录接口窗口内请求数（独立计数，避免正常使用API误伤登录）
	bannedUntil time.Time // 触发限流后的封禁截止时间
}

var limiter = &rateLimiter{
	visitors:  make(map[string]*visitor),
	lastSweep: time.Now(),
}

const (
	windowDuration = time.Minute
	sweepInterval  = 10 * time.Minute // 惰性清理过期访客，防止map无限增长
)

func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 动态获取当前生效配置，支持管理员热更新
		cfg := config.GetActive()
		if cfg == nil || !cfg.RateLimit.Enabled {
			c.Next()
			return
		}
		rl := cfg.RateLimit

		banDuration := time.Duration(rl.BanDurationSeconds) * time.Second
		if banDuration <= 0 {
			banDuration = 60 * time.Second
		}

		key := c.ClientIP()
		isLogin := c.Request.URL.Path == "/api/v1/auth/login"

		limiter.mu.Lock()
		now := time.Now()

		// 惰性清理长期不活跃的访客记录
		if now.Sub(limiter.lastSweep) > sweepInterval {
			for k, v := range limiter.visitors {
				if now.Sub(v.windowStart) > sweepInterval && now.After(v.bannedUntil) {
					delete(limiter.visitors, k)
				}
			}
			limiter.lastSweep = now
		}

		v, exists := limiter.visitors[key]
		if !exists {
			v = &visitor{windowStart: now}
			limiter.visitors[key] = v
		}

		// 窗口到期：重置计数（与请求行为无关，到点必重置）
		if now.Sub(v.windowStart) >= windowDuration {
			v.windowStart = now
			v.apiCount = 0
			v.loginCount = 0
		}

		// 封禁期内：直接拒绝，且不刷新任何计数/时间戳
		if now.Before(v.bannedUntil) {
			limiter.mu.Unlock()
			utils.TooManyRequests(c, "操作过于频繁，请稍后重试")
			c.Abort()
			return
		}

		var over bool
		if isLogin {
			v.loginCount++
			// 阈值<=0 视为不限制，避免管理员误配置导致所有登录被拒
			over = rl.LoginPerMinute > 0 && v.loginCount > rl.LoginPerMinute
		} else {
			v.apiCount++
			over = rl.APIPerMinute > 0 && v.apiCount > rl.APIPerMinute
		}

		// 超限：按配置封禁一段时间，期满自动恢复
		if over {
			v.bannedUntil = now.Add(banDuration)
		}
		limiter.mu.Unlock()

		if over {
			utils.TooManyRequests(c, "操作过于频繁，请稍后重试")
			c.Abort()
			return
		}

		c.Next()
	}
}

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		c.Next()
	}
}
