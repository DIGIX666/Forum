package data

import (
	"os"
	"testing"
)

func TestCreateDataBase(t *testing.T) {
	// delete the database if it exists
	os.Remove("./usersForum.db")
	CreateDataBase()
	if Db == nil {
		t.Fatal("The database should not be nil")
	}
}

func TestDataBaseRegister(t *testing.T) {
	os.Remove("./usersForum.db")
	CreateDataBase()
	ok := DataBaseRegister("test@example.com", "motdepasse")
	if !ok {
		t.Error("The registration should have been successful")
	}
	// Test register with an existing email
	ok2 := DataBaseRegister("test@example.com", "motdepasse")
	if ok2 {
		t.Error("email already exists")
	}
}
