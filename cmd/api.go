package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	// "github.com/golang-jwt/jwt/v5"
)

type APIServer struct {
	ListenAddr 	string
	Store		Storage	
}

type APIError struct {
	Error		string		`json:"error"`
}

type APIFunc func(http.ResponseWriter, *http.Request) error

func createNewServer(listenAddr string,store Storage) *APIServer {
	return &APIServer{
		ListenAddr: listenAddr,
		Store: store,
	}
}

func (s *APIServer) Run(){
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	log.Println("JSON API running on port ",s.ListenAddr)

	r.Get("/create",makeHTTPHandlerFunc(s.handleCreateAccount))

	http.ListenAndServe(":3000", r)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter,r *http.Request) error {
	createAccReq := new(CreateAccReq)
	if err := json.NewDecoder(r.Body).Decode(createAccReq); err != nil {
		return err
	}
	fmt.Println("Reach 1")

	account, err := NewAccount(createAccReq.Email,createAccReq.Password,createAccReq.Username)
	if err != nil {
		return err
	}
	fmt.Println("Reach 2")

	if err := s.Store.CreateAccount(account); err != nil {
		return err
	}
	fmt.Println("Reach 3")

	return WriteJSON(w,http.StatusOK,account)
}

func makeHTTPHandlerFunc(f APIFunc) http.HandlerFunc{
	return func(w http.ResponseWriter,r *http.Request){
		if err := f(w,r); err != nil{
			//error response
			WriteJSON(w,http.StatusBadRequest,APIError{Error: err.Error()})
		}
	}
}

func WriteJSON(w http.ResponseWriter, status int,v any) error {
	w.Header().Add("Content-Type","application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}