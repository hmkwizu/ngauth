package ngauth

// Error - error returned from API
type Error struct {
	Code    int
	Message string
}

// Language codes
const (
	LanguageEN = "en"
	LanguageTR = "tr"
)

var supportedLanguages = [...]string{LanguageEN, LanguageTR}

// API Error codes
const (
	//external errors, eg db, server etc
	ErrorInternalServerError = 1001
	ErrorBadRequest          = 1002
	ErrorDBError             = 1003
	ErrorBackendServerError  = 1004

	//API errors
	ErrorNotFound                    = 2001
	ErrorEmptyFields                 = 2002
	ErrorPasswordsDoNotMatch         = 2003
	ErrorUsernameExists              = 2004
	ErrorIncorrectUsernameOrPassword = 2005
	ErrorMissingFunctionParams       = 2006

	//API errors - validations
	ErrorInvalidEmail       = 2007
	ErrorInvalidPhoneNumber = 2008
	ErrorInvalidCountryCode = 2009
	ErrorInvalidOTPCode     = 2010
	ErrorExpiredOTPCode     = 2011

	ErrorAlreadyVerified  = 2012
	ErrorInvalidToken     = 2013
	ErrorGetVerifiedFirst = 2014
	ErrorWrongValueFor    = 2015
	ErrorUserNotFound     = 2016
	ErrorWaitFor          = 2017
)

var errorText = map[int]map[string]string{
	ErrorDBError:             map[string]string{LanguageEN: "Database error", LanguageTR: "Veri tabanı hatası"},
	ErrorInternalServerError: map[string]string{LanguageEN: "Internal Server Error", LanguageTR: "İç Sunucu Hatası"},
	ErrorBackendServerError:  map[string]string{LanguageEN: "Backend Server Error", LanguageTR: "Dış Sunucu Hatası"},

	ErrorEmptyFields:         map[string]string{LanguageEN: "Empty Field(s)", LanguageTR: "Boş alanları doldur"},
	ErrorPasswordsDoNotMatch: map[string]string{LanguageEN: "Passwords do not match", LanguageTR: "Parolalar uyuşmuyor"},
	ErrorUsernameExists:      map[string]string{LanguageEN: "Username Exists", LanguageTR: "Kullanıcı adı var"},

	ErrorIncorrectUsernameOrPassword: map[string]string{LanguageEN: "Incorrect username or password", LanguageTR: "Kullanıcı adı veya şifre yanlış"},
	ErrorMissingFunctionParams:       map[string]string{LanguageEN: "Missing function parameter(s)", LanguageTR: "Missing function parameter(s)"},
	ErrorNotFound:                    map[string]string{LanguageEN: "Not Found", LanguageTR: "Bulunamadı"},

	ErrorInvalidEmail:       map[string]string{LanguageEN: "Please enter a valid email", LanguageTR: "Please enter a valid email"},
	ErrorInvalidPhoneNumber: map[string]string{LanguageEN: "Please enter a valid phone number", LanguageTR: "Please enter a valid phone number"},
	ErrorInvalidCountryCode: map[string]string{LanguageEN: "Please enter a valid country code", LanguageTR: "Please enter a valid country code"},
	ErrorInvalidOTPCode:     map[string]string{LanguageEN: "Please enter a valid OTP code", LanguageTR: "Please enter a valid OTP code"},
	ErrorExpiredOTPCode:     map[string]string{LanguageEN: "Expired OTP!", LanguageTR: "Kodun süresi doldu!"},
	ErrorAlreadyVerified:    map[string]string{LanguageEN: "Already verified!", LanguageTR: "Already verified!"},
	ErrorInvalidToken:       map[string]string{LanguageEN: "Invalid token!", LanguageTR: "Geçersiz token!"},
	ErrorGetVerifiedFirst:   map[string]string{LanguageEN: "Get verified first!", LanguageTR: "önce doğrulanmalısın!"},
	ErrorWrongValueFor:      map[string]string{LanguageEN: "Wrong value for: ", LanguageTR: "Yanlış değer: "},
	ErrorUserNotFound:       map[string]string{LanguageEN: "User Not Found", LanguageTR: "Kullanıcı Bulunamadı"},
	ErrorWaitFor:            map[string]string{LanguageEN: "Please wait for", LanguageTR: "Lütfen bekleyin"},
}

// ErrorText - returns a text for the API error code. It returns the empty
// string if the errCode is unknown.
func ErrorText(lang string, errCode int) string {
	return errorText[errCode][lang]
}

// NewError - creates an error object for the given error code
func NewError(lang string, errCode int) *Error {
	return &Error{Code: errCode, Message: ErrorText(lang, errCode)}
}

// NewErrorWithMessage - creates an error object with message
func NewErrorWithMessage(errCode int, message string) *Error {
	return &Error{Code: errCode, Message: message}
}
