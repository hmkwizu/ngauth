package ngauth

import (
	"errors"

	"github.com/jinzhu/gorm"
)

// SQLRepository queries the database and returns results to the controller
// Add your methods here for the handler functions your create in the controller
type SQLRepository struct {
	DB *gorm.DB
}

// Init - initialize
func (r *SQLRepository) Init(config *Configuration) error {

	//make sure InitConfig was called
	if len(config.DBDriver) == 0 {
		return errors.New("DB Driver is empty")
	}

	if len(config.DBConnectionString) == 0 {
		return errors.New("DB ConnectionString is empty")
	}

	var err error
	r.DB, err = gorm.Open(config.DBDriver, config.DBConnectionString)

	if err != nil {
		return err
	}

	if err = r.DB.DB().Ping(); err != nil {
		return err
	}

	return nil
}

//############################# Utils #########################

// UpdateRecordByID - update a record by id. the id in the table has to be "id"
// columns is either map[string]interface{} or a struct
func (r *SQLRepository) UpdateRecordByID(tableName string, inRecordID interface{}, columns interface{}, lang string) *Error {

	if inRecordID == nil || len(tableName) == 0 {
		return NewError(lang, ErrorEmptyFields)
	}

	err := r.DB.Table(tableName).Where("id=?", inRecordID).UpdateColumns(columns)

	if err.Error != nil {
		return NewErrorWithMessage(ErrorDBError, err.Error.Error())
	}

	return nil
}

// CreateRecord - create a record
// record is a reference to struct, eg &user
func (r *SQLRepository) CreateRecord(tableName string, record interface{}, lang string) *Error {
	if len(tableName) == 0 || record == nil {
		return NewError(lang, ErrorEmptyFields)
	}

	err := r.DB.Table(tableName).Create(record)

	if err.Error != nil {
		return NewErrorWithMessage(ErrorDBError, err.Error.Error())
	}

	return nil
}

// GetRecord - get single record
func (r *SQLRepository) GetRecord(tableName string, tableColumns string, inRecordID interface{}, resultRecord interface{}, lang string) *Error {

	recordID := GetInt64OrZero(inRecordID)

	if inRecordID == nil || len(tableName) == 0 {
		return NewError(lang, ErrorEmptyFields)
	}

	err := r.DB.Table(tableName).Select(tableColumns).Where("id=?", recordID).Where("deleted_at IS NULL").First(resultRecord)
	//no rows error
	if err.RecordNotFound() {
		return nil
	}
	//any other error
	if err.Error != nil {
		return NewErrorWithMessage(ErrorDBError, err.Error.Error())
	}
	return nil
}

//############################# Users #########################

// UpdateUserByID - updates user by using ID
//columns map[string]interface{}
func (r *SQLRepository) UpdateUserByID(userID interface{}, columns interface{}, lang string) *Error {
	return r.UpdateRecordByID(Config.UsersTableName, userID, columns, lang)
}

// CreateUser - creates a user
func (r *SQLRepository) CreateUser(user User, lang string) (interface{}, *Error) {
	err := r.CreateRecord(Config.UsersTableName, &user, lang)
	return user.ID, err
}

// GetUserBy - get a user by using email/phonenumber
func (r *SQLRepository) GetUserBy(email string, phoneNo string, lang string) (*User, *Error) {

	if IsEmptyString(email) && IsEmptyString(phoneNo) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	//flag to show whether to use email or phonenumber
	useEmail := true
	if !IsEmptyString(phoneNo) {
		useEmail = false
		email = ""
	}

	query := r.DB.Table(Config.UsersTableName).Select("id,email,phone_number,password")

	if useEmail {
		query = query.Where("email=?", email)
	} else {
		query = query.Where("phone_number=?", phoneNo)
	}

	var user User
	err := query.Where("deleted_at IS NULL").First(&user)
	//no rows error
	if err.RecordNotFound() {
		return nil, nil
	}
	//any other error
	if err.Error != nil {
		return nil, NewErrorWithMessage(ErrorDBError, err.Error.Error())
	}
	return &user, nil
}

