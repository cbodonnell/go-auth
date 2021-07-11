package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	// TODO: See about replacing with: https://github.com/jackc/pgx/v4
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

func getUserByID(userID int) (User, error) {
	sql := "SELECT * FROM users WHERE id = $1;"
	var user User
	err := db.QueryRow(sql, userID).Scan(&user.ID, &user.Username, &user.Password, &user.Created, &user.UUID)
	if err != nil {
		return user, err
	}
	return user, nil
}

func getUserByName(username string) (User, error) {
	sql := "SELECT * FROM users WHERE username = $1;"
	var user User
	err := db.QueryRow(sql, username).Scan(&user.ID, &user.Username, &user.Password, &user.Created, &user.UUID)
	if err != nil {
		return user, err
	}
	return user, nil
}

func createUser(user User) (User, error) {
	tx, err := db.Begin()
	if err != nil {
		return user, err
	}
	sql := `INSERT INTO users (username, password, created, uuid) VALUES ($1, $2, $3, $4) RETURNING id;`
	err = tx.QueryRow(sql, user.Username, user.Password, user.Created, user.UUID).Scan(&user.ID)
	if err != nil {
		tx.Rollback()
		return user, err
	}
	sql = "INSERT INTO user_groups (user_id, group_id) VALUES ($1, 1);"
	_, err = tx.Exec(sql, user.ID)
	if err != nil {
		tx.Rollback()
		return user, err
	}
	tx.Commit()
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

func saveRefresh(userID int, refreshString string) error {
	sql := `INSERT INTO user_refresh (user_id, refresh) VALUES ($1, $2);`

	_, err := db.Exec(sql, userID, refreshString)
	if err != nil {
		return err
	}
	return nil
}

func validateRefresh(userID int, refreshString string) error {
	sql := `SELECT 1 FROM user_refresh
	WHERE user_id = $1
	AND refresh = $2;`

	var result int
	err := db.QueryRow(sql, userID, refreshString).Scan(&result)
	if err != nil {
		return err
	}
	if result != 1 {
		return errors.New("invalid refresh token")
	}
	return nil
}

func deleteRefresh(refreshString string) error {
	sql := `DELETE FROM user_refresh
	WHERE refresh = $1;`

	_, err := db.Exec(sql, refreshString)
	if err != nil {
		return err
	}
	return nil
}
