package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
)

type DB struct {
	path  string
	mutex *sync.RWMutex
	conn  *sql.DB
}

// Define the Chirp structure
type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	// Check if the file exists
	_, err := os.Stat(path)

	// If it does not exist, create it
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			// Handle potential error from creating the file
			return nil, err
		}
		// Close the file if created (good practice)
		file.Close()
	} else if err != nil {
		// Handle other error from os.Stat
		return nil, err
	}

	// Initialize the DB struct
	db := &DB{
		path:  path,
		mutex: &sync.RWMutex{},
	}

	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	data, err := os.ReadFile(db.path)
	if err != nil && !os.IsNotExist(err) {
		return Chirp{}, err
	}

	var dbStructure DBStructure

	if len(data) > 0 {
		if err := json.Unmarshal(data, &dbStructure); err != nil {
			return Chirp{}, err
		}
	} else {
		dbStructure.Chirps = make(map[int]Chirp)
	}

	newID := len(dbStructure.Chirps) + 1
	newChirp := Chirp{
		ID:   newID,
		Body: body,
	}

	dbStructure.Chirps[newID] = newChirp

	bytes, err := json.MarshalIndent(dbStructure, "", "  ")
	if err != nil {
		return Chirp{}, err
	}

	if err := os.WriteFile(db.path, bytes, 0666); err != nil {
		return Chirp{}, err
	}

	return newChirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	// Step 1: Lock the database for reading
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	// Step 2: Load the current state of the database
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	// Step 3: Extract the chirps from the database structure
	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	// Step 4: Sort the chirps by ID
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	// Step 5: Return the chirps
	return chirps, nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	// Check if the database file exists
	_, err := os.Stat(db.path)

	// If the file does not exist, create it
	if os.IsNotExist(err) {
		file, err := os.Create(db.path)
		if err != nil {
			// Handle error from file creation
			return err
		}
		// Close the file to release resources
		file.Close()
	} else if err != nil {
		// Handle other possible errors from os.Stat
		return err
	}

	return nil
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	// Step 1: Read the file contents
	data, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}

	// Step 2: Unmarshal the JSON data into a DBStructure
	var dbStructure DBStructure
	err = json.Unmarshal(data, &dbStructure)
	if err != nil {
		return DBStructure{}, err
	}

	return dbStructure, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	// Step 1: Marshal the DBStructure into JSON
	data, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	// Step 2: Write the JSON data to the file
	err = os.WriteFile(db.path, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// Implement the GetChirpByID function
func (db *DB) GetChirpByID(id int) (Chirp, error) {
	// Step 1: Lock the database for reading
	db.mutex.RLock()
	defer db.mutex.RUnlock()

	// Step 2: Load the current state of the database
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	// Step 3: Look for the chirp with the given ID
	chirp, exists := dbStructure.Chirps[id]
	if !exists {
		return Chirp{}, errors.New("chirp not found")
	}

	return chirp, nil
}

// CreateUser creates a new USER!!! and saves it to disk
func (db *DB) CreateUser(email string, password string) (User, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	data, err := os.ReadFile(db.path)
	if err != nil && !os.IsNotExist(err) {
		return User{}, err
	}

	var dbStructure DBStructure

	if len(data) > 0 {
		if err := json.Unmarshal(data, &dbStructure); err != nil {
			return User{}, err
		}
	}

	// Initializes the Users map if it is nil
	if dbStructure.Users == nil {
		dbStructure.Users = make(map[int]User)
	}

	// Check for existing email
	for _, user := range dbStructure.Users {
		if user.Email == email {
			return User{}, fmt.Errorf("email already in use")
		}
	}

	newID := len(dbStructure.Users) + 1
	newUser := User{
		ID:       newID,
		Email:    email,
		Password: password,
	}

	dbStructure.Users[newID] = newUser

	bytes, err := json.MarshalIndent(dbStructure, "", "  ")
	if err != nil {
		return User{}, err
	}

	if err := os.WriteFile(db.path, bytes, 0666); err != nil {
		return User{}, err
	}

	return newUser, nil
}

func (db *DB) GetUserByEmail(email string) (User, error) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	data, err := os.ReadFile(db.path)
	if err != nil && !os.IsNotExist(err) {
		return User{}, err
	}

	var dbStructure DBStructure

	if len(data) > 0 {
		if err := json.Unmarshal(data, &dbStructure); err != nil {
			return User{}, err
		}
	}

	// Iterate through users to find a match
	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}

	// Return an error if user not found
	return User{}, fmt.Errorf("user not found")
}
