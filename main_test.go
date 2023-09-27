package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
	"testing"
)

func TestSetCreds(t *testing.T) {
	// Call the setup function to initialize the test database
	setupSetCredsTest(t)
	// defer cleanup() // Ensure the test database is cleaned up after the test

	tests := []struct {
		profileName string
		url         string
		username    string
		password    string
	}{
		{"TestProfile1", "http://example.com", "testuser1", "testpass1"},
		{"TestProfile2", "http://test.com", "testuser2", "testpass2"},
	}

	for _, test := range tests {
		t.Run(test.profileName, func(t *testing.T) {
			// Call the setCreds function with test data and the test database
			setCreds(test.profileName, test.url, test.username, test.password)

			// Rest of your test logic
		})
	}
}

func setupSetCredsTest(t *testing.T) (*sql.DB, func()) {
	// Create and initialize the test database
	db, err := sql.Open("sqlite3", "credentials.db")
	if err != nil {
		t.Fatalf("error opening test database: %v", err)
	}

	// Create the table and perform any necessary setup steps
	createTableSQL := `
        CREATE TABLE IF NOT EXISTS credentials (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            profileName TEXT,
            url TEXT,
            username TEXT,
            password TEXT
        );
    `
	_, err = db.Exec(createTableSQL)
	if err != nil {
		t.Fatalf("error initializing test database: %v", err)
	}

	// Return a cleanup function to close the test database
	cleanup := func() {
		db.Close()
		// Add any additional cleanup steps, if necessary
	}

	return db, cleanup
}

func TestRetrieveCredentials(t *testing.T) {
	// Replace with your test database file name
	testDBName := "credentials.db"

	// Assuming that you have previously stored credentials in the test database
	tests := []struct {
		profileName   string
		expectedError bool // Expect an error for a non-existent profile
	}{
		{"TestProfile1", false},      // Expecting valid credentials
		{"NonExistentProfile", true}, // Expecting an error for a non-existent profile
	}

	for _, test := range tests {
		t.Run(test.profileName, func(t *testing.T) {
			retrievedURL, retrievedUser, retrievedPass := retrieveCredentials(testDBName, test.profileName)

			t.Logf("Retrieved URL: %s", retrievedURL)
			t.Logf("Retrieved Username: %s", retrievedUser)
			t.Logf("Retrieved Password: %s", retrievedPass)

			if test.expectedError {
				if retrievedURL != "" || retrievedUser != "" || retrievedPass != "" {
					t.Errorf("expected an error, but retrieved credentials: URL=%s, User=%s, Pass=%s", retrievedURL, retrievedUser, retrievedPass)
				}
			} else {
				if retrievedURL == "" || retrievedUser == "" || retrievedPass == "" {
					t.Errorf("expected (URL!=, User!=, Pass!=), got (URL=%s, User=%s, Pass=%s)", retrievedURL, retrievedUser, retrievedPass)
				}
			}
		})
	}
}
