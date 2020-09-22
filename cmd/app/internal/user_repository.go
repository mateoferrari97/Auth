package internal

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mateoferrari97/auth/internal"
)

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

const findUserByEmail = `SELECT COUNT(1) FROM login WHERE email = :email`

func (r *UserRepository) FindUserByEmail(email string) error {
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
		return fmt.Errorf("%w: db not found", internal.ErrResourceNotFound)
	}

	return nil
}

const getUserByEmail = `SELECT user._id, user.firstname, user.lastname, login.email, login.password
								FROM login
								INNER JOIN user
								ON user.id = login.user_id
								WHERE email = :email`

func (r *UserRepository) GetUserByEmail(email string) (User, error) {
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
		return User{}, fmt.Errorf("%w: db not found", internal.ErrResourceNotFound)
	}

	return User{ // nolint
		ID:        u.ID,
		Firstname: u.Firstname,
		Lastname:  u.Lastname,
		Email:     u.Email,
	}, nil
}

const (
	insertUserIntoUserTable  = `INSERT INTO user (_id, firstname, lastname) VALUES (:_id, :firstname, :lastname)`
	insertUserIntoLoginTable = `INSERT INTO login (email, password, user_id) VALUES (:email, :password, :user_id)`
)

func (r *UserRepository) SaveUser(newUser NewUser) (err error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("beggining tx: %v", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback() // nolint
		}
	}()

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
