package main

import (
	"database/sql"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id             int
	Email          string
	HashedPassword string
	CreatedAt      time.Time
}

func authenticate(email string, password string) (bool, int) {
	var hpw string
	var id int

	err := db.QueryRow(`
	SELECT
		 id
		,hashed_password
	FROM
		users
	WHERE
		email = $1`,
		email).Scan(&id, &hpw)
	if err == sql.ErrNoRows {
		log.Println("attempt by unregisted email address.")
		return false, 0
	} else if err != nil {
		log.Println(err)
		return false, 0
	}
	if checkPasswordHash(password, hpw) {
		return true, id
	} else {
		return false, 0
	}
}

func registUser(email string, password string) (int, error) {
	hpw, err := hashPassword(password)
	if err != nil {
		return 0, err
	}
	statement := `
	INSERT INTO users(
		email,
		hashed_password,
		created_at,
		updated_at)
	VALUES(
		$1,
		$2,
		$3,
		$4
	)
	RETURNING id`
	stmt, err := db.Prepare(statement)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	var new_id int
	err = stmt.QueryRow(email, hpw, time.Now(), time.Now()).Scan(&new_id)
	if err != nil {
		return 0, err
	}

	return new_id, nil
}

func isExistUser(email string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT true FROM users WHERE email = $1", email).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
