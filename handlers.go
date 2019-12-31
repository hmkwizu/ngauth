package ngauth

import (
	"net/http"
	"time"
)

//PwdHashFunc - function to hash a plain password
type PwdHashFunc = func(plainPassword string) string

//PwdCheckFunc - function to check if hashed and plain passwords match
type PwdCheckFunc = func(hashedPwd string, plainPwd string) bool

const otpForRegister = "REGISTER"
const otpForReset = "RESET"

// GenerateOTP - first step in registration, only email/phone is taken from user and otp code sent
func GenerateOTP(db Database, lang string, params map[string]interface{}, sendOTPCallback func(email, phoneNo, verifCode string)) (map[string]interface{}, *Error) {

	email := GetStringOrEmpty(params["email"])
	phoneNumber := GetStringOrEmpty(params["phone_number"])
	countryCode := GetStringOrEmpty(params["country_code"])
	otpFor := GetStringOrEmpty(params["otp_for"])

	//flag to show whether to use email or phonenumber
	useEmail := true
	if len(phoneNumber) > 0 {
		useEmail = false
		email = ""
	}

	//check for empty fields
	if useEmail && IsEmptyTextContent(email) || (!useEmail && IsEmptyTextContent(phoneNumber)) || IsEmptyTextContent(otpFor) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	if otpFor != otpForRegister && otpFor != otpForReset {
		return nil, NewErrorWithMessage(ErrorWrongValueFor, ErrorText(lang, ErrorWrongValueFor)+"otp_for")
	}

	//validate - email
	if useEmail {
		if !IsValidEmail(email) {
			return nil, NewError(lang, ErrorInvalidEmail)
		}

	} else {
		//validate - phone
		num, err := IsValidPhoneNumber(phoneNumber, countryCode, lang)
		if err != nil {
			return nil, err
		}
		phoneNumber = num
	}

	// //check for existing verified user
	// otp, err := db.GetOTP(email, phoneNumber, otpFor, lang)
	// if err != nil {
	// 	return nil, err
	// }

	// if otp != nil && otp.VerifiedAt > 0 {
	// 	return nil, NewError(lang, ErrorAlreadyVerified)
	// }

	verifCode := SecureRandomNumericStringStandard()

	expiresAt := ExpireAtTime(time.Duration(Config.OTPExpireMins) * time.Minute) //in 5mins

	_, err := db.CreateOTP(OTP{Code: verifCode, OTPFor: otpFor, Email: email, PhoneNumber: phoneNumber, ExpiresAt: NullTimeFrom(expiresAt), CreatedAt: NullTimeFrom(TimeNow())}, lang)
	if err != nil {
		return nil, err
	}

	//dispatch email sending
	if sendOTPCallback != nil {
		sendOTPCallback(email, phoneNumber, verifCode)
	}

	//Prepare the response
	response := make(map[string]interface{})
	response["code"] = http.StatusOK
	response["success"] = true

	return response, nil
}

