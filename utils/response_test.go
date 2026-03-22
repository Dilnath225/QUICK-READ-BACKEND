package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSuccessResponse(t *testing.T) {
	// Arrange test context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Act
	testData := map[string]string{"user": "alice"}
	SuccessResponse(c, "Data fetched successfully", testData)

	// Assert Status Code
	assert.Equal(t, http.StatusOK, w.Code)

	// Assert Body Payload Match
	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)

	assert.NoError(t, err)
	assert.True(t, response.Status)
	assert.Equal(t, "Data fetched successfully", response.Message)

	// Type casting back to standard map since JSON unmarshals interfaces to maps
	dataMap := response.Data.(map[string]interface{})
	assert.Equal(t, "alice", dataMap["user"])
}

func TestErrorResponse(t *testing.T) {
	// Arrange
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Act
	ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response Response
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.False(t, response.Status)
	assert.Equal(t, "Invalid credentials", response.Message)
	assert.Equal(t, "Invalid credentials", response.Error)
}
