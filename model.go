package ngauth

// User holds information about application users
type User struct {
	ID          interface{} `json:"id" bson:"_id,omitempty"`
	Name        string      `json:"name"`
	Username    string      `json:"username"`
	Password    string      `json:"-"`
	Email       string      `json:"email"`
	PhoneNumber string      `json:"phone_number"`
	CreatedAt   int64       `json:"created_at"`
	//EmailVerifiedAt null.Time   `json:"-" bson:"email_verified_at"`
}

//OTP - one time password
type OTP struct {
	ID          interface{} `json:"id" bson:"_id,omitempty"`
	PhoneNumber string      `json:"phone_number"`
	Email       string      `json:"email"`
	Code        string      `json:"-"`

	//otp for REGISTER, RESET, 2FA etc
	OTPFor         string `json:"otp_for"`
	VerificationID string `json:"verification_id"`
	VerifiedAt     int64  `json:"verified_at"`
	ExpiresAt      int64  `json:"expires_at"`
	CreatedAt      int64  `json:"created_at"`
}

//Session - one time password
type Session struct {
	ID           interface{} `json:"id" bson:"_id,omitempty"`
	UserID       interface{} `json:"user_id"`
	DeviceID     string      `json:"device_id"`
	DeviceName   string      `json:"device_name"`
	RefreshToken string      `json:"refresh_token"`
	CreatedAt    int64       `json:"created_at"`
}