// VerifyOTP - verify otp
func VerifyOTP(db Database, lang string, params map[string]interface{}) (map[string]interface{}, *Error) {

	email := GetStringOrEmpty(params["email"])
	phoneNumber := GetStringOrEmpty(params["phone_number"])
	countryCode := GetStringOrEmpty(params["country_code"])
	otpCode := GetStringOrEmpty(params["code"])
	otpFor := GetStringOrEmpty(params["otp_for"])

	//flag to show whether to use email or phonenumber
	useEmail := true
	if len(phoneNumber) > 0 {
		useEmail = false
		email = ""
	}

	//check for empty fields
	if IsEmptyTextContent(otpCode) {
		return nil, NewError(lang, ErrorInvalidOTPCode)
	}

	if useEmail && IsEmptyTextContent(email) || (!useEmail && IsEmptyTextContent(phoneNumber)) || IsEmptyTextContent(otpFor) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	if otpFor != otpForRegister && otpFor != otpForReset {
		return nil, NewErrorWithMessage(ErrorWrongValueFor, ErrorText(lang, ErrorWrongValueFor)+"otp_for")
	}

	//validate - email
	if useEmail {
		if !IsValidEmail(email) {
			return nil, NewError(lang, ErrorInvalidEmail)
		}

	} else {
		//validate - phone
		num, err := IsValidPhoneNumber(phoneNumber, countryCode, lang)
		if err != nil {
			return nil, err
		}
		phoneNumber = num
	}

	otp, err := db.GetOTP(email, phoneNumber, otpFor, lang)
	if err != nil {
		return nil, err
	}

	//empty result
	if otp == nil {
		return nil, NewError(lang, ErrorNotFound)
	}

	//invalid otp
	if otp != nil && otp.Code != otpCode {
		return nil, NewError(lang, ErrorInvalidOTPCode)
	}

	//already verified
	if otp != nil && otp.VerifiedAt.Valid {
		return nil, NewError(lang, ErrorAlreadyVerified)
	}

	//valid otp
	if otp != nil && otp.Code == otpCode && otp.ExpiresAt.Valid && otp.ExpiresAt.Time.After(TimeNow()) {

		verifID := GenerateUUID()

		//update db
		err = db.UpdateOTPByID(otp.ID, Map{"verified_at": TimeNow(), "verification_id": verifID}, lang)
		if err != nil {
			return nil, err
		}

		//Prepare the response
		response := make(map[string]interface{})
		response["code"] = http.StatusOK
		response["success"] = true
		response["verification_id"] = verifID
		return response, nil
	}

	//check if otp expired
	if otp != nil && otp.ExpiresAt.Valid && otp.ExpiresAt.Time.Before(TimeNow()) {
		return nil, NewError(lang, ErrorExpiredOTPCode)
	}

	return nil, NewError(lang, ErrorInvalidOTPCode)
}

// Register - register user
func Register(db Database, lang string, params map[string]interface{}, pwdHashCallback PwdHashFunc) (map[string]interface{}, *Error) {

	if pwdHashCallback == nil {
		return nil, NewError(lang, ErrorMissingFunctionParams)
	}

	username := GetStringOrEmpty(params["username"])
	email := GetStringOrEmpty(params["email"])
	phoneNumber := GetStringOrEmpty(params["phone_number"])
	countryCode := GetStringOrEmpty(params["country_code"])
	password := GetStringOrEmpty(params["password"])
	confirmPassword := GetStringOrEmpty(params["confirm_password"])
	verificationID := GetStringOrEmpty(params["verification_id"])

	//flag to show whether to use email or phonenumber
	useEmail := true
	if len(phoneNumber) > 0 {
		useEmail = false
		//email = ""
	}

	// empty - email or phone
	if useEmail && IsEmptyTextContent(email) || (!useEmail && IsEmptyTextContent(phoneNumber)) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	// empty - password/confirm password
	if IsEmptyString(password) || IsEmptyString(confirmPassword) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	//validate - email
	if useEmail {
		if !IsValidEmail(email) {
			return nil, NewError(lang, ErrorInvalidEmail)
		}

	} else {
		//validate - phone
		num, err := IsValidPhoneNumber(phoneNumber, countryCode, lang)
		if err != nil {
			return nil, err
		}
		phoneNumber = num
	}

	//validate - password
	if password != confirmPassword {
		return nil, NewError(lang, ErrorPasswordsDoNotMatch)
	}

	//check if user already exists
	regdUser, err := db.GetUserBy(email, phoneNumber, lang)
	if err != nil {
		return nil, err
	}

	if regdUser != nil {
		return nil, NewError(lang, ErrorUsernameExists)

	}

	//verify before registration
	if Config.VerifyBeforeRegister {

		otp, err := db.GetOTP(email, phoneNumber, otpForRegister, lang)
		if err != nil {
			return nil, err
		}

		//invalid otp verification
		if otp == nil || otp.VerificationID != verificationID {
			return nil, NewError(lang, ErrorGetVerifiedFirst)
		}

	}

	//now lets register the user
	hashedPassword := pwdHashCallback(password)
	user := User{Username: username, Email: email, PhoneNumber: phoneNumber, Password: hashedPassword, CreatedAt: NullTimeFrom(TimeNow())}
	result, err := db.CreateUser(user, lang)
	if err != nil {
		return nil, err
	}

	//Prepare the response
	response := make(map[string]interface{})
	response["code"] = http.StatusCreated
	response["success"] = true
	response["id"] = result

	return response, nil
}

