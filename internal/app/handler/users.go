package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"metoda/internal/app/repository"
	"metoda/internal/app/serializer"
)

// ─── Auth: Sign Up ─────────────────────────────────────────────────────────

// APISignUp Регистрация нового пользователя
// @Summary Регистрация
// @Tags auth
// @Accept json
// @Produce json
// @Param user body serializer.UserJSON true "Данные пользователя"
// @Success 201 {object} serializer.UserJSON
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string "Пользователь уже существует"
// @Failure 500 {object} map[string]string
// @Router /api/users/signup [post]
func (h *Handler) APISignUp(ctx *gin.Context) {
	var j serializer.UserJSON
	if err := ctx.ShouldBindJSON(&j); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	if j.Login == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("поле login обязательно"))
		return
	}
	if j.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("поле password обязательно"))
		return
	}

	u, err := h.Repository.CreateUser(j)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			h.errorHandler(ctx, http.StatusConflict, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.Header("Location", fmt.Sprintf("/api/users/%d", u.ID))
	ctx.JSON(http.StatusCreated, serializer.UserToJSON(u))
}

// ─── Auth: Sign In ─────────────────────────────────────────────────────────

// APISignIn Вход в систему (получение токена)
// @Summary Вход
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body serializer.UserJSON true "Логин и пароль"
// @Success 200 {object} map[string]interface{} "user + token"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string "Неверный логин или пароль"
// @Failure 500 {object} map[string]string
// @Router /api/users/signin [post]
func (h *Handler) APISignIn(ctx *gin.Context) {
	var j serializer.UserJSON
	if err := ctx.ShouldBindJSON(&j); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	if j.Login == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("поле login обязательно"))
		return
	}
	if j.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("поле password обязательно"))
		return
	}

	// ✅ Используем SignInWithToken для получения JWT
	u, token, err := h.Repository.SignInWithToken(j)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	// ✅ Устанавливаем токен в куку (HttpOnly)
	cookieName := "auth_token"
	maxAge := int(24 * time.Hour / time.Second)
	// secure := os.Getenv("APP_MODE") == "production" // раскомментируй для продакшена
	secure := false
	ctx.SetCookie(cookieName, token, maxAge, "/", "", secure, true)

	ctx.JSON(http.StatusOK, gin.H{
		"user":  serializer.UserToJSON(u),
		"token": token,
	})
}

// ─── Auth: Sign Out ────────────────────────────────────────────────────────

// APISignOut Выход из системы (инвалидация токена)
// @Summary Выход
// @Tags auth
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "status: signed_out"
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/signout [post]
func (h *Handler) APISignOut(ctx *gin.Context) {
	// ✅ Извлекаем токен из куки или заголовка
	token, err := ctx.Cookie("auth_token")
	if err != nil {
		authHeader := ctx.GetHeader("Authorization")
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}
	}

	// ✅ Добавляем токен в блэклист (Redis)
	if token != "" {
		_ = h.Repository.AddToBlacklist(ctx.Request.Context(), token, 24*time.Hour)
	}

	// ✅ ИСПРАВЛЕНО: вызываем метод экземпляра, а не глобальную функцию
	h.Repository.SignOut()

	// ✅ Удаляем куку на клиенте
	ctx.SetCookie("auth_token", "", -1, "/", "", false, true)

	ctx.JSON(http.StatusOK, gin.H{"status": "signed_out"})
}