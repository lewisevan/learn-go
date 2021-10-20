package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound  = errors.New("models: resource not found")
	ErrInvalidID = errors.New("models: ID provided was invalid")
)

const userPwPepper = "oDWHNpaC8zL5Tl1GkXzF"

type UserService struct {
	db *gorm.DB
}

type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
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
func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return us.AutoMigrate()
}

/*
 * Attempts to automatically migrate the users table.
 */
func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}

/*
 * Looks up a user given their user ID. Returns a user object
 * representing the user.
 */
func (us *UserService) ById(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

/*
 * Looks up a user given their email address. Returns a user
 * object representing the user.
 */
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

/*
 * Creates a new DB record for the provided User object, and will
 * backfill the gorm.Model fields
 */
func (us *UserService) Create(user *User) error {
	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(pwBytes), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return us.db.Create(user).Error
}

/*
 * Updates a user DB record with the data in the provided user object.
 * The user is found by ID, and all fields are updated.
 */
func (us *UserService) Update(user *User) error {
	return us.db.Save(user).Error
}

/*
 * Deletes the user DB record associated with the provided user ID.
 */
func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := User{
		Model: gorm.Model{
			ID: id,
		},
	}
	return us.db.Delete(&user).Error
}

/*
 * Queries the given DB and gets the first item in the resulting DB rows.
 * The row is parsed into the provided destination object pointer, which
 * can then be used by the caller to access the queried row.
 */
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}
