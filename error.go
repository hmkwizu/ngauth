package ngauth

// Error - error returned from API
type Error struct {
	Code    int
	Message string
}

// Language codes
const (
	LanguageEN = "en"
	LanguageSW = "sw"
	LanguageTR = "tr"
)

var supportedLanguages = [...]string{LanguageEN, LanguageSW, LanguageTR}

// API Error codes
const (
	//external errors, eg db, server etc
	ErrorInternalServerError = 1001
	ErrorBadRequest          = 1002
	ErrorDBError             = 1003
	ErrorBackendServerError  = 1004
	ErrorNotAuthorized       = 1005

	//API errors
	ErrorNotFound                       = 2001
	ErrorEmptyFields                    = 2002
	ErrorPasswordsDoNotMatch            = 2003
	ErrorUsernameExists                 = 2004
	ErrorIncorrectUsernameOrPassword    = 2005
	ErrorIncorrectPhoneNumberOrPassword = 2006
	ErrorIncorrectEmailOrPassword       = 2007
	ErrorMissingFunctionParams          = 2008

	//API errors - validations
	ErrorInvalidEmail       = 2009
	ErrorInvalidPhoneNumber = 2010
	ErrorInvalidCountryCode = 2011
	ErrorInvalidOTPCode     = 2012
	ErrorExpiredOTPCode     = 2013

	ErrorAlreadyVerified  = 2014
	ErrorInvalidToken     = 2015
	ErrorGetVerifiedFirst = 2016
	ErrorWrongValueFor    = 2017
	ErrorUserNotFound     = 2018
	ErrorWaitFor          = 2019
)

var errorText = map[int]map[string]string{
	ErrorDBError:             map[string]string{LanguageEN: "Database error", LanguageSW: "Database error", LanguageTR: "Veri tabanı hatası"},
	ErrorBadRequest:          map[string]string{LanguageEN: "Bad Request", LanguageSW: "Bad Request", LanguageTR: "Bad Request"},
	ErrorInternalServerError: map[string]string{LanguageEN: "Internal Server Error", LanguageSW: "Internal Server Error", LanguageTR: "İç Sunucu Hatası"},
	ErrorBackendServerError:  map[string]string{LanguageEN: "Backend Server Error", LanguageSW: "Backend Server Error", LanguageTR: "Dış Sunucu Hatası"},
	ErrorNotAuthorized:       map[string]string{LanguageEN: "Not Authorized", LanguageSW: "Hauna ruhusa", LanguageTR: "Not Authorized"},

	ErrorEmptyFields:         map[string]string{LanguageEN: "Empty Field(s)", LanguageSW: "Empty Field(s)", LanguageTR: "Boş alanları doldur"},
	ErrorPasswordsDoNotMatch: map[string]string{LanguageEN: "Passwords do not match", LanguageSW: "Passwords do not match", LanguageTR: "Parolalar uyuşmuyor"},
	ErrorUsernameExists:      map[string]string{LanguageEN: "User Exists", LanguageSW: "User Exists", LanguageTR: "Kullanıcı adı var"},

	ErrorIncorrectUsernameOrPassword:    map[string]string{LanguageEN: "Incorrect username or password", LanguageSW: "Jina la mtumiaji au neno la siri sio sahihi", LanguageTR: "Kullanıcı adı veya şifre yanlış"},
	ErrorIncorrectPhoneNumberOrPassword: map[string]string{LanguageEN: "Incorrect phone number or password", LanguageSW: "Namba ya simu au neno la siri sio sahihi", LanguageTR: "Telefon numara veya şifre yanlış"},
	ErrorIncorrectEmailOrPassword:       map[string]string{LanguageEN: "Incorrect email or password", LanguageSW: "Barua pepe au neno la siri sio sahihi", LanguageTR: "e-posta veya şifre yanlış"},
	ErrorMissingFunctionParams:          map[string]string{LanguageEN: "Missing function parameter(s)", LanguageSW: "Missing function parameter(s)", LanguageTR: "Missing function parameter(s)"},
	ErrorNotFound:                       map[string]string{LanguageEN: "Not Found", LanguageSW: "Not Found", LanguageTR: "Bulunamadı"},

	ErrorInvalidEmail:       map[string]string{LanguageEN: "Please enter a valid email", LanguageSW: "Tafadhali ingiza barua pepe iliyo sahihi", LanguageTR: "Please enter a valid email"},
	ErrorInvalidPhoneNumber: map[string]string{LanguageEN: "Please enter a valid phone number", LanguageSW: "Tafadhali ingiza namba ya simu iliyo sahihi", LanguageTR: "Please enter a valid phone number"},
	ErrorInvalidCountryCode: map[string]string{LanguageEN: "Please enter a valid country code", LanguageSW: "Tafadhali ingiza namba ya nchi iliyo sahihi", LanguageTR: "Please enter a valid country code"},
	ErrorInvalidOTPCode:     map[string]string{LanguageEN: "Please enter a valid OTP code", LanguageSW: "Tafadhali ingiza OTP sahihi", LanguageTR: "Please enter a valid OTP code"},
	ErrorExpiredOTPCode:     map[string]string{LanguageEN: "Expired OTP!", LanguageSW: "OTP imeisha muda wake!", LanguageTR: "Kodun süresi doldu!"},
	ErrorAlreadyVerified:    map[string]string{LanguageEN: "Already verified!", LanguageSW: "Tayari umethibitishwa!", LanguageTR: "Already verified!"},
	ErrorInvalidToken:       map[string]string{LanguageEN: "Invalid token!", LanguageSW: "Token batili!", LanguageTR: "Geçersiz token!"},
	ErrorGetVerifiedFirst:   map[string]string{LanguageEN: "Get verified first!", LanguageSW: "Tafadhali kamilisha uthibitisho kwanza!", LanguageTR: "önce doğrulanmalısın!"},
	ErrorWrongValueFor:      map[string]string{LanguageEN: "Wrong value for: ", LanguageSW: "Wrong value for: ", LanguageTR: "Yanlış değer: "},
	ErrorUserNotFound:       map[string]string{LanguageEN: "User Not Found", LanguageSW: "User Not Found", LanguageTR: "Kullanıcı Bulunamadı"},
	ErrorWaitFor:            map[string]string{LanguageEN: "Please wait for", LanguageSW: "Tafadhali subiri kwa", LanguageTR: "Lütfen bekleyin"},
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
