package ngauth

import (

	// mysql dialect for gorm (wrapper for go-sql-driver)
	_ "github.com/jinzhu/gorm/dialects/mysql"

	//pgsql dialet for gorm
	_ "github.com/jinzhu/gorm/dialects/postgres"

	//mssql dialet for gorm
	_ "github.com/jinzhu/gorm/dialects/mssql"

	"github.com/spf13/viper"
)

// Map is a convenient alias for a map[string]interface{} map, useful for
// dealing with JSON/BSON in a native way.  For instance:
//
//     Map{"a": 1, "b": true}
type Map map[string]interface{}

// Configuration holds configuration variables
type Configuration struct {
	//Port to listen to
	Port string

	//DBConnectionString - Connection String to database
	DBConnectionString string

	//DBDriver for database/sql,eg values mysql,postgres,mssql,sqlite3
	DBDriver           string
	DBPoolMaxIdleConns int
	DBPoolMaxOpenConns int

	UsersTableName    string
	OTPTableName      string
	SessionsTableName string

	//smtp
	SMTPHost     string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string

	//SignKey for generating JWT Tokens
	SignKey              []byte
	JWTAccessExpireMins  int
	JWTRefreshExpireMins int

	//otp
	OTPExpireTime int64
	OTPBanTime    int64
	OTPFindTime   int64
	OTPMaxRetry   int

	//only register verified users
	VerifyBeforeRegister bool

	//proxy
	UpstreamPublicURL  string
	UpstreamPrivateURL string
}

// Config holds configuration variables
var Config *Configuration

// ParseConfig parses environment variables to configuration
func ParseConfig(inConfig *Configuration) {

	LogInfo("Config: parsing config started")

	viper.AutomaticEnv()

	//read config file
	viper.SetConfigName(".config") // name of config file (without extension)
	viper.AddConfigPath(".")       // look for config in the working directory
	err := viper.ReadInConfig()    // Find and read the config file
	if err != nil {                // Handle errors reading the config file
		LogInfof("Config: error reading config file: %s \n", err)
	}

	viper.SetDefault("PORT", "8080")

	viper.SetDefault("DB_CONNECTION_STRING", "test:test@tcp(127.0.0.1:3306)/mydb?charset=utf8&parseTime=True&loc=Local")
	viper.SetDefault("DB_DRIVER", "mysql")
	viper.SetDefault("DB_POOL_MAX_IDLE_CONNS", "-1")
	viper.SetDefault("DB_POOL_MAX_OPEN_CONNS", "-1")

	viper.SetDefault("USERS_TABLE_NAME", "users")
	viper.SetDefault("OTP_TABLE_NAME", "otp")
	viper.SetDefault("SESSIONS_TABLE_NAME", "sessions")

	viper.SetDefault("OTP_EXPIRE_TIME", "300") //default 5mins
	viper.SetDefault("OTP_BAN_TIME", "300")    //default 5mins
	viper.SetDefault("OTP_FIND_TIME", "300")   //default 5mins
	viper.SetDefault("OTP_MAX_RETRY", "3")

	//at least 32 byte long for security
	viper.SetDefault("SIGN_KEY", "g4k591b582367a97acd7d1e5dc260729")
	viper.SetDefault("JWT_ACCESS_EXPIRE_MINS", "15")
	viper.SetDefault("JWT_REFRESH_EXPIRE_MINS", "1440")
	viper.SetDefault("VERIFY_BEFORE_REGISTER", "true")

	//############### GET VALUES FROM ENV
	inConfig.Port = viper.GetString("PORT")
	inConfig.DBConnectionString = viper.GetString("DB_CONNECTION_STRING")
	inConfig.DBDriver = viper.GetString("DB_DRIVER")
	inConfig.DBPoolMaxIdleConns = viper.GetInt("DB_POOL_MAX_IDLE_CONNS")
	inConfig.DBPoolMaxOpenConns = viper.GetInt("DB_POOL_MAX_OPEN_CONNS")

	inConfig.UsersTableName = viper.GetString("USERS_TABLE_NAME")
	inConfig.OTPTableName = viper.GetString("OTP_TABLE_NAME")
	inConfig.SessionsTableName = viper.GetString("SESSIONS_TABLE_NAME")

	inConfig.OTPExpireTime = viper.GetInt64("OTP_EXPIRE_TIME")
	inConfig.OTPBanTime = viper.GetInt64("OTP_BAN_TIME")
	inConfig.OTPFindTime = viper.GetInt64("OTP_FIND_TIME")
	inConfig.OTPMaxRetry = viper.GetInt("OTP_MAX_RETRY")

	inConfig.SignKey = []byte(viper.GetString("SIGN_KEY"))
	inConfig.JWTAccessExpireMins = viper.GetInt("JWT_ACCESS_EXPIRE_MINS")
	inConfig.JWTRefreshExpireMins = viper.GetInt("JWT_REFRESH_EXPIRE_MINS")
	inConfig.VerifyBeforeRegister = viper.GetBool("VERIFY_BEFORE_REGISTER")

	//proxy
	inConfig.UpstreamPublicURL = viper.GetString("UPSTREAM_PUBLIC_URL")
	inConfig.UpstreamPrivateURL = viper.GetString("UPSTREAM_PRIVATE_URL")

	LogInfo("Config: parsing config completed")

}

// SetConfig - set Config for the package
func SetConfig(inConfig *Configuration) {
	Config = inConfig
}
