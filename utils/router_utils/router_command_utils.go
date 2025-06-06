package router_utils

type RouterData struct {
	LoggerMiddleware      bool
	RecoverMiddleware     bool
	RateLimiterMiddleware bool
	AllowOriginMiddleware bool
	ImplementMiddlewares  bool
}
