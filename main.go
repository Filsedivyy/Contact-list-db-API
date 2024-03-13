package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/go-sql-driver/mysql"
)

type Contact struct {
	ID      int64
	Name    string
	Email   string
	Phone   uint64
	Created time.Time
}

var db *sql.DB

func addContact(contact Contact) (Contact, error) {
	stmt, err := db.Prepare("INSERT INTO contacts SET name=?, email=?, phone=?, created=?")
	if err != nil {
		return Contact{}, err
	}

	location, err := time.LoadLocation("Europe/Prague")
	if err != nil {
		panic(err)
	}

	created := time.Now().In(location)

	result, err := stmt.Exec(contact.Name, contact.Email, contact.Phone, created)
	if err != nil {
		return Contact{}, err
	}

	contact.ID, err = result.LastInsertId()
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

	defer rows.Close()

	for rows.Next() {
		var contact Contact
		err = rows.Scan(&contact.ID, &contact.Name, &contact.Email, &contact.Phone, &contact.Created)
		if err != nil {
			return nil, err
		}

		contactsFromDB = append(contactsFromDB, contact)
	}

	return contactsFromDB, nil
}
func httpAddContact(ctx *gin.Context) {
	var contact Contact
	err := ctx.BindJSON(&contact)
	if err != nil {
		panic(err)
	}
	_, err = addContact(contact)
	if err != nil {
		panic(err)
	}
	ctx.Status(http.StatusCreated)

}
func httpGetContacts(ctx *gin.Context) {
	contacts, err := getContactsFromDB()
	if err != nil {
		panic(err)
	}
	ctx.JSON(http.StatusOK, contacts)
}
func httpFindContactbyID(ctx *gin.Context) {
	id := ctx.Param("id")
	fmt.Println(id)
	ctx.Status(http.StatusOK)

	row := db.QueryRow("SELECT * FROM contacts WHERE id=?", id)
	if row == nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	var contact Contact
	err := row.Scan(&contact.ID, &contact.Name, &contact.Email, &contact.Phone, &contact.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.Status(http.StatusNotFound)
			return
		}
		fmt.Println(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.JSON(http.StatusOK, contact)
}

func httpDeleteContactbyID(ctx *gin.Context) {
	id := ctx.Param("id")
	fmt.Println(id)
	ctx.Status(http.StatusOK)
	row := db.QueryRow("DELETE  FROM contacts WHERE id=?", id)
	if row == nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.Status(http.StatusAccepted)
}
func main() {
	cfg := mysql.Config{
		User:      "root",
		Passwd:    "local",
		Net:       "tcp",
		Addr:      "127.0.0.1:3306",
		DBName:    "Contact-list",
		ParseTime: true,
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

	router := gin.Default()

	router.GET("/contacts", httpGetContacts)
	router.GET("/contact/:id", httpFindContactbyID)
	router.POST("/add", httpAddContact)
	router.DELETE("/delete/:id", httpDeleteContactbyID)
	err = router.Run(":7070")
	if err != nil {
		panic(err)
	}
}
