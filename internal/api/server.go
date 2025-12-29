package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"tokenlaunch/internal/storage"
)

type Server struct {
	echo *echo.Echo
	repo storage.MessageRepository
}

func NewServer(repo storage.MessageRepository) *Server {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	s := &Server{
		echo: e,
		repo: repo,
	}

	s.routes()

	return s
}

func (s *Server) routes() {
	s.echo.GET("/health", s.health)

	api := s.echo.Group("/api")
	api.GET("/messages", s.getMessages)
	api.GET("/messages/:id", s.getMessage)
}

func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

func (s *Server) Shutdown() error {
	return s.echo.Close()
}

func (s *Server) health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) getMessages(c echo.Context) error {
	messages, err := s.repo.FindAll(c.Request().Context(), 50, 0)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, messages)
}

func (s *Server) getMessage(c echo.Context) error {
	id := c.Param("id")

	msg, err := s.repo.FindByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	if msg == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
	}

	return c.JSON(http.StatusOK, msg)
}
