package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lewismevan/learn-go/hash"
	"github.com/lewismevan/learn-go/rand"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNotFound        = errors.New("models: resource not found")
	ErrInvalidID       = errors.New("models: ID provided was invalid")
	ErrInvalidPassword = errors.New("models: incorrect password provided")
)

const userPwPepper = "secret-pw-pepper"
const hmacSecretKey = "secret-hmac-key"

type UserService struct {
	db   *gorm.DB
	hmac hash.HMAC
}

type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	hmac := hash.NewHMAC(hmacSecretKey)

	return &UserService{
		db:   db,
		hmac: hmac,
	}, err
}

// Closes the UserService database connection.
func (us *UserService) Close() error {
	return us.db.Close()
}

// Drops the user table and then rebuilds it.
func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return us.AutoMigrate()
}

// Attempts to automatically migrate the users table.
func (us *UserService) AutoMigrate() error {
	if err := us.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}

// Looks up a user given their user ID. Returns a user object
// representing the user.
func (us *UserService) ById(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

// Looks up a user given their email address. Returns a user
// object representing the user.
func (us *UserService) ByEmail(email string) (*User, error) {
	var user User
	db := us.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

// Looks up a user given their remember token. This method will handle
// hashing the token before lookup
func (us *UserService) ByRemember(token string) (*User, error) {
	var user User
	hashedToken := us.hmac.Hash(token)
	db := us.db.Where("remember_hash = ?", hashedToken)
	err := first(db, &user)
	return &user, err
}

// Accepts an email address and password and determines whether both
// correctly map to a user.
func (us *UserService) Authenticate(email string, password string) (*User, error) {
	foundUser, err := us.ByEmail(email)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+userPwPepper))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrInvalidPassword
		default:
			return nil, err
		}
	}

	return foundUser, nil
}

// Creates a new DB record for the provided User object, and will
// backfill the gorm.Model fields
func (us *UserService) Create(user *User) error {
	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(pwBytes), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
	}

	user.RememberHash = us.hmac.Hash(user.Remember)

	return us.db.Create(user).Error
}

// Updates a user DB record with the data in the provided user object.
// The user is found by ID, and all fields are updated.
func (us *UserService) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = us.hmac.Hash(user.Remember)
	}

	return us.db.Save(user).Error
}

// Deletes the user DB record associated with the provided user ID.
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

// Queries the given DB and gets the first item in the resulting DB rows.
// The row is parsed into the provided destination object pointer, which
// can then be used by the caller to access the queried row.
func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}
	return err
}
