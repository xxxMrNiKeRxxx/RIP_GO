package handler

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"metoda/internal/app/repository"
	"metoda/internal/app/serializer"
	"net/http"
)

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

	u, err := h.Repository.SignIn(j)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	ctx.JSON(http.StatusOK, serializer.UserToJSON(u))
}

func (h *Handler) APISignOut(ctx *gin.Context) {
	repository.SignOut()
	ctx.JSON(http.StatusOK, gin.H{"status": "signed_out"})
}
