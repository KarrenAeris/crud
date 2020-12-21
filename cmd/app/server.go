package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"

	"github.com/KarrenAeris/crud/pkg/customers"
)

// Методы запросов
const (
	GET    = "GET"    // Метод получения GET
	POST   = "POST"   // Метод отправ/обновлении POST
	DELETE = "DELETE" // Метод удаления DELETE
)

//Server ...
type Server struct {
	mux         *mux.Router
	customerSvc *customers.Service
}

//NewServer ...
func NewServer(m *mux.Router, customerSvc *customers.Service) *Server {
	return &Server{mux: m, customerSvc: customerSvc}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

//Init инициализирует сервер (регистрирует все Handler'ы)
func (s *Server) Init() {
	s.mux.HandleFunc("/customers", s.handleGetAllCustomers).Methods(GET)
	s.mux.HandleFunc("/customers/{id:[0-9]+}", s.handleGetCustomerByID).Methods(GET)
	s.mux.HandleFunc("/api/customers", s.handleSave).Methods(POST)
	s.mux.HandleFunc("/customers/active", s.handleGetAllActiveCustomers).Methods(GET)
	s.mux.HandleFunc("/customers/{id:[0-9]+}/block", s.handleBlockByID).Methods(POST)
	s.mux.HandleFunc("/customers/{id:[0-9]+}/block", s.handleUnBlockByID).Methods(DELETE)
	s.mux.HandleFunc("/customers/{id:[0-9]+}", s.handleDelete).Methods(DELETE)

	s.mux.HandleFunc("/api/customers/token", s.handleCreateToken).Methods(POST)
	s.mux.HandleFunc("/api/customers/token/validate", s.handleValidateToken).Methods(POST)
	// s.mux.Use(middleware.Base(s.customerSvc.Auth))
}

// хендлер метод для извлечения всех клиентов
func (s *Server) handleGetAllCustomers(w http.ResponseWriter, r *http.Request) {

	items, err := s.customerSvc.All(r.Context())
	if err != nil {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	respondJSON(w, items)
}

// хендлер метод для извлечения всех активных клиентов
func (s *Server) handleGetAllActiveCustomers(w http.ResponseWriter, r *http.Request) {

	items, err := s.customerSvc.AllActive(r.Context())
	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	respondJSON(w, items)
}

func (s *Server) handleGetCustomerByID(w http.ResponseWriter, r *http.Request) {
	//получаем ID из параметра запроса
	// idP := r.URL.Query().Get("id")
	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// переобразуем его в число
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	item, err := s.customerSvc.ByID(r.Context(), id)
	if errors.Is(err, customers.ErrNotFound) {
		errorWriter(w, http.StatusNotFound, err)
		return
	}

	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	respondJSON(w, item)
}

func (s *Server) handleBlockByID(w http.ResponseWriter, r *http.Request) {
	//получаем ID из параметра запроса
	// idP := r.URL.Query().Get("id")
	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	item, err := s.customerSvc.ChangeActive(r.Context(), id, false)
	if errors.Is(err, customers.ErrNotFound) {
		errorWriter(w, http.StatusNotFound, err)
		return
	}

	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	respondJSON(w, item)
}

func (s *Server) handleUnBlockByID(w http.ResponseWriter, r *http.Request) {
	//получаем ID из параметра запроса
	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	item, err := s.customerSvc.ChangeActive(r.Context(), id, true)
	if errors.Is(err, customers.ErrNotFound) {
		errorWriter(w, http.StatusNotFound, err)
		return
	}

	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	respondJSON(w, item)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	//получаем ID из параметра запроса
	// idP := r.URL.Query().Get("id")
	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	item, err := s.customerSvc.Delete(r.Context(), id)
	if errors.Is(err, customers.ErrNotFound) {
		errorWriter(w, http.StatusNotFound, err)
		return
	}

	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	respondJSON(w, item)
}

func (s *Server) handleSave(w http.ResponseWriter, r *http.Request) {

	var item *customers.Customer

	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(item.Password), bcrypt.DefaultCost)
	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	item.Password = string(hashed)

	customer, err := s.customerSvc.Save(r.Context(), item)
	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	respondJSON(w, customer)
}

func (s *Server) handleCreateToken(w http.ResponseWriter, r *http.Request) {
	var item *struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	token, err := s.customerSvc.TokenForCustomer(r.Context(), item.Login, item.Password)

	if err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	respondJSON(w, map[string]interface{}{"status": "ok", "token": token})
}

func (s *Server) handleValidateToken(w http.ResponseWriter, r *http.Request) {
	var item *struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	id, err := s.customerSvc.AuthenticateCustomer(r.Context(), item.Token)

	if err != nil {
		status := http.StatusInternalServerError
		text := "internal error"

		if err == customers.ErrNoSuchUser {
			status = http.StatusNotFound
			text = "not found"
		}
		if err == customers.ErrExpireToken {
			status = http.StatusBadRequest
			text = "expired"
		}

		respondJSONWithCode(w, status, map[string]interface{}{"status": "fail", "reason": text})
		return
	}

	res := make(map[string]interface{})
	res["status"] = "ok"
	res["customerId"] = id

	respondJSONWithCode(w, http.StatusOK, res)
}

func errorWriter(w http.ResponseWriter, httpSts int, err error) {
	log.Print(err)
	http.Error(w, http.StatusText(httpSts), httpSts)
}

func respondJSON(w http.ResponseWriter, iData interface{}) {
	data, err := json.Marshal(iData)

	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	if err != nil {
		log.Print(err)
	}
}

func respondJSONWithCode(w http.ResponseWriter, sts int, iData interface{}) {
	data, err := json.Marshal(iData)

	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(sts)

	_, err = w.Write(data)
	if err != nil {
		log.Print(err)
	}
}
