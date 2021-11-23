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

// Ensures types adhere to UserDB interface
var _ UserDB = &userGorm{}
var _ UserDB = userValidator{}

type User struct {
	gorm.Model
	Name         string
	Email        string `gorm:"not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

// A set of methods used to manipulate and work with the user model
type UserService interface {
	// Verifies the provided email and password are correct.
	// If they are correct, the user corresponding to the email
	// will be returned. Otherwise, you will receive an
	// ErrNotFound, ErrInvalidPassword, or another error if something
	// unexpcted goes wrong
	Authenticate(email, password string) (*User, error)

	UserDB
}

// The implementation of the UserService interface
type userService struct {
	UserDB
}

func NewUserService(connectionInfo string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}

	return &userService{
		UserDB: &userValidator{
			UserDB: ug,
		},
	}, err
}

// Accepts an email address and password and determines whether both
// correctly map to a user.
func (us *userService) Authenticate(email string, password string) (*User, error) {
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

type userValidator struct {
	UserDB
}

type UserDB interface {
	// Methods for querying for single user
	ById(id uint) (*User, error)
	ByEmail(email string) (*User, error)
	ByRemember(token string) (*User, error)

	// Methods for altering users
	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error

	// Used to close a DB connection
	Close() error

	// Migration helpers
	AutoMigrate() error
	DestructiveReset() error
}

type userGorm struct {
	db   *gorm.DB
	hmac hash.HMAC
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	hmac := hash.NewHMAC(hmacSecretKey)

	return &userGorm{
		db:   db,
		hmac: hmac,
	}, nil
}

// Closes the UserService database connection.
func (ug *userGorm) Close() error {
	return ug.db.Close()
}

// Drops the user table and then rebuilds it.
func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}
	return ug.AutoMigrate()
}

// Attempts to automatically migrate the users table.
func (ug *userGorm) AutoMigrate() error {
	if err := ug.db.AutoMigrate(&User{}).Error; err != nil {
		return err
	}
	return nil
}

// Looks up a user given their user ID. Returns a user object
// representing the user.
func (ug *userGorm) ById(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	return &user, err
}

// Looks up a user given their email address. Returns a user
// object representing the user.
func (ug *userGorm) ByEmail(email string) (*User, error) {
	var user User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

// Looks up a user given their remember token. This method will handle
// hashing the token before lookup
func (ug *userGorm) ByRemember(token string) (*User, error) {
	var user User
	hashedToken := ug.hmac.Hash(token)
	db := ug.db.Where("remember_hash = ?", hashedToken)
	err := first(db, &user)
	return &user, err
}

// Creates a new DB record for the provided User object, and will
// backfill the gorm.Model fields
func (ug *userGorm) Create(user *User) error {
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

	user.RememberHash = ug.hmac.Hash(user.Remember)

	return ug.db.Create(user).Error
}

// Updates a user DB record with the data in the provided user object.
// The user is found by ID, and all fields are updated.
func (ug *userGorm) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = ug.hmac.Hash(user.Remember)
	}

	return ug.db.Save(user).Error
}

// Deletes the user DB record associated with the provided user ID.
func (ug *userGorm) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}
	user := User{
		Model: gorm.Model{
			ID: id,
		},
	}
	return ug.db.Delete(&user).Error
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
