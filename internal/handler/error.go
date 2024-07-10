package handler

import "errors"

var (
	errUnauthorized        = errors.New("unauthorized")
	errGetClaimsFromToken  = errors.New("failed to get claims from token")
	errGetUserIDFromClaims = errors.New("failed to get userID from claims")
)

type ErrorResponse struct {
	Error string `json:"error"`
}
