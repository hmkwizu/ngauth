package ngauth

import (
	"log"

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
	DBDriver string

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
	OTPExpireMins int

	//only register verified users
	VerifyBeforeRegister bool

	//proxy
	BackendPublicURL  string
	BackendPrivateURL string
}

// Config holds configuration variables
var Config Configuration

// DB - Database interface, MUST store pointer to struct
var DB Database

// InitDB opens the database connection
func InitDB() {

	//make sure InitConfig was called
	if len(Config.DBDriver) == 0 {
		InitConfig()
	}

	//easily swap repository implementation here
	DB = &SQLRepository{}

	err := DB.Init(Config)
	if err != nil {
		log.Println(err)
	}

}

// InitConfig initializes configuration variables
func InitConfig() {

	//read config file
	viper.SetConfigName(".config") // name of config file (without extension)
	viper.AddConfigPath(".")       // look for config in the working directory
	err := viper.ReadInConfig()    // Find and read the config file
	if err != nil {                // Handle errors reading the config file
		log.Printf("Fatal error reading config file: %s \n", err)
	}

	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8080")

	viper.SetDefault("DB_CONNECTION_STRING", "test:test@tcp(127.0.0.1:3306)/mydb?charset=utf8&parseTime=True&loc=Local")
	viper.SetDefault("DB_DRIVER", "mysql")

	viper.SetDefault("USERS_TABLE_NAME", "users")
	viper.SetDefault("OTP_TABLE_NAME", "otp")
	viper.SetDefault("SESSIONS_TABLE_NAME", "sessions")

	viper.SetDefault("OTP_EXPIRE_MINS", "5")

	//at least 32 byte long for security
	viper.SetDefault("SIGN_KEY", "g4k591b582367a97acd7d1e5dc260729")
	viper.SetDefault("JWT_ACCESS_EXPIRE_MINS", "15")
	viper.SetDefault("JWT_REFRESH_EXPIRE_MINS", "1440")
	viper.SetDefault("VERIFY_BEFORE_REGISTER", "true")

	//############### GET VALUES FROM ENV
	Config.Port = viper.GetString("PORT")
	Config.DBConnectionString = viper.GetString("DB_CONNECTION_STRING")
	Config.DBDriver = viper.GetString("DB_DRIVER")
	Config.UsersTableName = viper.GetString("USERS_TABLE_NAME")
	Config.OTPTableName = viper.GetString("OTP_TABLE_NAME")
	Config.SessionsTableName = viper.GetString("SESSIONS_TABLE_NAME")

	Config.OTPExpireMins = viper.GetInt("OTP_EXPIRE_MINS")
	Config.SignKey = []byte(viper.GetString("SIGN_KEY"))
	Config.JWTAccessExpireMins = viper.GetInt("JWT_ACCESS_EXPIRE_MINS")
	Config.JWTRefreshExpireMins = viper.GetInt("JWT_REFRESH_EXPIRE_MINS")
	Config.VerifyBeforeRegister = viper.GetBool("VERIFY_BEFORE_REGISTER")

	//proxy
	Config.BackendPublicURL = viper.GetString("BACKEND_PUBLIC_URL")
	Config.BackendPrivateURL = viper.GetString("BACKEND_PRIVATE_URL")

	LogInfo("Server is running on PORT " + Config.Port)
	LogInfo("DB DRIVER: " + Config.DBDriver)

}
