package data

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"runtime"
	"strings"
	"time"
)

// User struct holds information about User
// for logon/signup/authenicate purposes.
type User struct {
	// A user has following details.
	// this can be a database unique id.
	ID int
	// session id.
	UUID string
	// Name of the user.
	Name string
	// email of the user.
	Email string
	// password - encrypted string
	Password string
	// time the user is created.
	CreatedAt time.Time
}

// Session struct maintains details about
// current session of user.
type Session struct {
	// Id - this can be database id.
	ID int
	// UUID - this can be session id.
	UUID string
	// email of user
	Email string
	// UserID
	UserID int
	// createdTime of session
	CreatedAt time.Time
}

var key = []byte(func() (key string) {
	var wd string

	_, filename, _, ok := runtime.Caller(0)

	if ok {
		wd = strings.Split(filename, "user.go")[0]
	}

	bytes, err := ioutil.ReadFile(wd + "/yek.txt")

	panicErrs(err)

	key = string(bytes)

	return
}())

// CreateUser function creates user and stores user data into database.
func (u *User) CreateUser() (err error) {
	sqlStmt := "insert into users (uuid, name, email, password, created_at) " +
		" values($1,$2,$3,$4,$5) returning id,uuid,created_at"

	dbStmt, err := DB.Prepare(sqlStmt)

	if err != nil {
		log.Fatal(err)
	}

	defer dbStmt.Close()

	encPass, err := Encrypt([]byte(u.Password), key)

	panicErrs(err)

	err = dbStmt.QueryRow(createUUID(), u.Name, u.Email, encPass, time.Now()).Scan(&u.ID, &u.UUID, &u.CreatedAt)

	return
}

// GetUserByEmail function retrieves user details from
// db using email id provided as parameter.
func (u *User) GetUserByEmail(email string) (indicator int) {
	err := DB.QueryRow("select * from users where email=$1", email).
		Scan(&u.ID, &u.UUID, &u.Name, &u.Email, &u.Password, &u.CreatedAt)

	if err == sql.ErrNoRows {
		indicator = 1
		return
	}

	panicErrs(err)

	stringPass, err := Decrypt(u.Password, key)

	u.Password = string(stringPass)

	return
}

// CreateSession function inserts session id into db and returns session struct.
func (u *User) CreateSession() (session Session, err error) {

	stmt := "insert into sessions (uuid, email, user_id, created_at)" +
		" values($1,$2,$3,$4) returning id, uuid, email, user_id, created_at"

	dbStmt, err := DB.Prepare(stmt)

	defer dbStmt.Close()

	panicErrs(err)

	dbStmt.QueryRow(createUUID(), u.Email, u.ID, time.Now()).Scan(&session.ID, &session.UUID, &session.Email, &session.UserID, &session.CreatedAt)

	fmt.Println(session)
	return
}

// DeleteAllSessions function delete all sessions for user in database.
func (u *User) DeleteAllSessions() {
	// check if historic session for user id is present in database.
	// if present, delete the row.
	deleteSessionStmt := " delete from sessions where user_id=$1 "

	delDbStmt, err := DB.Prepare(deleteSessionStmt)

	defer delDbStmt.Close()

	delDbStmt.Exec(u.ID)

	panicErrs(err)
}

// DeleteSession function deletes active session,
// the session currently active in browser from database.
func (u *User) DeleteSession(uuid string) {
	stmt := " delete from sessions where uuid=$1 "

	_, err := DB.Exec(stmt, uuid)

	if err != nil {
		panicErrs(err)
	}
}

// CheckSession function checks if uuid parameter
// passed to method is a valid one or not and
// returns a boolean variable regarding validity.
func (session *Session) CheckSession(uuid string) (validSession bool, err error) {
	stmt := " select id, uuid, email, user_id, created_at from sessions " +
		" where uuid=$1 "

	err = DB.QueryRow(stmt, uuid).Scan(&session.ID, &session.UUID,
		&session.Email, &session.UserID, &session.CreatedAt)

	fmt.Println("err in checksession", err == sql.ErrNoRows)

	if err == sql.ErrNoRows {
		validSession = false
		err = nil
		return
	}

	if err != nil {
		return
	}

	validSession = true

	return
}

// GetUserByID functiton retrieves user id row by provided id
func (u *User) GetUserByID() {
	stmt := " select id, uuid, name, email, password, created_at " +
		" from users where id=$1 "

	err := DB.QueryRow(stmt, u.ID).Scan(&u.ID, &u.UUID, &u.Name, &u.Email, &u.Password, &u.CreatedAt)

	if err != nil {
		panicErrs(err)
	}
}
