package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/hmkwizu/ngauth"
)

// config holds configuration variables
var config ngauth.Configuration

// db - Database interface, MUST store pointer to struct
var db ngauth.Database

func routes() *chi.Mux {

	router := chi.NewRouter()

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	router.Use(
		render.SetContentType(render.ContentTypeJSON), // Set content-Type headers as application/json
		middleware.RealIP, // Set req.RemoteAddr correctly even behind proxy
		middleware.Logger, // Log API request calls
		cors.Handler,
		ngauth.LanguageDetector,
		middleware.DefaultCompress, // Compress results, mostly gzipping assets and json
		middleware.RedirectSlashes, // Redirect slashes to no slash URL versions
		middleware.Recoverer,       // Recover from panics without crashing server
	)

	//Not Found handler
	router.NotFound(http.HandlerFunc(ngauth.NotFoundErrorHandler))

	//Method Not Allowed handler
	router.MethodNotAllowed(http.HandlerFunc(ngauth.MethodNotAllowedErrorHandler))

	router.Get("/", IndexHandler)
	router.Get("/health", Health)

	//registration steps
	router.Post("/generate_otp", GenerateOTP)
	router.Post("/verify_otp", VerifyOTP)
	router.Post("/register", Register)

	//login
	router.Post("/login", Login)

	//get a new access token
	router.Post("/token", Token)

	//reset password, use generate_otp and verify_otp prior to this
	router.Post("/reset_password", ResetPassword)

	//change password, token required
	router.Post("/change_password", ChangePassword)

	//push token
	router.Post("/update_push_token", UpdatePushToken)

	//public routes
	router.Route("/pb", func(r chi.Router) {
		r.Get("/*", HandleAllPublic)
		r.Post("/*", HandleAllPublic)
		r.Put("/*", HandleAllPublic)
		r.Delete("/*", HandleAllPublic)
		r.Patch("/*", HandleAllPublic)
		r.Options("/*", HandleAllPublic)
	})

	//private routes - access token authentication done first, then proxy the request
	router.Route("/pt", func(r chi.Router) {
		r.Get("/*", HandleAllPrivate)
		r.Post("/*", HandleAllPrivate)
		r.Put("/*", HandleAllPrivate)
		r.Delete("/*", HandleAllPrivate)
		r.Patch("/*", HandleAllPrivate)
		r.Options("/*", HandleAllPrivate)
	})

	return router
}

func main() {

	ngauth.ParseConfig(&config)
	ngauth.SetConfig(&config)

	//initialize the database
	initDB()

	//create the routes
	router := routes()

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route) // Walk and print out all routes
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}

	ngauth.LogInfo("Server is running on PORT " + config.Port)
	ngauth.LogInfo("DB DRIVER: " + config.DBDriver)

	log.Fatal(http.ListenAndServe(":"+config.Port, router))
}

// initDB opens the database connection
func initDB() {

	//easily swap repository implementation here
	db = &ngauth.SQLRepository{}

	err := db.Init(&config)
	if err != nil {
		log.Println(err)
	}
}

// ###################### http handlers ##############

