package ngauth

// Database - Database interface, all db need to implement this
type Database interface {
	Init(config *Configuration) error
	//GetID(id interface{}) interface{}

	//GetUser(userID interface{}, lang string) (*User, *Error)
	GetUserBy(email string, phoneNo string, lang string) (*User, *Error)
	CreateUser(user User, lang string) (interface{}, *Error)
	UpdateUserByID(userID interface{}, columns interface{}, lang string) *Error

	//###########  OTP
	// GetOTP - returns the most current otp
	GetOTP(email string, phoneNo string, otpFor string, lang string) (*OTP, *Error)
	GetOTPs(email string, phoneNo string, otpFor string, offset int64, limit int64, lang string) ([]OTP, *Error)
	// CreateOTP - save otp to db
	CreateOTP(otp OTP, lang string) (interface{}, *Error)
	UpdateOTPByID(otpID interface{}, columns interface{}, lang string) *Error

	//########### Sessions
	CreateSession(session Session, lang string) (interface{}, *Error)
	GetSession(refreshToken string, lang string) (*Session, *Error)

	//########### Push Tokens
	CreateOrUpdatePushToken(pushToken PushToken, lang string) *Error
	GetPushToken(deviceID string, lang string) (*PushToken, *Error)
	GetPushTokensForUserID(userID interface{}, lang string) ([]PushToken, *Error)
	GetPushTokens(userIDs []interface{}, lang string) ([]PushToken, *Error)
}
