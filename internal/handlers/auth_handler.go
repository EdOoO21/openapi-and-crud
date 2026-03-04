package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/EdOoO21/openapi-and-crud/internal/service"
)

type AuthHandler struct {
	svc *service.AuthService
}

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid json", nil)
		return
	}
	if req.Role == "" {
		req.Role = "USER"
	}
	u, err := h.svc.Register(r.Context(), req.Email, req.Password, req.Role)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"id": u.ID.String(), "email": u.Email, "role": u.Role})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid json", nil)
		return
	}
	tokens, u, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "TOKEN_INVALID", "invalid credentials", nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"user": map[string]interface{}{
			"id":    u.ID.String(),
			"email": u.Email,
			"role":  u.Role,
		},
	})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid json", nil)
		return
	}
	tokens, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "REFRESH_TOKEN_INVALID", "invalid refresh token", nil)
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}
