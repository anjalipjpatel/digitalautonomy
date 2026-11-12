package api

import "net/http"

// HealthHandler godoc
// @Summary Check API health
// @Description Returns OK when the API is running.
// @Tags health
// @Produce text/plain
// @Success 200 {string} string "OK"
// @Router /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
