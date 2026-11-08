package device

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler serves the /devices endpoints.
type Handler struct {
	store *Store
}

func NewHandler(store *Store) *Handler {
	return &Handler{store: store}
}

// Register adds the device routes to r.
func (h *Handler) Register(r *gin.Engine) {
	r.POST("/devices", h.create)
	r.GET("/devices", h.list)
	r.GET("/devices/:id", h.get)
	r.PUT("/devices/:id", h.update)
	r.PATCH("/devices/:id", h.patch)
	r.DELETE("/devices/:id", h.delete)
}

// deviceRequest is the body for POST and PUT, both fields are required.
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

// patch only changes the fields that are in the body.
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
	if req.Name != nil {
		*req.Name = strings.TrimSpace(*req.Name)
	}
	if req.Brand != nil {
		*req.Brand = strings.TrimSpace(*req.Brand)
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

func (h *Handler) delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := h.store.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// list handles both "list all" and "search by brand" via ?brand=
func (h *Handler) list(c *gin.Context) {
	brand := strings.TrimSpace(c.Query("brand"))
	devices, err := h.store.List(c.Request.Context(), brand)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, devices)
}

// bindDevice reads and validates the body. On failure it writes the
// 400 response itself, so callers just return.
func bindDevice(c *gin.Context) (deviceRequest, bool) {
	var req deviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
		return req, false
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Brand = strings.TrimSpace(req.Brand)
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

// handleError sends 404 for missing devices and 500 for anything else.
// The real error is only logged, not returned to the client.
func handleError(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	log.Println("error:", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
