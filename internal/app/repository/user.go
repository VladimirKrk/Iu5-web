package repository

import (
	"Iu5-web/internal/app/api_types"
	"Iu5-web/internal/app/ds"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func (r *Repository) GetUserByID(id uint) (ds.User, error) {
	var user ds.User
	if err := r.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, fmt.Errorf("%w: user with id %d", ErrNotFound, id)
		}
		return ds.User{}, err
	}
	return user, nil
}

func (r *Repository) GetUserByLogin(login string) (ds.User, error) {
	var user ds.User
	if err := r.db.Where("login = ?", login).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ds.User{}, fmt.Errorf("%w: user with login '%s'", ErrNotFound, login)
		}
		return ds.User{}, err
	}
	return user, nil
}

func (r *Repository) CreateUser(req api_types.UserRegisterRequest) (ds.User, error) {
	_, err := r.GetUserByLogin(req.Login)
	if !errors.Is(err, ErrNotFound) {
		if err == nil {
			return ds.User{}, fmt.Errorf("%w: user with login '%s'", ErrAlreadyExists, req.Login)
		}
		return ds.User{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return ds.User{}, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := ds.User{
		Login:       req.Login,
		Password:    string(hashedPassword),
		IsModerator: false,
	}

	if err := r.db.Create(&newUser).Error; err != nil {
		return ds.User{}, fmt.Errorf("failed to create user: %w", err)
	}
	return newUser, nil
}

func (r *Repository) AuthenticateUser(req api_types.UserLoginRequest) (string, error) {
	user, err := r.GetUserByLogin(req.Login)
	if err != nil {
		return "", errors.New("invalid login or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", errors.New("invalid login or password")
	}

	return generateToken(user)
}

func (r *Repository) UpdateUser(user *ds.User) error {
	return r.db.Save(user).Error
}

func generateToken(user ds.User) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET environment variable is not set")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":      user.ID,
		"is_moderator": user.IsModerator,
		"exp":          time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}
