package ngauth

import (
	"gopkg.in/guregu/null.v3"
)

// User holds information about application users
type User struct {
	ID          interface{} `json:"id" bson:"_id,omitempty"`
	Name        string      `json:"name"`
	Username    string      `json:"username"`
	Password    string      `json:"-"`
	Email       string      `json:"email"`
	PhoneNumber string      `json:"phone_number"`
	PhotoURL    string      `json:"photo_url"`
	CreatedAt   null.Time   `json:"created_at"`
}

//OTP - one time password
type OTP struct {
	ID          interface{} `json:"id" bson:"_id,omitempty"`
	PhoneNumber string      `json:"phone_number"`
	Email       string      `json:"email"`
	Code        string      `json:"-"`

	//otp for REGISTER, RESET, 2FA etc
	OTPFor         string    `json:"otp_for"`
	VerificationID string    `json:"verification_id"`
	VerifiedAt     null.Time `json:"verified_at"`
	ExpiresAt      null.Time `json:"expires_at"`
	CreatedAt      null.Time `json:"created_at"`
}

//Session - one time password
type Session struct {
	ID           interface{} `json:"id" bson:"_id,omitempty"`
	UserID       interface{} `json:"user_id"`
	DeviceID     string      `json:"device_id"`
	DeviceName   string      `json:"device_name"`
	RefreshToken string      `json:"refresh_token"`
	CreatedAt    null.Time   `json:"created_at"`

	IPAddr    string `json:"ip_addr"`
	UserAgent string `json:"user_agent"`
}

//PushToken - push notification tokens
type PushToken struct {
	ID        interface{} `json:"id" bson:"_id,omitempty"`
	DeviceID  string      `json:"device_id"`
	DeviceOS  string      `json:"device_os"`
	PushToken string      `json:"push_token"`
	UserID    interface{} `json:"user_id"`
	CreatedAt null.Time   `json:"created_at"`
	UpdatedAt null.Time   `json:"updated_at"`

	IPAddr    string `json:"ip_addr"`
	UserAgent string `json:"user_agent"`
}
