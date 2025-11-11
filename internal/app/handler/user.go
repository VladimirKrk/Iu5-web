package handler

import (
	"Iu5-web/internal/app/api_types"
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// RegisterUser godoc
// @Summary      Register a new user
// @Description  Creates a new user account.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        user body api_types.UserRegisterRequest true "User Registration Info"
// @Success      201  {object} api_types.UserResponse
// @Failure      400  {object} api_types.ErrorResponse "Invalid input"
// @Failure      409  {object} api_types.ErrorResponse "User already exists"
// @Failure      500  {object} api_types.ErrorResponse "Internal server error"
// @Router       /register [post]
func (h *Handler) RegisterUser(c *gin.Context) {
	var req api_types.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	user, err := h.Repository.CreateUser(req)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusCreated, api_types.ConvertUserToResponse(user))
}

// AuthenticateUser godoc
// @Summary      Log in a user
// @Description  Authenticates a user and returns a JWT token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        credentials body api_types.UserLoginRequest true "User Credentials"
// @Success      200  {object} api_types.TokenResponse
// @Failure      400  {object} api_types.ErrorResponse "Invalid input"
// @Failure      401  {object} api_types.ErrorResponse "Invalid credentials"
// @Router       /login [post]
func (h *Handler) AuthenticateUser(c *gin.Context) {
	var req api_types.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	token, err := h.Repository.AuthenticateUser(req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// DeauthorizeUser godoc
// @Summary      Log out a user
// @Description  Invalidates the current user's JWT token by adding it to a blacklist.
// @Tags         Auth
// @Produce      json
// @Success      200  {object} api_types.StatusResponse
// @Failure      401  {object} api_types.ErrorResponse "Unauthorized"
// @Security     BearerAuth
// @Router       /logout [post]
func (h *Handler) DeauthorizeUser(c *gin.Context) {
	tokenString := extractTokenFromHeader(c.Request)
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token claims"})
		return
	}
	ttl, err := GetTokenTTL(claims)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "logged out (token already expired)"})
		return
	}
	if err := h.Repository.AddTokenToBlacklist(context.Background(), tokenString, ttl); err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "logged out"})
}

// GetUserMe godoc
// @Summary      Get current user's profile
// @Description  Retrieves the profile information for the authenticated user.
// @Tags         User
// @Produce      json
// @Success      200  {object} api_types.UserResponse
// @Failure      401  {object} api_types.ErrorResponse "Unauthorized"
// @Security     BearerAuth
// @Router       /users/me [get]
func (h *Handler) GetUserMe(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api_types.ConvertUserToResponse(user))
}

// UpdateUserMe godoc
// @Summary      Update current user's profile
// @Description  Updates the password for the authenticated user.
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        user body api_types.UserUpdateRequest true "Fields to update"
// @Success      200  {object} api_types.UserResponse
// @Failure      400  {object} api_types.ErrorResponse "Invalid input"
// @Failure      401  {object} api_types.ErrorResponse "Unauthorized"
// @Security     BearerAuth
// @Router       /users/me [put]
func (h *Handler) UpdateUserMe(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var req api_types.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	user, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(c, err)
		return
	}
	if req.Password != "" {
		user.Password = req.Password
	}
	if err := h.Repository.UpdateUser(&user); err != nil {
		h.errorHandler(c, err)
		return
	}
	c.JSON(http.StatusOK, api_types.ConvertUserToResponse(user))
}
