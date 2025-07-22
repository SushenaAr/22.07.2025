package main

import (
	"awesomeProject/internal/config"
	"awesomeProject/internal/handler"
	"awesomeProject/internal/middleware"
	"awesomeProject/internal/model"
	"awesomeProject/internal/repository"
	"awesomeProject/internal/service"
	"awesomeProject/internal/tools"
	"github.com/gin-gonic/gin"
)

// я использовал имитацию базы in memory и файловое храниллище
// хотя для такого задания база избыточна, но на больших проектах обычно
// исполоьзуется привязка юзеров к файлам в базе
// да и в таком проекте проще хранить данные
// о файлах в оперативке для быстрого и удобного доступа

func main() {
	//inits
	cfg := config.LoadConfig()
	tools.InitLogger(cfg)

	repo := repository.NewRepository(make(map[string]*model.Task))
	s := service.NewService(repo, cfg)

	//gin init
	r := gin.Default()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.RequestLoggerMiddleware())
	handler.InitRoutes(r, s)

	if err := r.Run(cfg.AdvertisedAddr); err != nil {
		return
	}

}
