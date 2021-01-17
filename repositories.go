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

// ping db
func pingDb(db *sql.DB) {
	err := db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func getUserByName(username string) (User, error) {
	sql := "SELECT * FROM users WHERE username = $1;"
	var user User
	err := db.QueryRow(sql, username).Scan(&user.ID, &user.Username, &user.Password, &user.Created, &user.StreamKey)
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

	sql = "INSERT INTO user_groups (user_id, group_id) VALUES ($1, 1);"
	_, err = db.Exec(sql, user.ID)
	if err != nil {
		return user, err
	}

	return user, nil
}

func getUserGroups(userID int) ([]Group, error) {
	sql := `SELECT groups.id, groups.name
	FROM groups 
	INNER JOIN user_groups 
	ON user_groups.group_id = groups.id
	INNER JOIN users
	ON users.id = user_groups.user_id
	WHERE users.id = $1;`
	rows, err := db.Query(sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []Group
	for rows.Next() {
		var group Group
		err = rows.Scan(&group.ID, &group.Name)
		if err != nil {
			return groups, err
		}
		groups = append(groups, group)
	}
	err = rows.Err()
	if err != nil {
		return groups, err
	}
	return groups, nil
}

func updatePassword(userID int, password string) error {
	sql := "UPDATE users SET password = $1 WHERE id = $2;"

	_, err := db.Exec(sql, password, userID)
	if err != nil {
		return err
	}
	return nil
}
