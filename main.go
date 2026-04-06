package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/DanilChugaev/pironman5-go/pkg/config"
	"github.com/DanilChugaev/pironman5-go/pkg/fan"
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

func HandleServerError(c *gin.Context, s int, m string) {
	errorData := make(map[string]string)

	errorData["path"] = c.Request.URL.Path
	errorData["method"] = c.Request.Method

	messageData := ""

	if m == "" {
		messageData = http.StatusText(s)
	} else {
		messageData = m
	}

	// добавить вывод err и errorData
	fmt.Println(messageData)

	c.JSON(s, ResponseDTO[map[string]string]{
		Success: false,
		Code:    s,
		Message: messageData,
		Data:    errorData,
	})
}

func main() {
	fmt.Println("🚀 Pironman5-Go v0.13.1")

	// == инициализируем дефолтный конфиг, если его нет ==
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Ошибка инициализации конфига: %v\n", err)
		return
	}

	// == запускаем контроль вентиляторов в фоне ==
	go fan.StartFanControlLoop(cfg.FanUpdateInterval)

	router := gin.Default()
	// gin.SetMode(gin.ReleaseMode)

	// == состояние сервера ==
	router.GET("/api/health", func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				HandleServerError(c, http.StatusInternalServerError, "")
			}
		}()

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
				HandleServerError(c, http.StatusInternalServerError, "")
			}
		}()

		// status.PrintStatus()
		statusData := status.GetStatus()

		c.JSON(http.StatusOK, ResponseDTO[status.RPIStatusDTO]{
			Success: true,
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Data:    statusData,
		})
	})

	// == получение конфига периферии ==
	router.GET("/api/config", func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				HandleServerError(c, http.StatusInternalServerError, "")
			}
		}()

		configData, err := config.LoadConfig()
		if err != nil {
			HandleServerError(c, http.StatusNotFound, "Ошибка получения конфига")
			return
		}

		c.JSON(http.StatusOK, ResponseDTO[*config.RPIConfigDTO]{
			Success: true,
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Data:    configData,
		})
	})

	// == обновление конфига периферии ==
	router.PUT("/api/config", func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				HandleServerError(c, http.StatusInternalServerError, "")
			}
		}()

		// Парсим входящие данные в структуру частичного обновления
		var updateData config.RPIConfigUpdate
		if err := c.ShouldBindJSON(&updateData); err != nil {
			HandleServerError(c, http.StatusBadRequest, "Некорректный объект в запросе")
			return
		}

		// Вызываем функцию обновления конфига
		updatedConfig, err := config.UpdateConfig(&updateData)
		if err != nil {
			HandleServerError(c, http.StatusInternalServerError, "Ошибка обновления конфига")
			return
		}

		c.JSON(http.StatusOK, ResponseDTO[*config.RPIConfigDTO]{
			Success: true,
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Data:    updatedConfig,
		})
	})

	// == обработка несуществующих роутов ==
	router.NoRoute(func(c *gin.Context) {
		HandleServerError(c, http.StatusNotFound, "Несуществующий роут")
	})

	log.Printf("Сервер на %s", httpPort)
	router.Run(httpPort)
}
