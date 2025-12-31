package api

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"tokenlaunch/internal/redis"
	"tokenlaunch/internal/storage"
)

//go:embed templates/*.html
var templateFS embed.FS

type Server struct {
	echo      *echo.Echo
	repo      storage.MessageRepository
	redis     *redis.Client
	templates *template.Template
	sse       *SSEBroker
}

type SSEBroker struct {
	clients map[chan string]bool
	mu      sync.RWMutex
}

func NewSSEBroker() *SSEBroker {
	return &SSEBroker{clients: make(map[chan string]bool)}
}

func (b *SSEBroker) Subscribe() chan string {
	ch := make(chan string, 10)
	b.mu.Lock()
	b.clients[ch] = true
	b.mu.Unlock()
	return ch
}

func (b *SSEBroker) Unsubscribe(ch chan string) {
	b.mu.Lock()
	delete(b.clients, ch)
	close(ch)
	b.mu.Unlock()
}

func (b *SSEBroker) Broadcast(msg string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

type Stats struct {
	Total        int
	Launches     int
	Endorsements int
}

type MessageView struct {
	ID             string
	Username       string
	Content        string
	Classification string
	TimeAgo        string
}

type AccountView struct {
	Username string
}

func NewServer(repo storage.MessageRepository, rdb *redis.Client) *Server {
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	tmpl := template.Must(template.ParseFS(templateFS, "templates/*.html"))

	s := &Server{
		echo:      e,
		repo:      repo,
		redis:     rdb,
		templates: tmpl,
		sse:       NewSSEBroker(),
	}

	s.routes()

	return s
}

func (s *Server) routes() {
	s.echo.GET("/", s.index)
	s.echo.GET("/health", s.health)
	s.echo.GET("/api/stats", s.stats)
	s.echo.GET("/api/messages", s.getMessages)
	s.echo.GET("/api/messages/:id", s.getMessage)
	s.echo.GET("/api/events", s.events)

	// Account management
	s.echo.GET("/api/accounts", s.getAccounts)
	s.echo.POST("/api/accounts", s.addAccount)
	s.echo.DELETE("/api/accounts/:username", s.removeAccount)
}

func (s *Server) Start(addr string) error {
	return s.echo.Start(addr)
}

func (s *Server) Shutdown() error {
	return s.echo.Close()
}

func (s *Server) Broadcast(msg string) {
	s.sse.Broadcast(msg)
}

func (s *Server) index(c echo.Context) error {
	messages, _ := s.repo.FindAll(c.Request().Context(), 20, 0)
	accounts, _ := s.redis.GetAccounts(c.Request().Context())

	views := make([]MessageView, len(messages))
	for i, m := range messages {
		views[i] = MessageView{
			ID:             m.ID,
			Username:       m.Username,
			Content:        m.Content,
			Classification: "",
			TimeAgo:        timeAgo(m.CreatedAt),
		}
	}

	accountViews := make([]AccountView, len(accounts))
	for i, a := range accounts {
		accountViews[i] = AccountView{Username: a}
	}

	total, launches, endorsements, _ := s.repo.GetStats(c.Request().Context())

	data := map[string]any{
		"Stats":    Stats{Total: total, Launches: launches, Endorsements: endorsements},
		"Messages": views,
		"Accounts": accountViews,
	}

	return s.render(c, "index.html", data)
}

func (s *Server) stats(c echo.Context) error {
	total, launches, endorsements, err := s.repo.GetStats(c.Request().Context())
	if err != nil {
		total, launches, endorsements = 0, 0, 0
	}
	stats := Stats{Total: total, Launches: launches, Endorsements: endorsements}
	return s.render(c, "stats", stats)
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

func (s *Server) getAccounts(c echo.Context) error {
	accounts, err := s.redis.GetAccounts(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return s.render(c, "accounts", accounts)
}

func (s *Server) addAccount(c echo.Context) error {
	username := strings.TrimSpace(c.FormValue("username"))
	username = strings.TrimPrefix(username, "@")

	if username == "" {
		return c.HTML(http.StatusBadRequest, `<div class="error">Username required</div>`)
	}

	exists, _ := s.redis.AccountExists(c.Request().Context(), username)
	if exists {
		return c.HTML(http.StatusConflict, `<div class="error">Already tracking</div>`)
	}

	if err := s.redis.AddAccount(c.Request().Context(), username); err != nil {
		return c.HTML(http.StatusInternalServerError, `<div class="error">Failed to add</div>`)
	}

	// Return updated account list
	accounts, _ := s.redis.GetAccounts(c.Request().Context())
	return s.render(c, "accounts", accounts)
}

func (s *Server) removeAccount(c echo.Context) error {
	username := c.Param("username")

	if err := s.redis.RemoveAccount(c.Request().Context(), username); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Return updated account list
	accounts, _ := s.redis.GetAccounts(c.Request().Context())
	return s.render(c, "accounts", accounts)
}

func (s *Server) events(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("X-Accel-Buffering", "no")

	ch := s.sse.Subscribe()
	defer s.sse.Unsubscribe(ch)

	// Send initial ping
	fmt.Fprintf(c.Response(), ": ping\n\n")
	c.Response().Flush()

	for {
		select {
		case <-c.Request().Context().Done():
			return nil
		case msg := <-ch:
			// SSE format: event name + data lines
			fmt.Fprintf(c.Response(), "event: message\n")
			lines := strings.Split(msg, "\n")
			for _, line := range lines {
				fmt.Fprintf(c.Response(), "data: %s\n", line)
			}
			fmt.Fprintf(c.Response(), "\n")
			c.Response().Flush()
		}
	}
}

func (s *Server) render(c echo.Context, name string, data any) error {
	c.Response().Header().Set("Content-Type", "text/html")
	err := s.templates.ExecuteTemplate(c.Response(), name, data)
	if err != nil {
		c.Logger().Error(err)
	}
	return err
}

func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
