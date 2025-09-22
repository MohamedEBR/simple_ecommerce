package main

import (
	"log"
	"net/http"

	"github.com/MohamedEBR/simple_ecommerce/cart-service/internal/api"
	"github.com/MohamedEBR/simple_ecommerce/cart-service/internal/config"
	"github.com/MohamedEBR/simple_ecommerce/cart-service/internal/db"
)

func main(){
	cfg := config.Load()

	conn, err := db.Connect(cfg.DatabaseURL)

	if err != nil {
		log.Fatal("DB connection failed:", err)
	}
	defer conn.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	api.RegisterCartRoutes(mux, conn)

	log.Println("Cart service listening on :" + cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, mux))
}