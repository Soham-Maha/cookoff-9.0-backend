package httphelpers

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(writer http.ResponseWriter, status int, v any) {
	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(status)
	err := json.NewEncoder(writer).Encode(map[string]any{
		"status": "success",
		"code":   status,
		"data":   v})
	if err != nil {
		panic(err)
	}

}

func WriteError(w http.ResponseWriter, status int, v any) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(map[string]any{
		"status": "error",
		"code":   status,
		"message": v,
	})
	if err != nil {
		panic(err)
	}
}
