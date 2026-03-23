package services

import (
	"testing"
	"QUICK-READ-BACKEND/config"
	"QUICK-READ-BACKEND/models"
	"github.com/stretchr/testify/assert"
)

func TestAuthRegisterAndLogin(t *testing.T) {
	// Clean the user table for isolation
	config.DB.Exec("DELETE FROM users")
	
	authService := new(AuthService)

	// 1. Test Valid Registration
	err := authService.Register("Test User", "test@user.com", "password123", "123 Test St", "patient")
	assert.NoError(t, err, "Register should succeed without error")

	// Verify DB inserted it
	var user models.User
	err = config.DB.Where("email = ?", "test@user.com").First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test User", user.Name)
	assert.False(t, user.IsVerified, "New users should be unverified initially")
	assert.NotEmpty(t, user.Password, "Password should be hashed and saved")

	// 2. Test Invalid Login (Unverified Account)
	_, _, err = authService.Login("test@user.com", "password123")
	assert.Error(t, err, "Login should fail for unverified user")
	assert.Contains(t, err.Error(), "account not verified")

	// 3. Mark as verified and test Valid Login
	config.DB.Model(&user).Update("is_verified", true)

	token, loggedInUser, err := authService.Login("test@user.com", "password123")
	assert.NoError(t, err, "Login should succeed for properly verified user")
	assert.NotEmpty(t, token, "Should generate and return a valid JWT string")
	assert.Equal(t, "test@user.com", loggedInUser.Email)
	assert.Equal(t, "patient", loggedInUser.Role)

	// 4. Test Invalid Login (Wrong Password)
	_, _, err = authService.Login("test@user.com", "wrongpass")
	assert.Error(t, err, "Login with bad password should gracefully fail")
	assert.Contains(t, err.Error(), "invalid email or password")
}

func TestUpdateProfile(t *testing.T) {
	// Setup user
	config.DB.Exec("DELETE FROM users")
	user := models.User{
		Name:    "Old Name",
		Email:   "update@test.com",
		Address: "Old Address",
	}
	config.DB.Create(&user)

	authService := new(AuthService)

	// Perform specific UpdateProfile update mapping test
	updatedUser, err := authService.UpdateProfile(user.ID, "New Name", "2000-01-01", "New Address", "1234567890", "", "")
	
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updatedUser.Name)
	assert.Equal(t, "New Address", updatedUser.Address)
	assert.Equal(t, "2000-01-01", updatedUser.DateOfBirth)
	assert.Equal(t, "update@test.com", updatedUser.Email, "Email should not have changed")
}

func TestDeleteUser(t *testing.T) {
	// Setup user
	config.DB.Exec("DELETE FROM users")
	user := models.User{
		Name:  "To Be Deleted",
		Email: "delete@test.com",
	}
	config.DB.Create(&user)

	authService := new(AuthService)

	// Delete user
	err := authService.DeleteUser(user.ID)
	assert.NoError(t, err)

	// Verify user is soft deleted (not found by First but exists in DB with DeletedAt set)
	var foundUser models.User
	err = config.DB.First(&foundUser, user.ID).Error
	assert.Error(t, err, "User should not be found after deletion")

	// Verify it still exists in DB (soft delete check)
	var deletedUser models.User
	err = config.DB.Unscoped().First(&deletedUser, user.ID).Error
	assert.NoError(t, err)
	assert.NotNil(t, deletedUser.DeletedAt)
}
