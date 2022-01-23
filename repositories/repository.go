package repositories

import "github.com/cheebz/go-auth/models"

type Repository interface {
	Close()
	GetUserByID(userID int) (models.User, error)
	GetUserByName(username string) (models.User, error)
	CreateUser(user models.User) (models.User, error)
	GetUserGroups(userID int) ([]models.Group, error)
	UpdatePassword(userID int, password string) error
	SaveRefresh(userID int, jti string) error
	ValidateRefresh(userID int, jti string) error
	InvalidateRefresh(jti string) error
	DeleteAllRefresh(userID int) error
	DeleteExpiredRefresh() error
}
