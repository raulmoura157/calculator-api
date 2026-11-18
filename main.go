package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

var errDivByZero = errors.New("division by zero")

type operation func(a, b int) (int, error)

var operations = map[string]operation{
	"sum": func(a, b int) (int, error) { return a + b, nil },
	"sub": func(a, b int) (int, error) { return a - b, nil },
	"mul": func(a, b int) (int, error) { return a * b, nil },
	// Integer division (truncated toward zero), as the spec defines result as int.
	"div": func(a, b int) (int, error) {
		if b == 0 {
			return 0, errDivByZero
		}
		return a / b, nil
	},
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	for name, op := range operations {
		mux.HandleFunc("GET /api/"+name, handle(op))
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
}

func handle(op operation) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		a, errA := strconv.Atoi(q.Get("term_one"))
		b, errB := strconv.Atoi(q.Get("term_two"))
		if errA != nil || errB != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "term_one and term_two must be integers"})
			return
		}
		result, err := op(a, b)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]int{"result": result})
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func main() {
	log.Println("listening on :8000")
	log.Fatal(http.ListenAndServe(":8000", newMux()))
}
