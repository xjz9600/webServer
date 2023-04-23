package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func GetDummyEndpoint(c *gin.Context) {
	fmt.Println("start ")
	resp := map[string]string{"hello": "world"}

	c.JSON(200, resp)

}

func main() {

	api := gin.Default()
	api.Use(DummyMiddleware, DummyMiddleware2)

	api.GET("/dummy", GetDummyEndpoint)

	api.Run(":5000")
}

func DummyMiddleware(c *gin.Context) {
	resp := map[string]string{"do not hello": "world"}
	c.JSON(404, resp)
	fmt.Println(" before Im a dummy!")
	c.Next()
	fmt.Println(" after Im a dummy!")

}
func DummyMiddleware2(c *gin.Context) {
	fmt.Println(" before Im a dummy2!")
	//c.Next()
	fmt.Println(" after Im a dummy2!")
}
