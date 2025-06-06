package middleware_utils

var MiddlewareOptions = []string{"RecoverMiddleware", "LoggingMiddleware", "JwtMiddleware", "LoggingMiddleware", "RateLimiterMiddleware"}

var MiddlewareOptionTemplatePathMap = map[string]string{
	"LoggingMiddleware":     LoggingMiddlewareTemplatePath,
	"JwtMiddleware":         JwtMiddlewareTemplatePath,
	"LoggerMiddleware":      LoggingMiddlewareTemplatePath,
	"RateLimiterMiddleware": RateLimiterMiddlewareTemplatePath,
}
