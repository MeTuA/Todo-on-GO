package routes

import (
	"fmt"
	"log"
	"metua/app/controller"

	"github.com/gin-gonic/gin"
)

func Init() {
	routes := gin.Default()
	fmt.Println("Server is ready!")

	routes.POST("/login", controller.TokenAuthMiddleware(), controller.Login)
	routes.POST("/todo", controller.TokenAuthMiddleware(), controller.CreateTodo)
	routes.POST("/logout", controller.Logout)

	log.Fatal(routes.Run(":8080"))

}
