package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/migomi3/internal/auth"
	"github.com/migomi3/internal/database"
)

func (cfg *apiConfig) healthEndpointHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	html := fmt.Sprintf(`
<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Unauthorized access", errors.New("user not authorized to access this endpoint"))
		return
	}

	cfg.fileserverHits.Store(0)
	cfg.metricsHandler(w, r)

	cfg.db.ClearUsers(r.Context())
}

func (cfg *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	requestBody := struct {
		Body string `json:"body"`
	}{}
	err = decoder.Decode(&requestBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Decoding error", err)
		return
	}

	id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	params := database.CreateChirpParams{
		Body:   requestBody.Body,
		UserID: id,
	}

	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Message exceeds character limit", err)
		return
	}

	params.Body = cleanMessage(params.Body)

	chirp, err := cfg.db.CreateChirp(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, chirp)
}

func (cfg *apiConfig) usersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	params := database.CreateUserParams{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Decoding error", err)
		return
	}

	params.HashedPassword, err = auth.HashPassword(params.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Password Hashing failed", err)
		return
	}

	u, err := cfg.db.CreateUser(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}

	user := User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
	}

	respondWithJSON(w, http.StatusCreated, user)
}

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving chirps", err)
		return
	}
	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Not valid id", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, chirp)
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	loginParams := LoginParameters{}
	err := decoder.Decode(&loginParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Decoding error", err)
		return
	}

	u, err := cfg.db.GetUser(r.Context(), loginParams.Email)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "user not found", err)
		return
	}

	err = auth.CheckPasswordHash(loginParams.Password, u.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	JWTTokenString, err := auth.MakeJWT(u.ID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error Creating JWT", err)
		return
	}

	refreshString, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error Creating Refresh Token", err)
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshString,
		UserID:    u.ID,
		ExpiresAt: time.Now().Add(time.Hour * 1440),
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Refresh token could not be created", err)
		return
	}

	user := User{
		ID:           u.ID,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		Email:        u.Email,
		Token:        JWTTokenString,
		RefreshToken: refreshString,
	}

	respondWithJSON(w, http.StatusOK, user)
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error getting bearer token", err)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), refreshTokenString)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Token not found", err)
		return
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "Token expired", err)
		return
	}

	if !refreshToken.RevokedAt.Time.Equal(sql.NullTime{}.Time) {
		respondWithError(w, http.StatusUnauthorized, "Token revoked", err)
		return
	}

	JWTTokenString, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error Creating JWT", err)
		return
	}

	resp := struct {
		Token string `json:"token"`
	}{
		Token: JWTTokenString,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	refreshTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Error getting bearer token", err)
		return
	}

	_, err = cfg.db.RevokeToken(r.Context(), refreshTokenString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to revoke token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) updateUserHandler(w http.ResponseWriter, r *http.Request) {
	JWTTokenString, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error getting bearer token", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	loginParams := LoginParameters{}
	err = decoder.Decode(&loginParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Decoding error", err)
		return
	}

	id, err := auth.ValidateJWT(JWTTokenString, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unable to validate token", err)
	}

	hashedPassword, err := auth.HashPassword(loginParams.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Password Hashing failed", err)
		return
	}

	params := database.UpdateLoginInfoParams{
		Email:          loginParams.Email,
		HashedPassword: hashedPassword,
		ID:             id,
	}
	u, err := cfg.db.UpdateLoginInfo(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "User not found", err)
	}

	user := User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Email:     u.Email,
		Token:     JWTTokenString,
	}
	respondWithJSON(w, http.StatusOK, user)
}
