package handler

import (
	"blog-go-microservice/app/service/article/internal/app/service"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

func GetUser(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "parameter error")
	}
	user, err := service.UserRepository.Get(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, user)
}