// GenerateOTP - generates otp and sends it
func GenerateOTP(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.GenerateOTP(db, lang, receivedData, sendOTPCallback)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// VerifyOTP - verifies otp
func VerifyOTP(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.VerifyOTP(db, lang, receivedData)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// Register - registers the user
func Register(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.Register(db, lang, receivedData, hashMake)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// Login - login
func Login(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)
	receivedData["ip_addr"] = r.RemoteAddr
	receivedData["user_agent"] = r.UserAgent()

	response, err := ngauth.Login(db, lang, receivedData, hashCheck)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// ResetPassword - reset user password
func ResetPassword(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.ResetPassword(db, lang, receivedData, hashMake)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// ChangePassword - changes user's password
func ChangePassword(w http.ResponseWriter, r *http.Request) {

	// Validate access token
	accessToken := ngauth.GetTokenFromHeader(r)
	_, err := ngauth.IsValidToken(accessToken)
	if err != nil {
		if err.Code == ngauth.ErrorInvalidToken {
			ngauth.HTTPErrorResponse(w, err.Message, http.StatusUnauthorized)
		} else {
			ngauth.ErrorResponse(w, err.Message, err.Code)
		}
		return
	}

	lang, receivedData := getParams(r)

	response, err := ngauth.ChangePassword(db, lang, receivedData, hashCheck, hashMake)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// Token - token
func Token(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.Token(db, lang, receivedData)
	if err != nil {
		if err.Code == ngauth.ErrorInvalidToken {
			ngauth.HTTPErrorResponse(w, err.Message, http.StatusUnauthorized)
		} else {
			ngauth.ErrorResponse(w, err.Message, err.Code)
		}
		return
	}

	render.JSON(w, r, response)
}

// UpdatePushToken - updates push token
func UpdatePushToken(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)
	receivedData["ip_addr"] = r.RemoteAddr
	receivedData["user_agent"] = r.UserAgent()

	response, err := ngauth.UpdatePushToken(db, lang, receivedData)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

func sendOTPCallback(email string, phoneNo string, code string) {
	ngauth.AsyncSendVerifCode(email, code)
}

func getParams(r *http.Request) (string, map[string]interface{}) {
	//add localization support
	lang := ngauth.LangFromContext(r.Context())

	//parse json body
	var receivedData map[string]interface{}
	json.NewDecoder(r.Body).Decode(&receivedData)

	return lang, receivedData
}

//##### password callbacks
func hashMake(plainPassword string) string {
	return ngauth.BcryptHashMake(plainPassword)
}

func hashCheck(hashedPassword string, plainPassword string) bool {
	return ngauth.BcryptHashCheck(hashedPassword, plainPassword)
}

//##### some default http handlers

// Health - returns health status of the microservice
func Health(w http.ResponseWriter, r *http.Request) {
	//Prepare the response
	response := make(map[string]interface{})
	response["status"] = "running..."

	render.JSON(w, r, response)
}

// IndexHandler - index handler
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	//Prepare the response
	response := make(map[string]interface{})
	response["message"] = "Ok"

	render.JSON(w, r, response)
}

//##### proxy http handlers

// HandleAllPublic - handles all public routes
func HandleAllPublic(w http.ResponseWriter, r *http.Request) {
	//TODO - get lang from getParams
	//we need to re-create another r.Body for the proxy
	lang := "en"

	handleAllUpstream(lang, config.UpstreamPublicURL, w, r)
}

// HandleAllPrivate - handles all private routes
func HandleAllPrivate(w http.ResponseWriter, r *http.Request) {
	//TODO - get lang from getParams
	//we need to re-create another r.Body for the proxy
	lang := "en"

	// Validate access token
	accessToken := ngauth.GetTokenFromHeader(r)
	_, err := ngauth.IsValidToken(accessToken)
	if err != nil {
		if err.Code == ngauth.ErrorInvalidToken {
			ngauth.HTTPErrorResponse(w, err.Message, http.StatusUnauthorized)
		} else {
			ngauth.ErrorResponse(w, err.Message, err.Code)
		}
		return
	}

	handleAllUpstream(lang, config.UpstreamPrivateURL, w, r)
}

func handleAllUpstream(lang string, upstreamURL string, w http.ResponseWriter, r *http.Request) {

	target, err := url.Parse(upstreamURL)

	if err != nil {
		ngauth.ErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//make sure we have a valid url, before proxing
	if ngauth.IsEmptyString(target.Scheme) && ngauth.IsEmptyString(target.Host) {
		ngauth.ErrorResponse(w, ngauth.ErrorText(lang, ngauth.ErrorBackendServerError), ngauth.ErrorBackendServerError)
		return
	}

	//initialize proxy
	proxy := httputil.NewSingleHostReverseProxy(target)

	modifyResponse := func(res *http.Response) error {
		if res.StatusCode == http.StatusBadGateway {
			ngauth.ErrorResponse(w, ngauth.ErrorText(lang, ngauth.ErrorBackendServerError), ngauth.ErrorBackendServerError)
		}
		return nil
	}
	errorHandler := func(wr http.ResponseWriter, req *http.Request, err error) {
		if err != nil {
			ngauth.ErrorResponse(wr, err.Error(), http.StatusInternalServerError)
		}
	}

	proxy.ModifyResponse = modifyResponse
	proxy.ErrorHandler = errorHandler

	proxy.ServeHTTP(w, r)
}
