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

type Contact struct {
	ID    int64
	Name  string
	Email string
	Phone uint64
}

func (h ourHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if req.Method == "POST" && req.URL.Path == "/add" {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}

		var contact Contact

		err = json.Unmarshal(body, &contact)
		if err != nil {
			panic(err)
		}

		_, err = addUser(contact)
		if err != nil {
			panic(err)
		}

		return
	}

	contacts, err := getContactsFromDB()
	if err != nil {
		panic(err)
	}

	json, err := json.Marshal(contacts)
	if err != nil {
		panic(err)
	}

	res.Write(json)
}

var db *sql.DB

func addUser(contact Contact) (Contact, error) {
	stmt, err := db.Prepare("INSERT INTO contacts SET name=?, email=?,phone=?")
	if err != nil {
		return Contact{}, err
	}

	result, err := stmt.Exec(contact.Name, contact.Email, contact.Phone)
	if err != nil {
		return Contact{}, err
	}

	contact.ID, err = result.LastInsertId()
	if err != nil {
		return Contact{}, err
	}

	return contact, nil
}

func getContactById(contactID int64) (Contact, error) {

	query := "SELECT * FROM contacts WHERE id =?"

	row := db.QueryRow(query, contactID)

	var contact Contact

	err := row.Scan(&contact.ID, &contact.Name, &contact.Email, &contact.Phone)
	if err != nil {
		return Contact{}, err
	}

	return contact, nil
}

func getContactsFromDB() ([]Contact, error) {
	var contactsFromDB []Contact

	rows, err := db.Query("SELECT * FROM contacts")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var contact Contact
		err = rows.Scan(&contact.ID, &contact.Name, &contact.Email)
		if err != nil {
			return nil, err
		}

		contactsFromDB = append(contactsFromDB, contact)
	}

	return contactsFromDB, nil
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
