package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/KarrenAeris/crud/cmd/app"
	"github.com/KarrenAeris/crud/pkg/customers"
	_ "github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/dig"
)

func main() {
	// адрес подключения
	// протокол://логи:пароль@хост:порт/бд

	host := "0.0.0.0"
	port := "9999"
	dsn := "postgres://app:pass@localhost:5432/db"

	if err := execute(host, port, dsn); err != nil {
		log.Print(err)
		return
	}

}

func execute(host, port, dsn string) (err error) {
	// получение указателя на структуру для работы с БД
	deps := []interface{}{
		app.NewServer,
		http.NewServeMux,
		func() (*pgxpool.Pool, error) {
			ctx, _ := context.WithTimeout(context.Background(), time.Second*5)
			return pgxpool.Connect(ctx, dsn)
		},
		customers.NewService,
		func(server *app.Server) *http.Server {
			return &http.Server{
				Addr:    net.JoinHostPort(host, port),
				Handler: server,
			}
		},
	}

	container := dig.New()
	for _, dep := range deps {
		err = container.Provide(dep)
		if err != nil {
			return err
		}
	}

	err = container.Invoke(func(server *app.Server) {
		server.Init()
	})
	if err != nil {
		return err
	}

	return container.Invoke(func(server *http.Server) error {
		return server.ListenAndServe()
	})
}
