package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-sql-driver/mysql"
)

type ourHandler struct{}

type User struct {
	ID    int64
	Name  string
	Email string
}

func (h ourHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" && req.URL.Path == "/add" {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}

		var user User

		err = json.Unmarshal(body, &user)
		if err != nil {
			panic(err)
		}

		_, err = addUser(user)
		if err != nil {
			panic(err)
		}

		return
	}

	users, err := getUsersFromDB()
	if err != nil {
		panic(err)
	}

	json, err := json.Marshal(users)
	if err != nil {
		panic(err)
	}

	res.Write(json)
}

var db *sql.DB

func addUser(user User) (User, error) {
	stmt, err := db.Prepare("INSERT INTO users SET name=?, email=?")
	if err != nil {
		return User{}, err
	}

	result, err := stmt.Exec(user.Name, user.Email)
	if err != nil {
		return User{}, err
	}

	user.ID, err = result.LastInsertId()
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func getUsersFromDB() ([]User, error) {
	var usersFromDB []User

	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			return nil, err
		}

		usersFromDB = append(usersFromDB, user)
	}

	return usersFromDB, nil
}

func main() {
	cfg := mysql.Config{
		User:   "root",
		Passwd: "local",
		Net:    "tcp",
		Addr:   "127.0.0.1:3306",
		DBName: "Contact-list",
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected!")

	err = http.ListenAndServe(":7070", ourHandler{})
	if err != nil {
		panic(err)
	}
}
