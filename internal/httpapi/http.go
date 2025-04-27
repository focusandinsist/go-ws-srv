package httpapi

import (
	"net/http"

	"go-ws-srv/internal/connection"

	"github.com/gin-gonic/gin"
)

func RunHTTPServer(connMgr *connection.ConnectionManager) {
	r := gin.Default()

	r.GET("/online", func(c *gin.Context) {
		users := connMgr.GetAllUserIDs()
		c.JSON(http.StatusOK, gin.H{"online_users": users})
	})

	r.POST("/send", func(c *gin.Context) {
		var req struct {
			UserID  string `json:"user_id"`
			Message string `json:"message"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := connMgr.SendMessageToUser(req.UserID, []byte(req.Message))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "sent"})
	})

	go r.Run(":8080")
}
