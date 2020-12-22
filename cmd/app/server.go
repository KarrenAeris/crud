package app

import (
	"encoding/json"

	"log"
	"net/http"


	"github.com/gorilla/mux"

	"github.com/KarrenAeris/crud/cmd/app/middleware"
	"github.com/KarrenAeris/crud/pkg/customers"
	"github.com/KarrenAeris/crud/pkg/managers"

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
	managerSvc  *managers.Service
}

//NewServer ...
func NewServer(m *mux.Router, cSvc *customers.Service, mSvc *managers.Service) *Server {
	return &Server{
		mux:         m,
		customerSvc: cSvc,
		managerSvc:  mSvc,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

//Init инициализирует сервер (регистрирует все Handler'ы)
func (s *Server) Init() {
	customersAuthenticateMd := middleware.Authenticate(s.customerSvc.IDByToken)
	customersSubrouter := s.mux.PathPrefix("/api/customers").Subrouter()
	customersSubrouter.Use(customersAuthenticateMd)

	customersSubrouter.HandleFunc("", s.handleCustomerRegistration).Methods("POST")
	customersSubrouter.HandleFunc("/token", s.handleCustomerGetToken).Methods("POST")
	customersSubrouter.HandleFunc("/products", s.handleCustomerGetProducts).Methods("GET")

	managersAuthenticateMd := middleware.Authenticate(s.managerSvc.IDByToken)
	managersSubRouter := s.mux.PathPrefix("/api/managers").Subrouter()
	managersSubRouter.Use(managersAuthenticateMd)
	managersSubRouter.HandleFunc("", s.handleManagerRegistration).Methods("POST")
	managersSubRouter.HandleFunc("/token", s.handleManagerGetToken).Methods("POST")
	managersSubRouter.HandleFunc("/sales", s.handleManagerGetSales).Methods("GET")
	managersSubRouter.HandleFunc("/sales", s.handleManagerMakeSales).Methods("POST")
	managersSubRouter.HandleFunc("/products", s.handleManagerGetProducts).Methods("GET")
	managersSubRouter.HandleFunc("/products", s.handleManagerChangeProducts).Methods("POST")
	managersSubRouter.HandleFunc("/products/{id:[0-9]+}", s.handleManagerRemoveProductByID).Methods("DELETE")
	managersSubRouter.HandleFunc("/customers", s.handleManagerGetCustomers).Methods("GET")
	managersSubRouter.HandleFunc("/customers", s.handleManagerChangeCustomer).Methods("POST")
	managersSubRouter.HandleFunc("/customers/{id:[0-9]+}", s.handleManagerRemoveCustomerByID).Methods("DELETE")
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
