package internal

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestFindUserByEmail(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))
	email := "mateo.ferrari97@gmail.com"

	mock.ExpectPrepare(`SELECT COUNT(1) FROM login WHERE email = ?`).
		WillReturnError(nil)

	mock.ExpectQuery(`SELECT COUNT(1) FROM login WHERE email = ?`).
		WithArgs(email).
		WillReturnError(nil).
		WillReturnRows(
			sqlmock.NewRows([]string{"COUNT(1)"}).
				AddRow(1),
		)

	// When
	err = r.FindUserByEmail(email)

	// Then
	require.NoError(t, err)
}

func TestFindUserByEmail_PreparingCountUsersFromEmailError(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))
	email := "mateo.ferrari97@gmail.com"

	mock.ExpectPrepare(`SELECT COUNT(1) FROM login WHERE email = ?`).
		WillReturnError(errors.New("preparing query error"))

	// When
	err = r.FindUserByEmail(email)

	// Then
	require.EqualError(t, err, "preparing query error")
}

func TestFindUserByEmail_ExecutingCountUsersFromEmailError(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))
	email := "mateo.ferrari97@gmail.com"

	mock.ExpectPrepare(`SELECT COUNT(1) FROM login WHERE email = ?`).
		WillReturnError(nil)

	mock.ExpectQuery(`SELECT COUNT(1) FROM login WHERE email = ?`).
		WithArgs(email).
		WillReturnError(errors.New("executing query error"))

	// When
	err = r.FindUserByEmail(email)

	// Then
	require.EqualError(t, err, "executing query error")
}

func TestFindUserByEmail_NotFound(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))
	email := "mateo.ferrari97@gmail.com"

	mock.ExpectPrepare(`SELECT COUNT(1) FROM login WHERE email = ?`).
		WillReturnError(nil)

	mock.ExpectQuery(`SELECT COUNT(1) FROM login WHERE email = ?`).
		WithArgs(email).
		WillReturnError(nil).
		WillReturnRows(
			sqlmock.NewRows([]string{"COUNT(1)"}).
				AddRow(0),
		)

	// When
	err = r.FindUserByEmail(email)

	// Then
	require.EqualError(t, err, "repository: resource not found")
}

func TestGetUserByEmail(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))

	email := "mateo.ferrari97@gmail.com"
	q := `SELECT user._id, user.firstname, user.lastname, login.email, login.password
			FROM login
			INNER JOIN user
			ON user.id = login.user_id
			WHERE email = ?`

	mock.ExpectPrepare(q).WillReturnError(nil)
	mock.ExpectQuery(q).
		WithArgs(email).
		WillReturnError(nil).
		WillReturnRows(
			sqlmock.NewRows([]string{"_id", "firstname", "lastname", "email", "password"}).
				AddRow(
					"88096ae1-129e-4ef8-8bdc-a8ace0753687",
					"mateo",
					"ferrari coronel",
					"mateo.ferrari97@gmail.com",
					"$2a$10$uAnfASxQBqdUlTlX8MV43utR.Cun0gr9MKdVpbG8Cy44jD1N2J4f."),
		)

	// When
	resp, err := r.GetUserByEmail(email)
	if err != nil {
		t.Fatal(err)
	}

	// Then
	require.Equal(t, "88096ae1-129e-4ef8-8bdc-a8ace0753687", resp.ID)
	require.Equal(t, "mateo", resp.Firstname)
	require.Equal(t, "ferrari coronel", resp.Lastname)
	require.Equal(t, "mateo.ferrari97@gmail.com", resp.Email)
}

func TestGetUserByEmail_PreparingSelectUsersQueryError(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))

	email := "mateo.ferrari97@gmail.com"
	q := `SELECT user._id, user.firstname, user.lastname, login.email, login.password
			FROM login
			INNER JOIN user
			ON user.id = login.user_id
			WHERE email = ?`

	mock.ExpectPrepare(q).WillReturnError(errors.New("preparing query error"))

	// When
	_, err = r.GetUserByEmail(email)

	// Then
	require.EqualError(t, err, "preparing query error")
}

func TestGetUserByEmail_ExecutingSelectUsersQueryError(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))

	email := "mateo.ferrari97@gmail.com"
	q := `SELECT user._id, user.firstname, user.lastname, login.email, login.password
			FROM login
			INNER JOIN user
			ON user.id = login.user_id
			WHERE email = ?`

	mock.ExpectPrepare(q).WillReturnError(nil)
	mock.ExpectQuery(q).
		WithArgs(email).
		WillReturnError(errors.New("executing query error"))

	// When
	_, err = r.GetUserByEmail(email)

	// Then
	require.EqualError(t, err, "executing query error")
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))

	email := "mateo.ferrari97@gmail.com"
	q := `SELECT user._id, user.firstname, user.lastname, login.email, login.password
			FROM login
			INNER JOIN user
			ON user.id = login.user_id
			WHERE email = ?`

	mock.ExpectPrepare(q).WillReturnError(nil)
	mock.ExpectQuery(q).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)

	// When
	_, err = r.GetUserByEmail(email)

	// Then
	require.EqualError(t, err, "repository: resource not found")
}

