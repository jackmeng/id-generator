package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"snowflake-server/app/snowflake"
)

func main() {
	// 定义命令行参数
	host := flag.String("host", "127.0.0.1", "server host")
	port := flag.Int("port", 6060, "server port")
	flag.Parse()

	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/generate/:nodeNo", func(c *gin.Context) {
		nodeNo := c.Param("nodeNo")
		nodeNoInt64, err := strconv.ParseInt(nodeNo, 10, 64)
		if err != nil {
			c.JSON(200, gin.H{
				"code":    404,
				"message": fmt.Sprintf("nodeNo params error: %s", err.Error()),
				"data": gin.H{
					"id": 0,
				},
			})
			return
		}
		id, err := snowflake.Generate(nodeNoInt64)
		if err != nil {
			c.JSON(200, gin.H{
				"code":    404,
				"message": fmt.Sprintf("id generate error: %s", err.Error()),
				"data": gin.H{
					"id": id,
				},
			})
			return
		}
		c.JSON(200, gin.H{
			"code":    200,
			"message": "success",
			"data": gin.H{
				"id": id,
			},
		})
	})

	// 拼接监听地址
	addr := fmt.Sprintf("%s:%d", *host, *port)
	fmt.Println("Starting server on: ", addr)

	err := router.Run(addr)
	if err != nil {
		return
	} // listen and serve on 0.0.0.0:8080
}
