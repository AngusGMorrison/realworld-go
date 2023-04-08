package user

import (
	"net/mail"
	"net/url"

	"github.com/google/uuid"
)

type EmailAddress string

// NewEmailAddressFromString returns a new EmailAddress from a raw string. If
// the raw string is not a valid email address according to RFC 5322 standards,
// an EmailAddressValidationError is returned.
//
// Note that RFC 5322 permits some seemingly bad email addresses such as
// mail@localdestination, since an email addresses don't necessarily need a
// public domain.
func NewEmailAddressFromString(raw string) (EmailAddress, error) {
	if _, err := mail.ParseAddress(raw); err != nil {
		return "", ErrEmailAddressUnparseable
	}

	return EmailAddress(raw), nil
}

type PasswordDigest string

// User is the central domain type for this package.
type User struct {
	username       string
	email          EmailAddress
	passwordDigest PasswordDigest
	bio            string
	imageURL       string
}

func NewUser(
	username string,
	email EmailAddress,
	digest PasswordDigest,
	bio string,
	imageURL string,
) *User {
	return &User{
		username:       username,
		email:          email,
		passwordDigest: digest,
		bio:            bio,
		imageURL:       imageURL,
	}
}

// Username returns the user's username.
func (u *User) Username() string {
	return u.username
}

// Email returns the user's email address.
func (u *User) Email() EmailAddress {
	return u.email
}

// Digest returns the user's password digest.
func (u *User) Digest() PasswordDigest {
	return u.passwordDigest
}

// Bio returns the user's bio.
func (u *User) Bio() string {
	return u.bio
}

// ImageURL returns the user's image URL.
func (u *User) ImageURL() string {
	return u.imageURL
}

// RegisterRequest describes the data required to register a new user.
type RegisterRequest struct {
	username       string
	email          EmailAddress
	passwordDigest PasswordDigest
}

// NewRegisterRequest instantiates a *RegisterRequest, validating rawEmail.
func NewRegisterRequest(
	username string,
	rawEmail string,
	password string,
) (*RegisterRequest, error) {
	validationErr := &ValidationError{}

	emailAddr, err := NewEmailAddressFromString(rawEmail)
	if err != nil {
		validationErr.Push(err)
	}

	pwDigest, err := digest(password)
	if err != nil {
		validationErr.Push(err)
	}

	if validationErr.Any() {
		return nil, validationErr
	}

	return &RegisterRequest{
		username:       username,
		email:          emailAddr,
		passwordDigest: pwDigest,
	}, nil
}

// AuthRequest describes the data required to authenticate a user.
type AuthRequest struct {
	email          EmailAddress
	passwordDigest PasswordDigest
}

// NewAuthRequest instantiates a *AuthRequest, validating rawEmail.
func NewAuthRequest(rawEmail string, password string) (*AuthRequest, error) {
	validationErr := &ValidationError{}

	emailAddr, err := NewEmailAddressFromString(rawEmail)
	if err != nil {
		validationErr.Push(err)
	}

	pwDigest, err := digest(password)
	if err != nil {
		validationErr.Push(err)
	}

	if validationErr.Any() {
		return nil, validationErr
	}

	return &AuthRequest{
		email:          emailAddr,
		passwordDigest: pwDigest,
	}, nil
}

func (ar *AuthRequest) Email() EmailAddress {
	return ar.email
}

func (ar *AuthRequest) PasswordDigest() PasswordDigest {
	return ar.passwordDigest
}

// UpdateRequest describes the data required to update a user. Since zero or
// more fields may be updated in a single request, pointer fields are required
// to distinguish the absence of a value (i.e. no change) from the zero value.
type UpdateRequest struct {
	userID         uuid.UUID
	email          *EmailAddress
	passwordDigest *PasswordDigest
	bio            *string
	imageURL       *string
}

// NewUpdateRequest instantiates a *UpdateRequest, validating rawEmail and
// imageURL.
func NewUpdateRequest(
	userID uuid.UUID,
	rawEmail *string,
	password *string,
	bio *string,
	imageURL *string,
) (*UpdateRequest, error) {
	req := &UpdateRequest{
		userID: userID,
		bio:    bio,
	}
	validationErr := &ValidationError{}

	if rawEmail != nil {
		emailAddr, err := NewEmailAddressFromString(*rawEmail)
		if err != nil {
			validationErr.Push(err)
		}
		req.email = &emailAddr
	}

	if password != nil {
		pwDigest, err := digest(*password)
		if err != nil {
			validationErr.Push(err)
		}
		req.passwordDigest = &pwDigest
	}

	if imageURL != nil {
		if _, err := url.Parse(*imageURL); err != nil {
			validationErr.Push(ErrImageURLUnparseable)
		}
		req.imageURL = imageURL
	}

	if validationErr.Any() {
		return nil, validationErr
	}

	return req, nil
}
