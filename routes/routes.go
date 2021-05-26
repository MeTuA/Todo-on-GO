package routes

import (
	"fmt"
	"log"
	"metua/app/controller"

	"github.com/gin-gonic/gin"
)

func Init() {
	routes := gin.Default()

	routes.POST("/login", controller.Login)
	log.Fatal(routes.Run(":8080"))
	fmt.Println("Server is ready!")
}
