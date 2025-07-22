package handler

import (
	"awesomeProject/internal/service"
	"github.com/gin-gonic/gin"
)

func InitRoutes(r *gin.Engine, s *service.Service) {
	r.GET("/resources/zip/:filename", s.GetZip)
	r.POST("/task", s.CreateTaskForZIPAchieve)
	r.PATCH("/task", s.AddFileIntoTask)
	r.GET("/task/status", s.GetStatusTask)
}
