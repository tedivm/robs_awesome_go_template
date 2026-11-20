package middleware

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Context key for authenticated user claims.
type ctxKey struct{}

var ctxKeyUser = ctxKey{}

// defaultPublicPaths are routes that should bypass authentication by default.
var defaultPublicPaths = []string{"/health", "/metrics", "/static/", "/openapi.json", "/openapi.yaml", "/api/v1/auth/login"}

// AuthOpts configures the Auth middleware.
type AuthOpts struct {
	// Secret is the JWT signing secret. If empty, authentication is disabled.
	Secret string
	// PublicPaths are routes that bypass authentication for ALL HTTP methods.
	// Defaults include: /health, /metrics, /static/, /openapi.json, /openapi.yaml, /api/v1/auth/login.
	PublicPaths []string
	// PublicReadPaths are routes that bypass authentication for read-only methods
	// (GET, HEAD, OPTIONS) only. Write methods (POST, PUT, PATCH, DELETE) still require auth.
	// Use this for APIs where reads are public but writes require authentication.
	PublicReadPaths []string
}

func isPublicPath(path string, publicPaths []string) bool {
	for _, p := range publicPaths {
		// Exact match always works. For prefix matches, require the pattern
		// to be longer than "/" so that "/" alone doesn't match every request.
		if path == p || (len(p) > 1 && strings.HasPrefix(path, p)) {
			return true
		}
	}
	return false
}

// GetClaimsFromContext extracts JWT claims from the request context.
func GetClaimsFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(ctxKeyUser).(jwt.MapClaims)
	return claims, ok
}

// Auth returns middleware that validates JWT tokens from the Authorization header.
// If opts.Secret is empty, authentication is disabled and all requests pass through.
// Public paths (defaults + opts.PublicPaths) bypass authentication for all methods.
// Public read paths (opts.PublicReadPaths) bypass authentication for GET/HEAD/OPTIONS only.
func Auth(opts AuthOpts) func(http.Handler) http.Handler {
	allPublic := append(defaultPublicPaths, opts.PublicPaths...)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth if secret is not configured
			if opts.Secret == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip auth for fully public paths
			if isPublicPath(r.URL.Path, allPublic) {
				next.ServeHTTP(w, r)
				return
			}

			// Skip auth for read-only methods on public read paths
			if isPublicPath(r.URL.Path, opts.PublicReadPaths) && isReadMethod(r.Method) {
				next.ServeHTTP(w, r)
				return
			}

			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}

			claims := jwt.MapClaims{}
			token, err := jwt.ParseWithClaims(parts[1], claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(opts.Secret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyUser, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isReadMethod returns true for HTTP methods considered read-only.
func isReadMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

// CORSCfg holds the configuration needed for CORS middleware.
type CORSCfg struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// CORS returns middleware that sets CORS headers based on configuration.
func CORS(cfg CORSCfg) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			allowed := len(cfg.AllowedOrigins) > 0 && (slices.Contains(cfg.AllowedOrigins, "*") || slices.Contains(cfg.AllowedOrigins, origin))

			if allowed {
				if len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
					w.Header().Set("Access-Control-Allow-Origin", "*")
				} else if origin != "" {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