// Login returns access_token if correct username and password are given.
func Login(db Database, lang string, params map[string]interface{}, pwdCheckCallback PwdCheckFunc) (map[string]interface{}, *Error) {

	if pwdCheckCallback == nil {
		return nil, NewError(lang, ErrorMissingFunctionParams)
	}

	email := GetStringOrEmpty(params["email"])
	phoneNumber := GetStringOrEmpty(params["phone_number"])
	countryCode := GetStringOrEmpty(params["country_code"])
	password := GetStringOrEmpty(params["password"])

	//flag to show whether to use email or phonenumber
	useEmail := true
	if len(phoneNumber) > 0 {
		useEmail = false
		//email = ""
	}

	//check for empty fields
	if useEmail && IsEmptyTextContent(email) || (!useEmail && IsEmptyTextContent(phoneNumber)) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	// empty - password
	if IsEmptyString(password) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	//validate - email
	if useEmail {
		if !IsValidEmail(email) {
			return nil, NewError(lang, ErrorInvalidEmail)
		}

	} else {
		//validate - phone
		num, err := IsValidPhoneNumber(phoneNumber, countryCode, lang)
		if err != nil {
			return nil, err
		}
		phoneNumber = num
	}

	user, err := db.GetUserBy(email, phoneNumber, lang)
	if err != nil {
		return nil, err
	}

	//no record found
	if user == nil {
		return nil, NewError(lang, ErrorNotFound)
	}

	//passwords do not match
	if !pwdCheckCallback(user.Password, password) {
		return nil, NewError(lang, ErrorIncorrectUsernameOrPassword)
	}

	//access token
	accessToken, err := GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	//refresh token
	refreshToken, err := GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	//save refresh token to db
	_, err = db.CreateSession(Session{UserID: user.ID, RefreshToken: refreshToken, CreatedAt: NullTimeFrom(TimeNow())}, lang)
	if err != nil {
		return nil, err
	}

	//Prepare the response
	response := make(map[string]interface{})
	response["code"] = http.StatusOK
	response["success"] = true
	response["id"] = user.ID
	response["access_token"] = accessToken
	response["refresh_token"] = refreshToken

	return response, nil
}

// ChangePassword - changes password of a user
func ChangePassword(db Database, lang string, params map[string]interface{}, pwdCheckCallback PwdCheckFunc, pwdHashCallback PwdHashFunc) (map[string]interface{}, *Error) {

	if pwdCheckCallback == nil || pwdHashCallback == nil {
		return nil, NewError(lang, ErrorMissingFunctionParams)
	}

	email := GetStringOrEmpty(params["email"])
	phoneNumber := GetStringOrEmpty(params["phone_number"])
	countryCode := GetStringOrEmpty(params["country_code"])
	password := GetStringOrEmpty(params["password"])
	newPassword := GetStringOrEmpty(params["new_password"])
	confirmNewPassword := GetStringOrEmpty(params["confirm_new_password"])

	//flag to show whether to use email or phonenumber
	useEmail := true
	if len(phoneNumber) > 0 {
		useEmail = false
		//email = ""
	}

	//check for empty fields
	if useEmail && IsEmptyTextContent(email) || (!useEmail && IsEmptyTextContent(phoneNumber)) {
		return nil, NewError(lang, ErrorEmptyFields)
	}
	// empty - password/confirm password
	if IsEmptyString(password) || IsEmptyString(newPassword) || IsEmptyString(confirmNewPassword) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	//validate - email
	if useEmail {
		if !IsValidEmail(email) {
			return nil, NewError(lang, ErrorInvalidEmail)
		}

	} else {
		//validate - phone
		num, err := IsValidPhoneNumber(phoneNumber, countryCode, lang)
		if err != nil {
			return nil, err
		}
		phoneNumber = num
	}

	//validate - new password
	if newPassword != confirmNewPassword {
		return nil, NewError(lang, ErrorPasswordsDoNotMatch)
	}

	user, err := db.GetUserBy(email, phoneNumber, lang)
	if err != nil {
		return nil, err
	}

	//no record found
	if user == nil {
		return nil, NewError(lang, ErrorNotFound)
	}

	//user's password incorrect
	if !pwdCheckCallback(user.Password, password) {
		return nil, NewError(lang, ErrorIncorrectUsernameOrPassword)
	}

	//now let's change password
	hashedPassword := pwdHashCallback(newPassword)
	err = db.UpdateUserByID(user.ID, Map{"password": hashedPassword}, lang)
	if err != nil {
		return nil, err
	}

	//Prepare the response
	response := make(map[string]interface{})
	response["code"] = http.StatusOK
	response["success"] = true
	response["id"] = user.ID

	return response, nil
}

