package internal

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

var ErrNotFound = errors.New("repository: resource not found")

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) Repository {
	return &UserRepository{
		db: db,
	}
}

type user struct {
	ID        string `db:"_id"`
	Firstname string `db:"firstname"`
	Lastname  string `db:"lastname"`
	Email     string `db:"email"`
	Password  string `db:"password"`
}

func (r *UserRepository) FindUserByEmail(email string) error {
	const findUserByEmail = `SELECT COUNT(1) FROM login WHERE email = :email`

	stmt, err := r.db.PrepareNamed(findUserByEmail)
	if err != nil {
		return err
	}

	defer stmt.Close()

	queryParams := map[string]interface{}{"email": email}

	var count int
	err = stmt.Get(&count, queryParams)

	if err != nil {
		return err
	}

	if count == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *UserRepository) GetUserByEmail(email string) (User, error) {
	const getUserByEmail = `SELECT user._id, user.firstname, user.lastname, login.email, login.password
								FROM login
								INNER JOIN user
								ON user.id = login.user_id
								WHERE email = :email`

	stmt, err := r.db.PrepareNamed(getUserByEmail)
	if err != nil {
		return User{}, err
	}

	defer stmt.Close()

	queryParams := map[string]interface{}{"email": email}

	var u user
	err = stmt.Get(&u, queryParams)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return User{}, err
	}

	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNotFound
	}

	return User{ // nolint
		ID:        u.ID,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
	}, nil
}

func (r *UserRepository) SaveUser(newUser NewUser) error {
	const (
		insertUserIntoUserTable  = `INSERT INTO user (_id, firstname, lastname) VALUES (:_id, :firstname, :lastname)`
		insertUserIntoLoginTable = `INSERT INTO login (email, password, user_id) VALUES (:email, :password, :user_id)`
	)

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("beggining tx: %v", err)
	}

	result, err := tx.NamedExec(insertUserIntoUserTable, map[string]interface{}{
		"_id":       newUser.ID,
		"firstname": newUser.Firstname,
		"lastname":  newUser.Lastname,
	})

	if err != nil {
		return err
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("getting last insert id: %v", err)
	}

	_, err = tx.NamedExec(insertUserIntoLoginTable, map[string]interface{}{
		"email":    newUser.Email,
		"password": newUser.Password,
		"user_id":  lastID,
	})

	if err != nil {
		return err
	}

	return tx.Commit()
}
