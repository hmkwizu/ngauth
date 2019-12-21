package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/hmkwizu/ngauth"
)

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

	return router
}

func main() {

	ngauth.InitConfig()

	//initialize the database
	ngauth.InitDB()

	//create the routes
	router := routes()

	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route) // Walk and print out all routes
		return nil
	}
	if err := chi.Walk(router, walkFunc); err != nil {
		log.Panicf("Logging err: %s\n", err.Error()) // panic if there is an error
	}

	log.Fatal(http.ListenAndServe(":"+ngauth.Config.Port, router))
}

// GenerateOTP - generates otp and sends it
func GenerateOTP(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.GenerateOTP(ngauth.DB, lang, receivedData, sendOTPCallback)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// VerifyOTP - verifies otp
func VerifyOTP(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.VerifyOTP(ngauth.DB, lang, receivedData)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// Register - registers the user
func Register(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.Register(ngauth.DB, lang, receivedData, hashMake)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// Login - login
func Login(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.Login(ngauth.DB, lang, receivedData, hashCheck)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// ResetPassword - reset user password
func ResetPassword(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.ResetPassword(ngauth.DB, lang, receivedData, hashMake)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// ChangePassword - changes user's password
func ChangePassword(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.ChangePassword(ngauth.DB, lang, receivedData, hashCheck, hashMake)
	if err != nil {
		ngauth.ErrorResponse(w, err.Message, err.Code)
		return
	}

	render.JSON(w, r, response)
}

// Token - token
func Token(w http.ResponseWriter, r *http.Request) {

	lang, receivedData := getParams(r)

	response, err := ngauth.Token(ngauth.DB, lang, receivedData)
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
