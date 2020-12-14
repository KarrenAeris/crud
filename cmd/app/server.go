package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/KarrenAeris/crud/pkg/customers"
	"github.com/KarrenAeris/crud/cmd/app/middleware"
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
	// s.mux.HandleFunc("/customers.getAll", s.handleGetAllCustomers)
	s.mux.HandleFunc("/customers", s.handleGetAllCustomers).Methods(GET)
	// s.mux.HandleFunc("/customers.getById", s.handleGetCustomerByID)
	s.mux.HandleFunc("/customers/{id:[0-9]+}", s.handleGetCustomerByID).Methods(GET)
	// s.mux.HandleFunc("/customers.save", s.handleSave)
	s.mux.HandleFunc("/customers", s.handleSave).Methods(POST)
	s.mux.HandleFunc("/customers/active", s.handleGetAllActiveCustomers).Methods(GET)
	s.mux.HandleFunc("/customers/{id:[0-9]+}/block", s.handleBlockByID).Methods(POST)
	s.mux.HandleFunc("/customers/{id:[0-9]+}/block", s.handleUnBlockByID).Methods(DELETE)
	s.mux.HandleFunc("/customers/{id:[0-9]+}", s.handleDelete).Methods(DELETE)

	s.mux.Use(middleware.Base(s.customerSvc.Auth))
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
		//вызываем фукцию для ответа с ошибкой
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
	//если получили ошибку то отвечаем с ошибкой
	if err != nil {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	//получаем баннер из сервиса
	item, err := s.customerSvc.ByID(r.Context(), id)

	if errors.Is(err, customers.ErrNotFound) {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusNotFound, err)
		return
	}

	//если получили ошибку то отвечаем с ошибкой
	if err != nil {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}
	//вызываем функцию для ответа в формате JSON
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
	// переобразуем его в число
	id, err := strconv.ParseInt(idParam, 10, 64)
	//если получили ошибку то отвечаем с ошибкой
	if err != nil {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	//получаем баннер из сервиса
	item, err := s.customerSvc.ChangeActive(r.Context(), id, false)

	if errors.Is(err, customers.ErrNotFound) {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusNotFound, err)
		return
	}

	//если получили ошибку то отвечаем с ошибкой
	if err != nil {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}
	//вызываем функцию для ответа в формате JSON
	respondJSON(w, item)
}

func (s *Server) handleUnBlockByID(w http.ResponseWriter, r *http.Request) {
	//получаем ID из параметра запроса
	idParam, ok := mux.Vars(r)["id"]
	if !ok {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// переобразуем его в число
	id, err := strconv.ParseInt(idParam, 10, 64)
	//если получили ошибку то отвечаем с ошибкой
	if err != nil {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	//получаем баннер из сервиса
	item, err := s.customerSvc.ChangeActive(r.Context(), id, true)

	if errors.Is(err, customers.ErrNotFound) {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusNotFound, err)
		return
	}

	//если получили ошибку то отвечаем с ошибкой
	if err != nil {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}
	//вызываем функцию для ответа в формате JSON
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

	// переобразуем его в число
	id, err := strconv.ParseInt(idParam, 10, 64)
	//если получили ошибку то отвечаем с ошибкой
	if err != nil {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusBadRequest, err)
		return
	}

	//получаем баннер из сервиса
	item, err := s.customerSvc.Delete(r.Context(), id)

	if errors.Is(err, customers.ErrNotFound) {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusNotFound, err)
		return
	}

	//если получили ошибку то отвечаем с ошибкой
	if err != nil {
		//вызываем фукцию для ответа с ошибкой
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}
	//вызываем функцию для ответа в формате JSON
	respondJSON(w, item)
}

func (s *Server) handleSave(w http.ResponseWriter, r *http.Request) {
	/*
		ip := r.FormValue("id")
		name := r.FormValue("name")
		phone := r.FormValue("phone")

		id, err := strconv.Atoi(ip)
		if err != nil {
			errorWriter(w, http.StatusBadRequest, err)
			return
		}
		//проверка на пустату
		if name == "" && phone == "" {
			errorWriter(w, http.StatusBadRequest, err)
			return
		}
	*/

	var item *customers.Customer
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		log.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	customer, err := s.customerSvc.Save(r.Context(), item)

	if err != nil {
		errorWriter(w, http.StatusInternalServerError, err)
		return
	}
	respondJSON(w, customer)
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
