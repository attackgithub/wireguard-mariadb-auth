package main

import (
	"crypto/subtle"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"regexp"
)

var db = &sql.DB{}

func dbInit() {
	// create DB connection
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbString := dbUsername + ":" +
		dbPassword + "@tcp(" +
		dbHost + ":" +
		dbPort + ")/" +
		dbName
	var err error
	db, err = sql.Open("mysql", dbString)
	if err != nil {
		panic(err)
	}
}

func authenticate(device string, token string) bool {
	matchedDevice, _ := regexp.MatchString("^[a-zA-Z0-9]*$", device)
	matchedToken, _ := regexp.MatchString("^[a-zA-Z0-9]*$", token)
	if !matchedDevice || !matchedToken {
		return false
	}
	var col string
	sqlStatement := `SELECT token FROM devices WHERE device=?;`
	scanErr := db.QueryRow(sqlStatement, device).Scan(&col)
	if scanErr != nil {
		if scanErr != sql.ErrNoRows {
			panic(scanErr)
		}
		return false
	}
	if subtle.ConstantTimeCompare([]byte(col), []byte(token)) == 1 {
		return true
	}
	return false
}
