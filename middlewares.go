package ngauth

import (
	"context"
	"encoding/json"
	"net/http"
)

type key string

const contextKeyLang key = "lang"

//LanguageDetector - checks language from cookie,url query and sets it in context
func LanguageDetector(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//default
		lang := ""

		//read cookie
		var cookie, err = r.Cookie("lang")
		if err == nil {

			// if we have value, set it
			if len(cookie.Value) > 0 {
				lang = cookie.Value
			}
		}

		//url query
		langQuery := r.FormValue("lang")
		if len(lang) == 0 && len(langQuery) > 0 {
			lang = langQuery
		}

		//check if language is supported
		if len(lang) > 0 && !ArrayContains(lang, supportedLanguages[:]) {
			lang = LanguageEN
		}

		//check lang, if still empty, set default
		if len(lang) == 0 {
			lang = LanguageEN
		}

		//add lang to context
		ctx := context.WithValue(r.Context(), contextKeyLang, lang)

		// language is set, continue
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LangFromContext - get lang from context
func LangFromContext(ctx context.Context) string {
	lang, _ := ctx.Value(contextKeyLang).(string)
	return lang
}

// ErrorResponse - writes API error response with http status 200 OK,
// the actual api error code is written in the json body
func ErrorResponse(w http.ResponseWriter, message string, apiErrCode int) {

	jsonBytes := jsonErrorBytes(message, apiErrCode)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

// HTTPErrorResponse - writes an error response with correct http headers ie status, content type etc
func HTTPErrorResponse(w http.ResponseWriter, message string, code int) {

	jsonBytes := jsonErrorBytes(message, code)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(code)
	w.Write(jsonBytes)
}

func jsonErrorBytes(message string, code int) []byte {
	response := make(map[string]interface{})
	response["success"] = false
	response["code"] = code
	response["message"] = message

	jsonString, _ := json.Marshal(response)

	return jsonString
}

// NotFoundErrorHandler - Handler for not found error
func NotFoundErrorHandler(w http.ResponseWriter, r *http.Request) {
	HTTPErrorResponse(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	return
}

// MethodNotAllowedErrorHandler - Handler for method not allowed error
func MethodNotAllowedErrorHandler(w http.ResponseWriter, r *http.Request) {
	HTTPErrorResponse(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	return
}
