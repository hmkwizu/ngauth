package ngauth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/jordan-wright/email"
	"github.com/nyaruka/phonenumbers"
	"github.com/spf13/cast"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/guregu/null.v3"
)

//regex for email validation
var rgxEmail = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// ExpireAtUTC returns unix timestamp expire_at given a time duration
func ExpireAtUTC(tm time.Duration) int64 {
	return NowTimestamp() + int64(tm.Seconds())
}

// ExpireAtTime returns time.Time expire_at given a time duration
func ExpireAtTime(tm time.Duration) time.Time {
	return TimeNow().Add(tm)
}

// NowTimestamp returns a unix timestamp
func NowTimestamp() int64 {
	return time.Now().UTC().Unix()
}

// TimeNow returns time now
func TimeNow() time.Time {
	return time.Now()
}

// LogInfo - logs a message to stdout
func LogInfo(v ...interface{}) {
	// fmt.Fprintln(os.Stdout, msg)
	log.SetOutput(os.Stdout)
	log.Println(v...)
}

// LogInfof - logs a message to stdout
func LogInfof(format string, v ...interface{}) {
	log.SetOutput(os.Stdout)
	log.Printf(format, v...)
}

// LogError - logs a message to stderr
func LogError(v ...interface{}) {
	// fmt.Fprintln(os.Stderr, msg)
	log.SetOutput(os.Stderr)
	log.Println(v...)
}

// LogErrorf - logs a message to stderr
func LogErrorf(format string, v ...interface{}) {
	log.SetOutput(os.Stdout)
	log.Printf(format, v...)
}

// table - lookup table for the secure number generator
var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}

// SecureRandomNumericString - generates a random numeric code
func SecureRandomNumericString(max int) string {
	b := make([]byte, max)
	n, err := io.ReadAtLeast(rand.Reader, b, max)
	if n != max {
		panic(err)
	}
	for i := 0; i < len(b); i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	return string(b)
}

// SecureRandomNumericStringStandard - generates random numeric code of length 6
func SecureRandomNumericStringStandard() string {
	return SecureRandomNumericString(6)
}

// GetStringOrEmpty - get string or empty
// to be used in post body submissions, be sure val is a string
func GetStringOrEmpty(val interface{}) string {
	if val == nil {
		return ""
	}

	return cast.ToString(val)
}

// GetInt64OrZero - get int64 or zero
// to be used in post body submissions, be sure val is a int64
func GetInt64OrZero(val interface{}) int64 {
	if val == nil {
		return 0
	}

	return cast.ToInt64(val)
}

// GetFloat64OrZero - get float64 or zero
// to be used in post body submissions, be sure val is a float64
func GetFloat64OrZero(val interface{}) float64 {
	if val == nil {
		return 0
	}

	return cast.ToFloat64(val)
}

// GetBoolOrFalse - get bool or false
// to be used in post body submissions, be sure val is a float64
func GetBoolOrFalse(val interface{}) bool {
	if val == nil {
		return false
	}

	return cast.ToBool(val)
}

// GetNullableFloat - get a nullable float
func GetNullableFloat(val interface{}) null.Float {
	if val == nil {
		return null.NewFloat(0, false)
	}

	newVal := cast.ToFloat64(val)

	return null.FloatFrom(newVal)
}

// GetNullableInt - get a nullable int
func GetNullableInt(val interface{}) null.Int {
	if val == nil {
		return null.NewInt(0, false)
	}

	newVal := cast.ToInt64(val)

	return null.IntFrom(newVal)
}

// GetNullableBool - get a nullable bool
func GetNullableBool(val interface{}) null.Bool {
	if val == nil {
		return null.NewBool(false, false)
	}

	newVal := cast.ToBool(val)

	return null.BoolFrom(newVal)
}

// GetNullableString - get a nullable string
func GetNullableString(val interface{}) null.String {
	if val == nil {
		return null.NewString("", false)
	}

	newVal := cast.ToString(val)

	return null.StringFrom(newVal)
}

// GetNullableTime - get a nullable time
func GetNullableTime(val interface{}) null.Time {
	if val == nil {
		return null.NewTime(time.Time{}, false)
	}

	newVal := cast.ToTime(val)

	return null.TimeFrom(newVal)
}

// ArrayContains - checks if an array contains a string
func ArrayContains(lookup string, arr []string) bool {

	for _, val := range arr {
		if val == lookup {
			return true
		}
	}
	return false
}

// IsEmptyString - checks if string is empty or not, no trimming of whitespace
// for whitespace trimming use IsEmptyTextContent instead
func IsEmptyString(s string) bool {
	if len(s) == 0 {
		return true
	}
	return false
}

// IsEmptyTextContent - checks whether string is empty or contains only whitespace
func IsEmptyTextContent(s string) bool {
	if len(s) == 0 {
		return true
	}

	r := []rune(s)
	l := len(r)

	for l > 0 {
		l--
		if !unicode.IsSpace(r[l]) {
			return false
		}
	}

	return true
}

// GenerateUUID - generate uuid
func GenerateUUID() string {
	ID := uuid.New()
	return hex.EncodeToString(ID[:])
}