func TestSaveUser(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))
	user := NewUser{
		ID:        "id",
		Firstname: "mateo",
		Lastname:  "ferrari coronel",
		Email:     "mateo.ferrari97@gmail.com",
		Password:  "123",
	}

	firstQuery := `INSERT INTO user (_id, firstname, lastname) VALUES (?, ?, ?)`
	secondQuery := `INSERT INTO login (email, password, user_id) VALUES (?, ?, ?)`

	mock.ExpectBegin()

	mock.ExpectExec(firstQuery).
		WithArgs(user.ID, user.Firstname, user.Lastname).
		WillReturnError(nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(secondQuery).
		WithArgs(user.Email, user.Password, 1).
		WillReturnError(nil).
		WillReturnResult(sqlmock.NewResult(2, 1))

	mock.ExpectCommit().WillReturnError(nil)

	// When
	err = r.SaveUser(user)

	// Then
	require.NoError(t, err)
}

func TestSaveUser_BeginTxError(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))
	user := NewUser{
		ID:        "id",
		Firstname: "mateo",
		Lastname:  "ferrari coronel",
		Email:     "mateo.ferrari97@gmail.com",
		Password:  "123",
	}

	mock.ExpectBegin().WillReturnError(errors.New("begging tx error"))

	// When
	err = r.SaveUser(user)

	// Then
	require.EqualError(t, err, "beggining tx: begging tx error")
}

func TestSaveUser_InsertingUserIntoUserTableError(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))
	user := NewUser{
		ID:        "id",
		Firstname: "mateo",
		Lastname:  "ferrari coronel",
		Email:     "mateo.ferrari97@gmail.com",
		Password:  "123",
	}

	firstQuery := `INSERT INTO user (_id, firstname, lastname) VALUES (?, ?, ?)`

	mock.ExpectBegin()

	mock.ExpectExec(firstQuery).
		WithArgs(user.ID, user.Firstname, user.Lastname).
		WillReturnError(errors.New("preparing query error"))

	mock.ExpectRollback()

	// When
	err = r.SaveUser(user)

	// Then
	require.EqualError(t, err, "preparing query error")
}

func TestSaveUser_GettingLastIDError(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))
	user := NewUser{
		ID:        "id",
		Firstname: "mateo",
		Lastname:  "ferrari coronel",
		Email:     "mateo.ferrari97@gmail.com",
		Password:  "123",
	}

	firstQuery := `INSERT INTO user (_id, firstname, lastname) VALUES (?, ?, ?)`

	mock.ExpectBegin()

	mock.ExpectExec(firstQuery).
		WithArgs(user.ID, user.Firstname, user.Lastname).
		WillReturnError(nil).
		WillReturnResult(sqlmock.NewErrorResult(errors.New("db error")))

	mock.ExpectCommit().WillReturnError(nil)

	mock.ExpectRollback()

	// When
	err = r.SaveUser(user)

	// Then
	require.EqualError(t, err, "getting last insert id: db error")
}

func TestSaveUser_InsertingUserIntoLoginTableError(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))
	user := NewUser{
		ID:        "id",
		Firstname: "mateo",
		Lastname:  "ferrari coronel",
		Email:     "mateo.ferrari97@gmail.com",
		Password:  "123",
	}

	firstQuery := `INSERT INTO user (_id, firstname, lastname) VALUES (?, ?, ?)`
	secondQuery := `INSERT INTO login (email, password, user_id) VALUES (?, ?, ?)`

	mock.ExpectBegin()

	mock.ExpectExec(firstQuery).
		WithArgs(user.ID, user.Firstname, user.Lastname).
		WillReturnError(nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(secondQuery).
		WithArgs(user.Email, user.Password, 1).
		WillReturnError(errors.New("db error"))

	mock.ExpectCommit().WillReturnError(nil)

	mock.ExpectRollback()

	// When
	err = r.SaveUser(user)

	// Then
	require.EqualError(t, err, "db error")
}

func TestSaveUser_CommittingError(t *testing.T) {
	// Given
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		t.Fatalf("starting sql mock: %v", err)
	}

	defer db.Close()

	r := NewUserRepository(sqlx.NewDb(db, "mysql"))
	user := NewUser{
		ID:        "id",
		Firstname: "mateo",
		Lastname:  "ferrari coronel",
		Email:     "mateo.ferrari97@gmail.com",
		Password:  "123",
	}

	firstQuery := `INSERT INTO user (_id, firstname, lastname) VALUES (?, ?, ?)`
	secondQuery := `INSERT INTO login (email, password, user_id) VALUES (?, ?, ?)`

	mock.ExpectBegin()

	mock.ExpectExec(firstQuery).
		WithArgs(user.ID, user.Firstname, user.Lastname).
		WillReturnError(nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec(secondQuery).
		WithArgs(user.Email, user.Password, 1).
		WillReturnError(nil).
		WillReturnResult(sqlmock.NewResult(2, 1))

	mock.ExpectCommit().WillReturnError(errors.New("db error"))

	mock.ExpectRollback()

	// When
	err = r.SaveUser(user)

	// Then
	require.EqualError(t, err, "db error")
}
