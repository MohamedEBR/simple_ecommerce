package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/MohamedEBR/simple_ecommerce/cart-service/internal/db"
)

// Wire up REST routes with distinct paths to avoid collisions.
func RegisterCartRoutes(mux *http.ServeMux, conn *sql.DB) {
	// Cart-level
	mux.HandleFunc("POST /carts", createCart(conn))                          // create
	mux.HandleFunc("GET /carts/{cartID}", viewCart(conn))                    // read
	mux.HandleFunc("DELETE /carts/{cartID}/items", emptyCart(conn))          // empty all items

	// Item-level
	mux.HandleFunc("POST /carts/{cartID}/items", addItem(conn))              // add
	mux.HandleFunc("DELETE /carts/{cartID}/items/{productID}", deleteItem(conn))   // remove one
	mux.HandleFunc("PATCH /carts/{cartID}/items/{productID}", decreaseItem(conn))  // decrease qty
}

// ========== CART LEVEL ==========

// POST /carts   Body: { "user_id": "uuid" }
func createCart(conn *sql.DB) http.HandlerFunc {
	type req struct{ UserID string `json:"user_id"` }
	type resp struct{ CartID string `json:"cart_id"` }

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r); return
		}
		var body req
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == "" {
			http.Error(w, "invalid body", http.StatusBadRequest); return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		id, err := db.CreateCart(ctx, conn, body.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError); return
		}
		writeJSON(w, http.StatusCreated, resp{CartID: id})
	}
}

// GET /carts/{cartID}
func viewCart(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r); return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/") // ["carts","{cartID}"]
		if len(parts) != 2 || parts[0] != "carts" {
			http.NotFound(w, r); return
		}
		cartID := parts[1]

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		rows, err := db.GetCart(ctx, conn, cartID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError); return
		}
		writeJSON(w, http.StatusOK, map[string]any{"cart_id": cartID, "items": rows})
	}
}

// DELETE /carts/{cartID}/items  → empty the cart (remove all items)
func emptyCart(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.NotFound(w, r); return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/") // ["carts","{cartID}","items"]
		if len(parts) != 3 || parts[0] != "carts" || parts[2] != "items" {
			http.NotFound(w, r); return
		}
		cartID := parts[1]

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := db.EmptyCart(ctx, conn, cartID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError); return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// ========== ITEM LEVEL ==========

// DELETE /carts/{cartID}/items/{productID}
func deleteItem(conn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.NotFound(w, r); return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/") // ["carts","{cartID}","items","{productID}"]
		if len(parts) != 4 || parts[0] != "carts" || parts[2] != "items" {
			http.NotFound(w, r); return
		}
		cartID, productID := parts[1], parts[3]

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		// RemoveItem returns only error (no sql.Result), so don't use `_, err :=`
		if err := db.RemoveItem(ctx, conn, cartID, productID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError); return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// POST /carts/{cartID}/items   Body: { "product_id": "...", "quantity": 2 }
func addItem(conn *sql.DB) http.HandlerFunc {
	type req struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.NotFound(w, r); return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/") // ["carts","{cartID}","items"]
		if len(parts) != 3 || parts[0] != "carts" || parts[2] != "items" {
			http.NotFound(w, r); return
		}
		cartID := parts[1]

		var body req
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ProductID == "" || body.Quantity <= 0 {
			http.Error(w, "invalid body", http.StatusBadRequest); return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := db.AddItem(ctx, conn, cartID, body.ProductID, body.Quantity); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError); return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

// PATCH /carts/{cartID}/items/{productID}   Body: { "quantity": N }   (decrease by N)
func decreaseItem(conn *sql.DB) http.HandlerFunc {
	type req struct {
		Quantity int `json:"quantity"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.NotFound(w, r); return
		}
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/") // ["carts","{cartID}","items","{productID}"]
		if len(parts) != 4 || parts[0] != "carts" || parts[2] != "items" {
			http.NotFound(w, r); return
		}
		cartID, productID := parts[1], parts[3]

		var body req
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Quantity <= 0 {
			http.Error(w, "invalid body: need positive 'quantity'", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		if err := db.DecreaseItem(ctx, conn, cartID, productID, body.Quantity); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError); return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