// IsValidEmail - check whether email is valid or not
func IsValidEmail(email string) bool {
	return rgxEmail.MatchString(email)
}

// IsValidPhoneNumber - check whether phone number is valid or not
// returns the number if valid, else nil
func IsValidPhoneNumber(phoneNumber string, countryCode string, lang string) (string, *Error) {

	if IsEmptyString(countryCode) {
		return "", NewError(lang, ErrorInvalidCountryCode)
	}

	num, err := phonenumbers.Parse(phoneNumber, countryCode)
	if err != nil {
		return "", NewErrorWithMessage(ErrorInternalServerError, err.Error())
	}

	if !phonenumbers.IsValidNumber(num) {
		return "", NewError(lang, ErrorInvalidPhoneNumber)
	}

	//return phoneNumber to international format, no spaces
	return phonenumbers.Format(num, phonenumbers.E164), nil
}

// AsyncSendVerifCode - sends verification code in a goroutine
func AsyncSendVerifCode(toEmail string, code string) {
	subject := "Verification Code"
	body := fmt.Sprintf("<p>Your verification code is <b>%s</b>. </p>", code)
	go SendEmail(toEmail, subject, body)
}

// SendEmail - sends emails
func SendEmail(toEmail string, subject string, body string) error {

	//Return if email configuration not complete
	if len(Config.SMTPUsername) == 0 || len(Config.SMTPPassword) == 0 || len(Config.SMTPHost) == 0 || len(Config.SMTPPort) == 0 {
		return errors.New("SMTP Username or Password empty")
	}

	e := email.NewEmail()
	e.From = Config.SMTPFrom
	e.To = []string{toEmail}
	e.Subject = subject
	e.HTML = []byte(body)
	return e.Send(Config.SMTPHost+":"+Config.SMTPPort, smtp.PlainAuth("", Config.SMTPUsername, Config.SMTPPassword, Config.SMTPHost))
}

//######## BCRYPT

// BcryptHashMake - create hashed password
func BcryptHashMake(pwd string) string {

	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

// BcryptHashCheck - bcrypt compare hashed and plain passwords
func BcryptHashCheck(hashedPwd string, plainPwd string) bool {

	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(plainPwd))
	if err != nil {
		return false
	}

	return true
}

//############## JWT

// IsValidToken - check if jwt token is valid
// if no error then token is valid
// return: claims map, error
func IsValidToken(tokenStr string) (map[string]interface{}, *Error) {

	if tokenStr == "" {
		return nil, NewErrorWithMessage(ErrorInvalidToken, "No token found")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return Config.SignKey, nil
	})

	if err != nil {
		if verr, ok := err.(*jwt.ValidationError); ok {
			if verr.Errors&jwt.ValidationErrorExpired > 0 {
				return nil, NewErrorWithMessage(ErrorInvalidToken, "Token expired")
			} else if verr.Errors&jwt.ValidationErrorIssuedAt > 0 {
				return nil, NewErrorWithMessage(ErrorBadRequest, "Token iat invalid")
			} else if verr.Errors&jwt.ValidationErrorNotValidYet > 0 {
				return nil, NewErrorWithMessage(ErrorBadRequest, "Token nbf invalid")
			}
		}
		return nil, NewErrorWithMessage(ErrorBadRequest, err.Error())
	}

	if token == nil || !token.Valid {
		return nil, NewErrorWithMessage(ErrorInvalidToken, err.Error())
	}

	//valid
	var claims jwt.MapClaims
	if tokenClaims, ok := token.Claims.(jwt.MapClaims); ok {
		claims = tokenClaims
	}

	return claims, nil
}

// GenerateAccessToken - generates access token
func GenerateAccessToken(userID interface{}) (string, *Error) {
	return GenerateToken(userID, Config.JWTAccessExpireMins, nil)
}

// GenerateRefreshToken - generates refresh token
func GenerateRefreshToken(userID interface{}) (string, *Error) {
	return GenerateToken(userID, Config.JWTRefreshExpireMins, nil)
}

// GenerateToken - generates signed token
func GenerateToken(userID interface{}, expireMins int, params map[string]interface{}) (string, *Error) {
	//Generate JWT Token
	// NOTE: Don't add sensitive info to the token, eg. password

	claims := jwt.MapClaims{
		"id":  userID,
		"iat": NowTimestamp(),                                       //issued at NOW!
		"exp": ExpireAtUTC(time.Duration(expireMins) * time.Minute), //expires in n minutes
	}

	//add extra params
	if params != nil {
		for k, val := range params {
			if val != nil {
				claims[k] = val
			}
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	//Sign the Token
	tokenString, err := token.SignedString(Config.SignKey)

	if err != nil {
		return tokenString, NewErrorWithMessage(ErrorInternalServerError, err.Error())
	}

	return tokenString, nil
}

// GetTokenFromHeader - gets access_token from header
func GetTokenFromHeader(r *http.Request) string {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 {
		// Bearer token not in proper format
		return ""
	}

	return strings.TrimSpace(splitToken[1])
}
