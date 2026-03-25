package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DanilChugaev/pironman5-go/pkg/status"
	"github.com/gin-gonic/gin"
)

const (
	// dataPin  = 10 // GPIO10 = BCM 10, физ. пин 19 (MOSI)
	// ledCount = 4
	httpPort = ":34001"
)

type ResponseDTO[T any] struct {
	Success bool   `json:"success"`
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func main() {
	fmt.Println("🚀 Pironman5-Go v0.10 — go + python scripts")

	router := gin.Default()

	// == состояние сервера ==
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, ResponseDTO[string]{
			Success: true,
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Data:    "Сервер работает",
		})
	})

	// == мониторинг железа ==
	router.GET("/api/status", func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, ResponseDTO[string]{
					Success: true,
					Code:    http.StatusInternalServerError,
					Message: http.StatusText(http.StatusInternalServerError),
					Data:    "Что-то не работает",
				})
			}
		}()

		// status.PrintStatus()
		statusObj := status.GetStatus()

		c.JSON(http.StatusOK, ResponseDTO[status.RPIStatusDTO]{
			Success: true,
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Data:    statusObj,
		})
	})

	// TODO: как сделать так, чтобы при запросе несуществующего урла сервер не падал

	// == конфигурация периферии ==
	router.GET("/api/config", func(c *gin.Context) {
		// statusObj := status.GetStatus()

		// c.JSON(http.StatusOK, ResponseDTO[status.RPIStatusDTO]{
		// 	Success: true,
		// 	Code:    http.StatusOK,
		// 	Message: http.StatusText(http.StatusOK),
		// 	Data:    statusObj,
		// })
	})

	// router.POST("/api/rgb", func(c *gin.Context) {
	// 	col := c.Query("c")

	// 	c.JSON(http.StatusOK, gin.H{"success": true, "color": col})
	// })

	// == обработка несуществующих маршрутов ==
	router.NoRoute(func(c *gin.Context) {
		errorData := make(map[string]string)

		errorData["path"] = c.Request.URL.Path
		errorData["method"] = c.Request.Method

		c.JSON(404, ResponseDTO[map[string]string]{
			Success: false,
			Code:    http.StatusNotFound,
			Message: http.StatusText(http.StatusNotFound),
			Data:    errorData,
		})
	})

	log.Printf("Сервер на %s", httpPort)
	router.Run(httpPort)
}
