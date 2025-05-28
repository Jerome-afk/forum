package utils

import (
	"testing"
)

func TestPasswordHashing(t *testing.T) {
	// Test hashing a password
	password := "testpassword123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	if hashedPassword == "" {
		t.Fatal("Hashed password is empty")
	}
	if hashedPassword == password {
		t.Fatal("Hashed password is the same as the original password")
	}

	// Test checking a correct password
	if !CheckPasswordHash(password, hashedPassword) {
		t.Fatal("Password check failed for correct password")
	}

	// Test checking an incorrect password
	if CheckPasswordHash("wrongpassword", hashedPassword) {
		t.Fatal("Password check passed for incorrect password")
	}

	// Test hashing different passwords produces different hashes
	hashedPassword2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash second password: %v", err)
	}
	if hashedPassword == hashedPassword2 {
		t.Fatal("Two hashes of the same password are identical, which should not happen with bcrypt")
	}

	// Test hashing a different password
	differentPassword := "differentpassword456"
	hashedDifferentPassword, err := HashPassword(differentPassword)
	if err != nil {
		t.Fatalf("Failed to hash different password: %v", err)
	}
	if hashedDifferentPassword == hashedPassword {
		t.Fatal("Hashes of different passwords are the same")
	}

	// Test checking the different password against its own hash
	if !CheckPasswordHash(differentPassword, hashedDifferentPassword) {
		t.Fatal("Password check failed for correct different password")
	}

	// Test checking the different password against the first hash
	if CheckPasswordHash(differentPassword, hashedPassword) {
		t.Fatal("Password check passed for incorrect password against first hash")
	}

	// Test checking the first password against the different password hash
	if CheckPasswordHash(password, hashedDifferentPassword) {
		t.Fatal("Password check passed for incorrect password against different hash")
	}
}
