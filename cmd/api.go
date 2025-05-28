package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	// "fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
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
	r.Post("/login",makeHTTPHandlerFunc(s.handleLogin))
	r.Get("/account/{id}",withJWTAuth(makeHTTPHandlerFunc(s.handleGetAccount),s.Store))
	http.ListenAndServe(":3000", r)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter,r *http.Request) error {
	createAccReq := new(CreateAccReq)
	if err := json.NewDecoder(r.Body).Decode(createAccReq); err != nil {
		return err
	}

	account, err := NewAccount(createAccReq.Email,createAccReq.Password,createAccReq.Username)
	if err != nil {
		return err
	}

	if err := s.Store.CreateAccount(account); err != nil {
		return err
	}

	return WriteJSON(w,http.StatusOK,account)
}

func (s *APIServer) handleLogin(w http.ResponseWriter,r *http.Request) error {
	loginreq := new(LoginAccReq)
	if err := json.NewDecoder(r.Body).Decode(loginreq); err != nil {
		return err
	}
	if err := s.Store.Login(loginreq); err != nil {
		return err
	}
	tokenString,err := createJWT(loginreq)
	if err != nil {
		return err
	}
	w.Header().Set("jwt",tokenString)
	return WriteJSON(w,http.StatusOK,tokenString)
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter,r *http.Request) error {
	account := r.Context().Value("account").(*Account)
	return WriteJSON(w, http.StatusOK, account)
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

func createJWT(loginreq *LoginAccReq) (string,error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,jwt.MapClaims{
		"issued_time" 	: time.Now().Unix(),
		"email"			: loginreq.Email,
	})
	return  token.SignedString([]byte("Test"))
}

func validateJWT(JWT string) (*jwt.Token,error) {
	token,err := jwt.Parse(JWT,func(t *jwt.Token) (interface{}, error) {
		return []byte("Test"),nil
	},jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return token,err
	}
	return token,nil
}

func withJWTAuth(handlerFunc http.HandlerFunc,s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter,r *http.Request) {
		fmt.Println("Executing JWT middleware")
		tokenString := r.Header.Get("Jwt")
		token,err := validateJWT(tokenString)
		if err != nil || !token.Valid{
			permissionDenied(w)
			fmt.Println("JWT is invalid")
			return 
		}

		emailID := chi.URLParam(r,"id")

		account,err := s.GetAccountByName(emailID)
		if(err != nil){
			WriteJSON(w,http.StatusBadRequest,APIError{Error: "No user exists in the db with that id"})
			return
		}
		claims := token.Claims.(jwt.MapClaims)
		fmt.Println(claims)
		if account.Email != string(claims["email"].(string)){
			fmt.Println("Wrong user trying to access data")
		}
		ctx := context.WithValue(r.Context(), "account", account)
		r = r.WithContext(ctx)
		handlerFunc(w, r)
	}
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w,http.StatusForbidden,APIError{Error: "Access Denied"})
}