package ngauth

import (
	"fmt"
	"net/http"
	"time"

	"gopkg.in/guregu/null.v3"
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

	//check for existing verified user -- when registering
	if Config.VerifyBeforeRegister && otpFor == otpForRegister {
		//check if user already exists
		regdUser, err := db.GetUserBy(email, phoneNumber, lang)
		if err != nil {
			return nil, err
		}

		if regdUser != nil {
			return nil, NewError(lang, ErrorUsernameExists)
		}
	}

	//check for unregistered user -- when resetting password
	if otpFor == otpForReset {

		regdUser, err := db.GetUserBy(email, phoneNumber, lang)
		if err != nil {
			return nil, err
		}

		if regdUser == nil {
			return nil, NewError(lang, ErrorUserNotFound)
		}
	}

	//ban users from resending too many times in a short time
	otpList, err := db.GetOTPs(email, phoneNumber, otpFor, 0, int64(Config.OTPMaxRetry), lang)
	if err != nil {
		return nil, err
	}

	if otpList != nil && len(otpList) >= Config.OTPMaxRetry {
		otp1 := otpList[0]
		otp2 := otpList[Config.OTPMaxRetry-1]

		if otp1.CreatedAt.Valid && otp2.CreatedAt.Valid {
			diffBounds := int64(otp1.CreatedAt.Time.Sub(otp2.CreatedAt.Time) / time.Second)

			diffNowToLastOTP := int64(time.Now().Sub(otp1.CreatedAt.Time) / time.Second)

			//check if time less than findtime
			if diffBounds < Config.OTPFindTime && diffNowToLastOTP < Config.OTPBanTime {
				waitFor := Config.OTPBanTime - diffNowToLastOTP
				waitForMsg := fmt.Sprintf("%s: %d seconds", ErrorText(lang, ErrorWaitFor), waitFor)
				return nil, NewErrorWithMessage(ErrorBadRequest, waitForMsg)
			}

		}

	}

	verifCode := SecureRandomNumericStringStandard()

	expiresAt := ExpireAtTime(time.Duration(Config.OTPExpireTime) * time.Second)

	_, err = db.CreateOTP(OTP{Code: verifCode, OTPFor: otpFor, Email: email, PhoneNumber: phoneNumber, ExpiresAt: null.TimeFrom(expiresAt), CreatedAt: null.TimeFrom(TimeNow())}, lang)
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

	name := GetStringOrEmpty(params["name"])
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
	user := User{Name: name, Username: username, Email: email, PhoneNumber: phoneNumber, Password: hashedPassword, CreatedAt: null.TimeFrom(TimeNow())}
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

	ipAddr := GetStringOrEmpty(params["ip_addr"])
	userAgent := GetStringOrEmpty(params["user_agent"])

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
	_, err = db.CreateSession(Session{UserID: user.ID, RefreshToken: refreshToken, CreatedAt: null.TimeFrom(TimeNow()), IPAddr: ipAddr, UserAgent: userAgent}, lang)
	if err != nil {
		return nil, err
	}

	//Prepare the response
	response := make(map[string]interface{})
	response["code"] = http.StatusOK
	response["success"] = true
	response["id"] = user.ID
	response["name"] = user.Name
	response["photo_url"] = user.PhotoURL
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

// UpdatePushToken - update push device token
func UpdatePushToken(db Database, lang string, params map[string]interface{}) (map[string]interface{}, *Error) {

	token := GetStringOrEmpty(params["push_token"])
	deviceID := GetStringOrEmpty(params["device_id"])
	deviceOS := GetStringOrEmpty(params["device_os"])

	if IsEmptyString(token) || IsEmptyString(deviceID) || IsEmptyString(deviceOS) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	ipAddr := GetStringOrEmpty(params["ip_addr"])
	userAgent := GetStringOrEmpty(params["user_agent"])

	createdAt := null.TimeFrom(TimeNow())
	push := PushToken{DeviceID: deviceID, DeviceOS: deviceOS, PushToken: token, IPAddr: ipAddr, UserAgent: userAgent, CreatedAt: createdAt, UpdatedAt: createdAt}

	err := db.CreateOrUpdatePushToken(push, lang)
	if err != nil {
		return nil, err
	}

	//Prepare the response
	response := make(map[string]interface{})
	response["code"] = http.StatusOK
	response["success"] = true

	return response, nil
}
