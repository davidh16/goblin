package middleware_utils

var middlewareOptions = []string{"LoggingMiddleware", "JwtMiddleware", "LoggingMiddleware", "RateLimiterMiddleware"}

var middlewareOptionTemplatePathMap = map[string]string{
	"LoggingMiddleware":     LoggingMiddlewareTemplatePath,
	"JwtMiddleware":         JwtMiddlewareTemplatePath,
	"LoggerMiddleware":      LoggingMiddlewareTemplatePath,
	"RateLimiterMiddleware": RateLimiterMiddlewareTemplatePath,
}
