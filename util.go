package util

import "net/http"
// WebResponse is a wrapper function for returning a web response that includes a standard text message.
func WebResponse(w http.ResponseWriter, r *http.Request, code int, message string){
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Header().Set("Content-Type","application/JSON")
	w.WriteHeader(code)
	w.Write([]byte(message))
	

}