// ResetPassword - reset password of a user
func ResetPassword(db Database, lang string, params map[string]interface{}, pwdHashCallback PwdHashFunc) (map[string]interface{}, *Error) {

	if pwdHashCallback == nil {
		return nil, NewError(lang, ErrorMissingFunctionParams)
	}

	email := GetStringOrEmpty(params["email"])
	phoneNumber := GetStringOrEmpty(params["phone_number"])
	countryCode := GetStringOrEmpty(params["country_code"])
	password := GetStringOrEmpty(params["password"])
	confirmPassword := GetStringOrEmpty(params["confirm_password"])
	verificationID := GetStringOrEmpty(params["verification_id"])

	//flag to show whether to use email or phonenumber
	useEmail := true
	if len(phoneNumber) > 0 {
		useEmail = false
		//email = ""
	}

	//check for empty fields
	if useEmail && IsEmptyTextContent(email) || (!useEmail && IsEmptyTextContent(phoneNumber)) {
		return nil, NewError(lang, ErrorEmptyFields)
	}
	// empty - password/confirm password
	if IsEmptyString(password) || IsEmptyString(confirmPassword) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	//validate - email
	if useEmail {
		if !IsValidEmail(email) {
			return nil, NewError(lang, ErrorInvalidEmail)
		}

	} else {
		//validate - phone
		num, err := IsValidPhoneNumber(phoneNumber, countryCode, lang)
		if err != nil {
			return nil, err
		}
		phoneNumber = num
	}

	//validate - password
	if password != confirmPassword {
		return nil, NewError(lang, ErrorPasswordsDoNotMatch)
	}

	user, err := db.GetUserBy(email, phoneNumber, lang)
	if err != nil {
		return nil, err
	}

	//no record found
	if user == nil {
		return nil, NewError(lang, ErrorNotFound)
	}

	//check verification_id for the user
	otp, err := db.GetOTP(email, phoneNumber, otpForReset, lang)
	if err != nil {
		return nil, err
	}

	//invalid otp verification
	if otp == nil || otp.VerificationID != verificationID {
		return nil, NewError(lang, ErrorGetVerifiedFirst)
	}

	//now let's reset password
	hashedPassword := pwdHashCallback(password)
	err = db.UpdateUserByID(user.ID, Map{"password": hashedPassword}, lang)
	if err != nil {
		return nil, err
	}

	//Prepare the response
	response := make(map[string]interface{})
	response["code"] = http.StatusOK
	response["success"] = true
	response["id"] = user.ID

	return response, nil
}

// Token - refresh access_token using your refresh_token
func Token(db Database, lang string, params map[string]interface{}) (map[string]interface{}, *Error) {

	userID := GetStringOrEmpty(params["user_id"])
	refreshToken := GetStringOrEmpty(params["refresh_token"])

	// Validate refresh token
	err := IsValidToken(refreshToken)
	if err != nil {
		return nil, err
	}

	//check for empty fields
	if IsEmptyTextContent(refreshToken) || IsEmptyTextContent(userID) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	//check for existing verified user
	session, err := db.GetSession(refreshToken, lang)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, NewError(lang, ErrorInvalidToken)
	}

	//access token
	accessToken, err := GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}

	//Prepare the response
	response := make(map[string]interface{})
	response["code"] = http.StatusOK
	response["success"] = true
	response["id"] = userID
	response["access_token"] = accessToken
	response["refresh_token"] = refreshToken

	return response, nil
}
