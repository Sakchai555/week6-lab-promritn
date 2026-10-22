package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"moviehub/auth"
	"moviehub/handler"
	"moviehub/mw"
	"moviehub/store"
)

func main() {
	// Validate environment variables
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET ต้องยาวอย่างน้อย 32 ตัวอักษร")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("กรุณากำหนด DATABASE_URL")
	}

	// Connect to database
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal("สร้างการเชื่อมต่อฐานข้อมูลไม่สำเร็จ: ", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatal("ต่อฐานข้อมูลไม่ติด: ", err)
	}

	// JWT
	tickets := &auth.Issuer{
		Secret: []byte(secret),
		TTL:    time.Hour,
	}

	// Handlers
	h := &handler.Handler{
		Store:   &store.Store{Pool: pool},
		Tickets: tickets,
	}

	// Routes
	mux := http.NewServeMux()

	// Authentication
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc(
		"GET /api/me",
		mw.RequireLogin(tickets, h.Me),
	)

	// Movies and reviews
	mux.HandleFunc("GET /api/movies", h.SearchMovies)
	mux.HandleFunc("GET /api/movies/{id}/reviews", h.ListReviews)
	mux.HandleFunc(
		"POST /api/movies/{id}/reviews",
		mw.RequireLogin(tickets, h.AddReview),
	)
	mux.HandleFunc(
		"DELETE /api/reviews/{reviewId}",
		mw.RequireLogin(tickets, h.DeleteReview),
	)

	// Admin
	mux.HandleFunc(
		"GET /api/admin/members",
		mw.RequireLogin(
			tickets,
			mw.RequireRole("admin", h.AdminMembers),
		),
	)
	mux.HandleFunc(
		"GET /api/admin/stats",
		mw.RequireLogin(
			tickets,
			mw.RequireRole("admin", h.AdminStats),
		),
	)

	// CORS
	origins := []string{
		"http://localhost:3000",
		"https://moviehub-670710337-wrhu.vercel.app",
	}

	// Start server
	log.Println("MovieHub API ฟังอยู่ที่ :8080")

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mw.Logging(mw.CORS(origins, mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
