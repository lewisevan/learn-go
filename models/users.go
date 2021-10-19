package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var (
	ErrNotFound = errors.New("models: resource not found")
)

type UserService struct {
	db *gorm.DB
}

type User struct {
	gorm.Model
	Name  string
	Email string `gorm:"not null;unique_index"`
}

func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}
	return &UserService{
		db: db,
	}, err
}

/*
 * Closes the UserService database connection.
 */
func (us *UserService) Close() error {
	return us.db.Close()
}

/*
 * Drops the user table and then rebuilds it.
 */
func (us *UserService) DestructiveReset() {
	us.db.DropTableIfExists(&User{})
	us.db.AutoMigrate(&User{})
}

/*
 * Looks up a user given their user ID. Returns a user object
 * representing the user.
 */
func (us *UserService) ById(id uint) (*User, error) {
	var user User
	err := us.db.Where("id = ?", id).First(&user).Error

	switch err {
	case nil:
		return &user, nil
	case gorm.ErrRecordNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

/*
 * Creates a new DB record for the provided User object, and will
 * backfill the gorm.Model fields
 */
func (us *UserService) Create(user *User) error {
	return us.db.Create(user).Error
}

/*
 * Updates a user DB record with the data in the provided user object.
 * The user is found by ID, and all fields are updated.
 */
func (us *UserService) Update(user *User) error {
	return us.db.Save(user).Error
}
