package test

import (
	"testing"

	"github.com/Kaikai20040827/graduation/internal/service"
)

func TestUserServiceCreateAuthenticate(t *testing.T) {
	db := newTestDB(t)
	us := service.NewUserService(db)

	user, err := us.CreateUser("alice", "alice@example.com", "password123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.Password != "" {
		t.Fatalf("expected password stripped")
	}

	if _, err := us.Authenticate("alice@example.com", "password123"); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if _, err := us.Authenticate("alice@example.com", "wrong"); err == nil {
		t.Fatalf("expected invalid credentials")
	}
}

func TestUserServiceChangePassword(t *testing.T) {
	db := newTestDB(t)
	us := service.NewUserService(db)

	user, err := us.CreateUser("bob", "bob@example.com", "oldpass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := us.ChangePassword(user.ID, "oldpass", "newpass"); err != nil {
		t.Fatalf("ChangePassword: %v", err)
	}
	if _, err := us.Authenticate("bob@example.com", "newpass"); err != nil {
		t.Fatalf("Authenticate newpass: %v", err)
	}
}

func TestUserServiceUpdateProfile(t *testing.T) {
	db := newTestDB(t)
	us := service.NewUserService(db)

	user, err := us.CreateUser("carol", "carol@example.com", "password")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	updated, err := us.UpdateProfile(user.ID, "carol2", "")
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}
	if updated.Username != "carol2" {
		t.Fatalf("username not updated")
	}
}

func TestUserServiceChangeUsername(t *testing.T) {
	db := newTestDB(t)
	us := service.NewUserService(db)

	if _, err := us.CreateUser("dave", "dave@example.com", "password"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := us.ChangeUsername("dave@example.com", "dave2"); err != nil {
		t.Fatalf("ChangeUsername: %v", err)
	}
}

func TestUserServiceDeleteUser(t *testing.T) {
	db := newTestDB(t)
	us := service.NewUserService(db)

	if _, err := us.CreateUser("erin", "erin@example.com", "password"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if err := us.DeleteUser("erin@example.com", "password"); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	if _, err := us.Authenticate("erin@example.com", "password"); err == nil {
		t.Fatalf("expected user deleted")
	}
}
