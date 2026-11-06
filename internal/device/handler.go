package device

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Register(r *gin.Engine) {
	r.POST("/devices", h.create)
	r.GET("/devices/:id", h.get)
}

type deviceRequest struct {
	Name  string `json:"name"`
	Brand string `json:"brand"`
}

func (h *Handler) create(c *gin.Context) {
	var req deviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}
	if req.Name == "" || req.Brand == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and brand are required"})
		return
	}

	d := Device{
		Name:      req.Name,
		Brand:     req.Brand,
		CreatedAt: time.Now().UTC(),
	}
	if err := h.store.Create(c.Request.Context(), &d); err != nil {
		handleError(c, err)
		return
	}

	c.Header("Location", fmt.Sprintf("/devices/%d", d.ID))
	c.JSON(http.StatusCreated, d)
}

func (h *Handler) get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	d, err := h.store.Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func handleError(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	log.Println("error:", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
