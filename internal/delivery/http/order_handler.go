package http

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"github.com/Sergi-Ch/WB_L0_2025/domain"
	"github.com/Sergi-Ch/WB_L0_2025/internal/service"
	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(s *service.OrderService) *OrderHandler {
	return &OrderHandler{service: s}
}

//go:embed web/*
var content embed.FS

func (h *OrderHandler) RegisterRoutes(r chi.Router) {
	// Фронт
	webFS, err := fs.Sub(content, "web")
	if err != nil {
		log.Printf("Failed to create sub filesystem: %v", err)

		fs := http.FileServer(http.Dir("./internal/delivery/http/web"))
		r.Handle("/*", fs)
		log.Println("Serving static files with FileServer")
	} else {
		fs := http.FileServer(http.FS(webFS))
		r.Handle("/*", fs)
		log.Println("Serving static files with embed")
	}

	// endpoints
	r.Get("/order/{order_uid}", h.GetOrderByID)
	r.Post("/order", h.CreateOrder)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
}

// GET /order/{order_uid}
func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "order_uid")
	log.Printf("Fetching order: %s", orderID)

	order, err := h.service.GetOrderByID(r.Context(), orderID)
	if err != nil {
		log.Printf("Order not found: %s, error: %v", orderID, err)
		http.Error(w, "order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		log.Printf("JSON encoding error: %v", err)
		http.Error(w, "error of encoding json", http.StatusInternalServerError)
	}
}

// POST /order для теста
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var order domain.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if err := h.service.SaveOrder(r.Context(), &order); err != nil {
		log.Printf("failed to save order: %v", err)
		http.Error(w, "failed to save order", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
