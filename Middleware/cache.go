package middleware

import (
	"QUICK-READ-SYSTEM/config"
	"crypto/sha256"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)