//############################ OTP #################################

// CreateOTP - creates one time password
func (r *SQLRepository) CreateOTP(otp OTP, lang string) (interface{}, *Error) {

	if len(otp.Code) == 0 {
		return -1, NewError(lang, ErrorEmptyFields)
	}
	err := r.CreateRecord(Config.OTPTableName, &otp, lang)
	return otp.ID, err
}

// GetOTP - get otp by using email/phoneNo
func (r *SQLRepository) GetOTP(email string, phoneNo string, otpFor string, lang string) (*OTP, *Error) {

	if (IsEmptyString(email) && IsEmptyString(phoneNo)) || IsEmptyString(otpFor) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	//flag to show whether to use email or phonenumber
	useEmail := true
	if !IsEmptyString(phoneNo) {
		useEmail = false
		email = ""
	}

	query := r.DB.Table(Config.OTPTableName).Select("*").Where("otp_for=?", otpFor)

	//email
	if useEmail {
		query = query.Where("email=?", email)
	} else {
		//phone number
		query = query.Where("phone_number=?", phoneNo)

	}

	var otp OTP
	err := query.Order("created_at DESC").First(&otp)
	//no rows error
	if err.RecordNotFound() {
		return nil, nil
	}
	//any other error
	if err.Error != nil {
		return nil, NewErrorWithMessage(ErrorDBError, err.Error.Error())
	}
	return &otp, nil
}

// GetOTPs - get otps by using email/phoneNo
func (r *SQLRepository) GetOTPs(email string, phoneNo string, otpFor string, offset int64, limit int64, lang string) ([]OTP, *Error) {

	if (IsEmptyString(email) && IsEmptyString(phoneNo)) || IsEmptyString(otpFor) {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	//flag to show whether to use email or phonenumber
	useEmail := true
	if !IsEmptyString(phoneNo) {
		useEmail = false
		email = ""
	}

	query := r.DB.Table(Config.OTPTableName).Select("*").Where("otp_for=?", otpFor)

	//email
	if useEmail {
		query = query.Where("email=?", email)
	} else {
		//phone number
		query = query.Where("phone_number=?", phoneNo)

	}

	//limit and offset
	if limit > 0 && offset >= 0 {
		query = query.Limit(limit).Offset(offset)
	}

	// select
	results := make([]OTP, 0, 10)
	err := query.Order("created_at DESC").Find(&results)

	//no rows error
	if err.RecordNotFound() {
		return nil, nil
	}
	//any other error
	if err.Error != nil {
		return nil, NewErrorWithMessage(ErrorDBError, err.Error.Error())
	}
	return results, nil
}

// UpdateOTPByID - updates otp by using ID
//columns map[string]interface{}
func (r *SQLRepository) UpdateOTPByID(otpID interface{}, columns interface{}, lang string) *Error {
	return r.UpdateRecordByID(Config.OTPTableName, otpID, columns, lang)
}

//################## Session

// CreateSession - creates a session
func (r *SQLRepository) CreateSession(session Session, lang string) (interface{}, *Error) {

	if len(session.RefreshToken) == 0 {
		return -1, NewError(lang, ErrorEmptyFields)
	}
	err := r.CreateRecord(Config.SessionsTableName, &session, lang)
	return session.ID, err
}

// GetSession - get session by refresh token
func (r *SQLRepository) GetSession(refreshToken string, lang string) (*Session, *Error) {

	if len(refreshToken) == 0 {
		return nil, NewError(lang, ErrorEmptyFields)
	}

	var session Session
	err := r.DB.Table(Config.SessionsTableName).Select("*").Where("refresh_token=?", refreshToken).First(&session)
	//no rows error
	if err.RecordNotFound() {
		return nil, nil
	}
	//any other error
	if err.Error != nil {
		return nil, NewErrorWithMessage(ErrorDBError, err.Error.Error())
	}
	return &session, nil
}
