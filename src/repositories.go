package main

import (
	"database/sql"
	"fmt"
	"log"

	// TODO: See about replacing with: https://github.com/jackc/pgx
	_ "github.com/lib/pq"
)

// db instance
var db *sql.DB

// connect to db
func connectDb(s DataSource) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		s.Host, s.Port, s.User, s.Password, s.Dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Connected to %s as %s\n", s.Dbname, s.User)
	return db
}

func getUserByName(username string) (User, error) {
	sql := "SELECT * FROM users WHERE username = $1;"
	var user User
	err := db.QueryRow(sql, username).Scan(&user.ID, &user.Username, &user.Password, &user.Created)
	if err != nil {
		return user, err
	}
	return user, nil
}

func createUser(user User) (User, error) {
	sql := "INSERT INTO users (username, password, created) VALUES ($1, $2, $3) RETURNING id;"

	err := db.QueryRow(sql, user.Username, user.Password, user.Created).Scan(&user.ID)
	if err != nil {
		return user, err
	}
	return user, nil
}
