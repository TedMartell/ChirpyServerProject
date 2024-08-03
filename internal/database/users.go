package database

import (
	"errors"
)

type User struct {
	ID             int          `json:"id"`
	Email          string       `json:"email"`
	HashedPassword string       `json:"hashed_password"`
	RefreshToken   RefreshToken `json:"refresh_token"`
	IsChirpyRed    bool         `json:"is_chirpy_red"`
}

var ErrAlreadyExists = errors.New("already exists")

func (db *DB) CreateUser(email, hashedPassword string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}
	newID := len(dbStructure.Users) + 1
	user := User{
		ID:             newID,
		Email:          email,
		HashedPassword: hashedPassword,
		IsChirpyRed:    false, // defaulting to false
	}
	dbStructure.Users[newID] = user
	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func (db *DB) GetUser(id int) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, ErrNotExist
	}

	return user, nil
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, ErrNotExist
}

func (db *DB) UpdateUser(id int, email, hashedPassword string, isChirpyRed bool) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, ok := dbStructure.Users[id]
	if !ok {
		return User{}, ErrNotExist
	}

	user.Email = email
	user.HashedPassword = hashedPassword
	user.IsChirpyRed = isChirpyRed
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
