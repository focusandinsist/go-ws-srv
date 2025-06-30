package httpapi

import "github.com/gin-gonic/gin"

func HandleCGGetAllOnLineUsers(c *gin.Context) {
	serv.HandleCGGetAllOnLineUsers(c)
}

func HandleCGSend(c *gin.Context) {
	serv.HandleCGSend(c)
}
