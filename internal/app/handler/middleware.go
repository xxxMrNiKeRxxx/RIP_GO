package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"metoda/internal/app/auth"
)

const (
	// AuthCookieName имя cookie с JWT (сессия).
	AuthCookieName = "auth_token"
	ctxUserID      = "auth_user_id"
	ctxIsModerator = "auth_is_moderator"
)

func bearerPrefix() string { return "Bearer " }

// extractJWT из cookie или заголовка Authorization.
func extractJWT(r *http.Request) string {
	if c, err := r.Cookie(AuthCookieName); err == nil && c.Value != "" {
		return c.Value
	}
	h := r.Header.Get("Authorization")
	p := bearerPrefix()
	if h != "" && strings.HasPrefix(h, p) {
		return strings.TrimPrefix(h, p)
	}
	return ""
}

// AuthMiddleware проверяет JWT и кладёт user_id / is_moderator в контекст Gin.
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractJWT(c.Request)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		// ✅ Исправлено: IsTokenBlacklisted → IsBlacklisted
		blacklisted, err := h.Repository.IsBlacklisted(context.Background(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"status": "error", "description": err.Error()})
			return
		}
		if blacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		claims, err := auth.ParseAndValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		uid, err := auth.UserIDFromClaims(claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "unauthorized"})
			return
		}

		// Сохраняем userID в репозиторий (для методов, которые используют r.GetUserID())
		h.Repository.SetUserID(int(uid))

		// Сохраняем в контекст запроса (для хелперов authUserIDUint, isModeratorFromCtx)
		c.Set(ctxUserID, uid)
		c.Set(ctxIsModerator, auth.IsModeratorFromClaims(claims))
		c.Next()
	}
}

// RequireModerator разрешает только пользователям с is_moderator (после AuthMiddleware).
func (h *Handler) RequireModerator() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(ctxIsModerator)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "forbidden"})
			return
		}
		isMod, ok := v.(bool)
		if !ok || !isMod {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"status": "forbidden"})
			return
		}
		c.Next()
	}
}

// ─── Helpers ───────────────────────────────────────────────────────────────

func authUserIDUint(ctx *gin.Context) (uint, error) {
	v, ok := ctx.Get(ctxUserID)
	if !ok {
		return 0, errors.New("нет user_id в контексте")
	}
	id, ok := v.(uint)
	if !ok || id == 0 {
		return 0, errors.New("неверный user_id")
	}
	return id, nil
}

func isModeratorFromCtx(ctx *gin.Context) bool {
	v, ok := ctx.Get(ctxIsModerator)
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}