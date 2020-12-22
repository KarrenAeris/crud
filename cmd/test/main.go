package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/KarrenAeris/crud/pkg/customers"

	"os"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	// адрес подключения
	// протокол://логи:пароль@хост:порт/бд

	dsn := "postgres://app:pass@localhost:5432/db"
	// получение указателя на структуру для работы с БД
	connectCtx, _ := context.WithTimeout(context.Background(), time.Second*5)
	pool, err := pgxpool.Connect(connectCtx, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
		return
	}
	// закрытие структуры
	defer pool.Close()

	log.Print(pool.Stat().TotalConns()) // 1 подключение
	// контекст для запросов
	ctx := context.Background()
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Release()
	log.Print(pool.Stat().TotalConns()) // 1 подключение
	conn, err = pool.Acquire(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	conn.Release()
	log.Print(pool.Stat().TotalConns()) // 2, несмотря на то, что закрыли

	// новое подключение не создастся, будет переиспользовано предыдущее
	log.Print(pool.Stat().TotalConns()) // 1 подключение
	conn, err = pool.Acquire(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	conn.Release()
	log.Print(pool.Stat().TotalConns()) // 2, несмотря на то, что закрыли

	// TODO: заросы
	sql := `SELECT id, name, phone, active, created FROM customers WHERE id = $1;`

	item := &customers.Customer{}
	// var id string
	err = pool.QueryRow(ctx, sql, 1).Scan(&item.ID, &item.Name, &item.Phone, &item.Active, &item.Created)
	if errors.Is(err, pgx.ErrNoRows) {
		log.Println("No Rows")
		return
	}
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(item.Name)

	// создаём слайс для хранения результатов
	items := []*Managers{}

	// делаем сам запрос
	sql = `SELECT id, name, salary FROM managers WHERE active;`
	rows, err := pool.Query(ctx, sql)
	if err != nil {
		log.Println(err)
		return
	}
	// rows нужно закрыть
	defer rows.Close()
	// rows.Next() возвращает true до тех пор, пока далбше есть строка
	for rows.Next() {
		item := &Managers{}
		err = rows.Scan(&item.ID, &item.Name, &item.Salary)
		if err != nil {
			log.Println(err)
			return
		}
		items = append(items, item)
	}
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(items[0])

	sql = `UPDATE managers SET active = false WHERE id = 1`
	_, err = pool.Exec(ctx, sql)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println("success")
}

// Managers ...
type Managers struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Salary int64  `json:"salary"`
}
