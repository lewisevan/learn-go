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
var _ UserDB = &userValidator{}

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

	hmac := hash.NewHMAC(hmacSecretKey)

	uv := &userValidator{
		hmac:   hmac,
		UserDB: ug,
	}

	return &userService{
		UserDB: uv,
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
	hmac hash.HMAC
}

type userValidatorFunc func(*User) error

func runUserValidatorFuncs(user *User, fns ...userValidatorFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

// Hashes the remember token and then calls the ByRemember method
// on the UserDB layer.
func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{
		Remember: token,
	}
	if err := runUserValidatorFuncs(&user, uv.hmacRemember); err != nil {
		return nil, err
	}
	return uv.UserDB.ByRemember(user.RememberHash)
}

func (uv *userValidator) Create(user *User) error {
	err := runUserValidatorFuncs(user, uv.bcryptPassword, uv.setRememberIfUnset, uv.hmacRemember)
	if err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

func (uv *userValidator) Update(user *User) error {
	err := runUserValidatorFuncs(user, uv.bcryptPassword, uv.hmacRemember)
	if err != nil {
		return err
	}

	return uv.UserDB.Update(user)
}

func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id
	err := runUserValidatorFuncs(&user, uv.idGreaterThan(0))
	if err != nil {
		return err
	}
	return uv.UserDB.Delete(id)
}

// Hashes a user's password with a predefined pepper and bcrypt
// if the password field is not empty
func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}

	pwBytes := []byte(user.Password + userPwPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(pwBytes), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

// Hashes a remember token if the remember token is not empty
func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) setRememberIfUnset(user *User) error {
	if user.Remember != "" {
		return nil
	}
	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

func (uv *userValidator) idGreaterThan(n uint) userValidatorFunc {
	return userValidatorFunc(func(user *User) error {
		if user.ID <= n {
			return ErrInvalidID
		}
		return nil
	})
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
	db *gorm.DB
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	return &userGorm{
		db: db,
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

// Looks up a user given their remember token. The remember token
// should already be hashed.
func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	db := ug.db.Where("remember_hash = ?", rememberHash)
	err := first(db, &user)
	return &user, err
}

// Creates a new DB record for the provided User object, and will
// backfill the gorm.Model fields
func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

// Updates a user DB record with the data in the provided user object.
// The user is found by ID, and all fields are updated.
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

// Deletes the user DB record associated with the provided user ID.
func (ug *userGorm) Delete(id uint) error {
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
