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
	r.GET("/devices", h.list)
	r.GET("/devices/:id", h.get)
	r.PUT("/devices/:id", h.update)
	r.PATCH("/devices/:id", h.patch)
}

type deviceRequest struct {
	Name  string `json:"name"`
	Brand string `json:"brand"`
}

// patchRequest uses pointers so we can tell which fields were sent.
type patchRequest struct {
	Name  *string `json:"name"`
	Brand *string `json:"brand"`
}

func (h *Handler) create(c *gin.Context) {
	req, ok := bindDevice(c)
	if !ok {
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

// update replaces name and brand. created_at never changes.
func (h *Handler) update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	req, ok := bindDevice(c)
	if !ok {
		return
	}

	d, err := h.store.Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	d.Name = req.Name
	d.Brand = req.Brand

	if err := h.store.Update(c.Request.Context(), d); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *Handler) patch(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req patchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return
	}
	if (req.Name != nil && *req.Name == "") || (req.Brand != nil && *req.Brand == "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and brand cannot be empty"})
		return
	}

	d, err := h.store.Get(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	if req.Name != nil {
		d.Name = *req.Name
	}
	if req.Brand != nil {
		d.Brand = *req.Brand
	}

	if err := h.store.Update(c.Request.Context(), d); err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

// list handles both "list all" and "search by brand" via ?brand=
func (h *Handler) list(c *gin.Context) {
	devices, err := h.store.List(c.Request.Context(), c.Query("brand"))
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, devices)
}

func bindDevice(c *gin.Context) (deviceRequest, bool) {
	var req deviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return req, false
	}
	if req.Name == "" || req.Brand == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and brand are required"})
		return req, false
	}
	return req, true
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
