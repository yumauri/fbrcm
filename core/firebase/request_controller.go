package firebase

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultMaxConcurrentRequests = 5
	defaultRequestsPerMinute     = 0
	defaultRateLimitCooldown     = 30 * time.Second
	defaultMaxRequestAttempts    = 5
	defaultBaseRetryDelay        = time.Second
	defaultMaxRetryDelay         = 10 * time.Second
	defaultRetryJitterPercent    = 50
)

// RequestPolicy controls proactive pacing and recovery after rate limiting.
// A zero RequestsPerMinute disables proactive pacing without disabling shared
// 429 cooldowns.
type RequestPolicy struct {
	MaxConcurrentRequests int
	RequestsPerMinute     int
	RateLimitCooldown     time.Duration
	Retry                 RetryPolicy
}

type RetryPolicy struct {
	MaxAttempts   int
	BaseDelay     time.Duration
	MaxDelay      time.Duration
	JitterPercent int
}

func DefaultRequestPolicy() RequestPolicy {
	return RequestPolicy{
		MaxConcurrentRequests: defaultMaxConcurrentRequests,
		RequestsPerMinute:     defaultRequestsPerMinute,
		RateLimitCooldown:     defaultRateLimitCooldown,
		Retry: RetryPolicy{
			MaxAttempts:   defaultMaxRequestAttempts,
			BaseDelay:     defaultBaseRetryDelay,
			MaxDelay:      defaultMaxRetryDelay,
			JitterPercent: defaultRetryJitterPercent,
		},
	}
}

type requestSchedule struct {
	nextAllowed    time.Time
	cooldownUntil  time.Time
	rateLimitCount int
}

// RequestController coordinates pacing and 429 cooldowns across HTTP clients.
// Schedules are isolated by API host and quota consumer.
type RequestController struct {
	mu       sync.Mutex
	policy   RequestPolicy
	schedule map[string]requestSchedule
	slots    chan struct{}
}

func NewRequestController(policy RequestPolicy) *RequestController {
	defaults := DefaultRequestPolicy()
	if policy.MaxConcurrentRequests <= 0 {
		policy.MaxConcurrentRequests = defaults.MaxConcurrentRequests
	}
	if policy.RequestsPerMinute < 0 {
		policy.RequestsPerMinute = 0
	}
	if policy.RateLimitCooldown <= 0 {
		policy.RateLimitCooldown = defaultRateLimitCooldown
	}
	if policy.Retry == (RetryPolicy{}) {
		policy.Retry = defaults.Retry
	} else {
		if policy.Retry.MaxAttempts <= 0 {
			policy.Retry.MaxAttempts = defaults.Retry.MaxAttempts
		}
		if policy.Retry.BaseDelay <= 0 {
			policy.Retry.BaseDelay = defaults.Retry.BaseDelay
		}
		if policy.Retry.MaxDelay < policy.Retry.BaseDelay {
			policy.Retry.MaxDelay = max(defaults.Retry.MaxDelay, policy.Retry.BaseDelay)
		}
		if policy.Retry.JitterPercent < 0 || policy.Retry.JitterPercent > 100 {
			policy.Retry.JitterPercent = defaults.Retry.JitterPercent
		}
	}
	return &RequestController{
		policy:   policy,
		schedule: make(map[string]requestSchedule),
		slots:    make(chan struct{}, policy.MaxConcurrentRequests),
	}
}

func (c *RequestController) Policy() RequestPolicy {
	if c == nil {
		return DefaultRequestPolicy()
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.policy
}

func (c *RequestController) wait(ctx context.Context, key string) error {
	if c == nil {
		return nil
	}
	for {
		c.mu.Lock()
		now := time.Now()
		state := c.schedule[key]
		readyAt := state.nextAllowed
		if state.cooldownUntil.After(readyAt) {
			readyAt = state.cooldownUntil
		}
		if !readyAt.After(now) {
			if c.policy.RequestsPerMinute > 0 {
				interval := time.Minute / time.Duration(c.policy.RequestsPerMinute)
				state.nextAllowed = now.Add(interval)
				c.schedule[key] = state
			}
			c.mu.Unlock()
			return nil
		}
		c.mu.Unlock()

		if err := sleepContext(ctx, time.Until(readyAt)); err != nil {
			return err
		}
	}
}

func (c *RequestController) recordRateLimit(key string, resp *http.Response) time.Duration {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state := c.schedule[key]
	state.rateLimitCount++
	delay, provided := retryAfterDelay(resp)
	if !provided {
		delay = linearRateLimitDelay(c.policy.RateLimitCooldown, state.rateLimitCount)
	}
	until := time.Now().Add(delay)
	if until.After(state.cooldownUntil) {
		state.cooldownUntil = until
	}
	c.schedule[key] = state
	return delay
}

func linearRateLimitDelay(base time.Duration, count int) time.Duration {
	if count <= 1 {
		return base
	}
	if base > time.Duration(1<<63-1)/time.Duration(count) {
		return time.Duration(1<<63 - 1)
	}
	return base * time.Duration(count)
}

func (c *RequestController) recordNonRateLimit(key string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	state := c.schedule[key]
	state.rateLimitCount = 0
	c.schedule[key] = state
}

func (c *RequestController) acquire(ctx context.Context) error {
	if c == nil {
		return nil
	}
	select {
	case c.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *RequestController) release() {
	if c == nil {
		return
	}
	select {
	case <-c.slots:
	default:
	}
}

func requestScheduleKey(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	host := strings.ToLower(strings.TrimSpace(req.URL.Hostname()))
	consumer := strings.ToLower(strings.TrimSpace(req.Header.Get("X-Goog-User-Project")))
	return host + "\x00" + consumer
}

type requestControllerContextKey struct{}

// WithRequestController binds one shared controller to clients created with
// the returned context.
func WithRequestController(ctx context.Context, controller *RequestController) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if controller == nil {
		controller = defaultRequestController
	}
	return context.WithValue(ctx, requestControllerContextKey{}, controller)
}

func requestControllerFromContext(ctx context.Context) *RequestController {
	if ctx != nil {
		if controller, ok := ctx.Value(requestControllerContextKey{}).(*RequestController); ok && controller != nil {
			return controller
		}
	}
	return defaultRequestController
}

var defaultRequestController = NewRequestController(DefaultRequestPolicy())
