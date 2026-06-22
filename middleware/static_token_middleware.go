package middleware

import (
	"project-workflow-backend/util"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StaticTokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		staticToken := os.Getenv("STATIC_TOKEN")
		tokenString := c.GetHeader("Authorization")
		productString := c.GetHeader("Product")
		logrus.Info("tokenString:", tokenString)
		logrus.Info("Product Type:", productString)

		if tokenString == "" {
			util.RespondWithError(c, http.StatusUnauthorized, "Missing token.")
			c.Abort()
			return
		}

		if tokenString != "Bearer "+staticToken {
			logrus.Error("Unauthorized")
			util.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		c.Next()
	}
}
