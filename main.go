package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/go-sql-driver/mysql"
)

type Contact struct {
	ID      int64     `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Phone   uint64    `json:"phone"`
	Created time.Time `json:"created"`
}

type ContactFragment struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
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
func getContactFragmentFromDB() ([]ContactFragment, error) {
	var contactFragmentFromDB []ContactFragment

	rows, err := db.Query("SELECT id,name FROM contacts")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var contact ContactFragment
		err = rows.Scan(&contact.ID, &contact.Name)
		if err != nil {
			return nil, err
		}

		contactFragmentFromDB = append(contactFragmentFromDB, contact)
	}

	return contactFragmentFromDB, nil

}

func getContactsFromDB() ([]Contact, error) {
	contactsFromDB := []Contact{}

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
func httpGetContactsFragment(ctx *gin.Context) {
	contacts, err := getContactFragmentFromDB()
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
func updateContact(contact Contact, contactID int64) (Contact, error) {
	stmt, err := db.Prepare("UPDATE contacts SET name=?, email=?, phone=?, created=? WHERE id=?")
	if err != nil {
		return Contact{}, err
	}
	defer stmt.Close()

	location, err := time.LoadLocation("Europe/Prague")
	if err != nil {
		return Contact{}, err
	}

	created := time.Now().In(location)

	_, err = stmt.Exec(contact.Name, contact.Email, contact.Phone, created, contactID)
	if err != nil {
		return Contact{}, err
	}

	contact.ID = contactID

	return contact, nil
}

func httpUpdateContact(ctx *gin.Context) {
	id := ctx.Param("id")
	var contact Contact
	err := ctx.BindJSON(&contact)
	if err != nil {
		panic(err)
	}

	contactID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	updatedContact, err := updateContact(contact, contactID)
	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, updatedContact)
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

	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "DELETE", "PUT"},
		AllowHeaders:    []string{"Content-Type"},
	}))
	router.GET("/contacts", httpGetContacts)
	router.GET("/contact/:id", httpFindContactbyID)
	router.GET("/contact/id/name", httpGetContactsFragment)
	router.POST("/add", httpAddContact)
	router.DELETE("/delete/:id", httpDeleteContactbyID)
	router.PUT("/update/:id", httpUpdateContact)
	err = router.Run(":7070")
	if err != nil {
		panic(err)
	}
}
