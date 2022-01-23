package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/cheebz/go-auth/config"
	"github.com/cheebz/go-auth/models"
	"github.com/jackc/pgx/v4/pgxpool"
)

type PSQLRepository struct {
	Conf config.Configuration
	Db   *pgxpool.Pool
}

func NewPSQLRepository(conf config.Configuration) Repository {
	return &PSQLRepository{
		Conf: conf,
		Db:   connectDb(conf.Db),
	}
}

// connect to db
func connectDb(s config.DataSource) *pgxpool.Pool {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		s.Host, s.Port, s.User, s.Password, s.Dbname)
	db, err := pgxpool.Connect(context.Background(), psqlInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	log.Printf("Connected to %s as %s\n", s.Dbname, s.User)
	return db
}

func (r *PSQLRepository) Close() {
	r.Db.Close()
}

func (r *PSQLRepository) GetUserByID(userID int) (models.User, error) {
	sql := "SELECT * FROM users WHERE id = $1;"
	var user models.User
	err := r.Db.QueryRow(context.Background(), sql, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Created,
		&user.UUID,
	)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (r *PSQLRepository) GetUserByName(username string) (models.User, error) {
	sql := "SELECT * FROM users WHERE username = $1;"
	var user models.User
	err := r.Db.QueryRow(context.Background(), sql, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Created,
		&user.UUID,
	)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (r *PSQLRepository) CreateUser(user models.User) (models.User, error) {
	tx, err := r.Db.Begin(context.Background())
	if err != nil {
		return user, err
	}
	sql := `INSERT INTO users (username, password, created, uuid) VALUES ($1, $2, $3, $4) RETURNING id;`
	err = tx.QueryRow(context.Background(), sql, user.Username, user.Password, user.Created, user.UUID).Scan(&user.ID)
	if err != nil {
		tx.Rollback(context.Background())
		return user, err
	}
	sql = "INSERT INTO user_groups (user_id, group_id) VALUES ($1, 1);"
	_, err = tx.Exec(context.Background(), sql, user.ID)
	if err != nil {
		tx.Rollback(context.Background())
		return user, err
	}
	tx.Commit(context.Background())
	return user, nil
}

func (r *PSQLRepository) GetUserGroups(userID int) ([]models.Group, error) {
	sql := `SELECT groups.id, groups.name
	FROM groups 
	INNER JOIN user_groups 
	ON user_groups.group_id = groups.id
	INNER JOIN users
	ON users.id = user_groups.user_id
	WHERE users.id = $1;`
	rows, err := r.Db.Query(context.Background(), sql, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var groups []models.Group
	for rows.Next() {
		var group models.Group
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

func (r *PSQLRepository) UpdatePassword(userID int, password string) error {
	sql := "UPDATE users SET password = $1 WHERE id = $2;"

	_, err := r.Db.Exec(context.Background(), sql, password, userID)
	if err != nil {
		return err
	}
	return nil
}

func (r *PSQLRepository) SaveRefresh(userID int, jti string) error {
	sql := `INSERT INTO user_refresh (user_id, jti, expires)
	VALUES ($1, $2, current_timestamp + ($3 || ' seconds')::interval);`

	_, err := r.Db.Exec(context.Background(), sql, userID, jti, fmt.Sprintf("%d", r.Conf.RefreshMaxAge))
	if err != nil {
		return err
	}
	return nil
}

func (r *PSQLRepository) ValidateRefresh(userID int, jti string) error {
	sql := `SELECT 1 FROM user_refresh
	WHERE user_id = $1
	AND jti = $2;`

	var result int
	err := r.Db.QueryRow(context.Background(), sql, userID, jti).Scan(&result)
	if err != nil {
		return err
	}
	if result != 1 {
		return errors.New("invalid refresh token")
	}
	return nil
}

// Invalidate refresh to account for concurrent requests
func (r *PSQLRepository) InvalidateRefresh(jti string) error {
	sql := `UPDATE user_refresh
	SET expires = current_timestamp + (2 || ' minutes')::interval
	WHERE jti = $1;`

	_, err := r.Db.Exec(context.Background(), sql, jti)
	if err != nil {
		return err
	}
	return nil
}

func (r *PSQLRepository) DeleteAllRefresh(userID int) error {
	sql := `DELETE FROM user_refresh
	WHERE user_id = $1;`

	_, err := r.Db.Exec(context.Background(), sql, userID)
	if err != nil {
		return err
	}
	return nil
}

func (r *PSQLRepository) DeleteExpiredRefresh() error {
	sql := `DELETE FROM user_refresh
	WHERE expires < current_timestamp;`

	_, err := r.Db.Exec(context.Background(), sql)
	if err != nil {
		return err
	}
	return nil
}
