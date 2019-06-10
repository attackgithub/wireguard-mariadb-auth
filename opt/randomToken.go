package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

func generateToken() string {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var randBytes [61]byte
	n, err := rand.Read(randBytes[:])
	if err != nil {
		panic(err)
	}
	if n != len(randBytes) {
		panic(errors.New("generateToken() failed: short random read"))
	}

	for i, b := range randBytes {
		randBytes[i] = letters[b%byte(len(letters))]
	}

	return string(randBytes[:])

}

func getRandomCredentials() (string, string) {

	randomDevice, randomToken := generateToken(), generateToken()

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
	db, err := sql.Open("mysql", dbString)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// test DB connection
	// this will be removed soon
	dbQuery := "INSERT INTO devices (device, token, d) VALUES ('" +
		randomDevice + "', '" +
		randomToken + "', NOW())"
	_, err = db.Query(dbQuery)
	if err != nil {
		panic(err.Error())
	}

	return randomDevice, randomToken
}
