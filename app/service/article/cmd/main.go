package main

import (
	"blog-go-microservice/app/service/article/internal/app/engine"
	"blog-go-microservice/app/service/article/internal/app/model"
	"blog-go-microservice/app/service/article/internal/app/service"
	"fmt"
	"github.com/douyu/jupiter"
	"log"
)

func main() {
	eng := engine.NewEngine()
	eng.RegisterHooks(jupiter.StageAfterStop, func() error {
		fmt.Println("exit jupiter app ...")
		return nil
	})

	model.Init()
	service.Init()
	if err := eng.Run(); err != nil {
		log.Fatal(err)
	}
}
