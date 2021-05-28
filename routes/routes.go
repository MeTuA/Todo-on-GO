package routes

import (
	"fmt"
	"log"
	"metua/app/controller"

	"github.com/gin-gonic/gin"
)

func Init() {
	routes := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	fmt.Println("Server is ready!")
	routes.POST("/login", controller.Login)
	routes.POST("/todo", controller.CreateTodo)
	log.Fatal(routes.Run(":8080"))

}
