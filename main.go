package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"credits/database"
	"credits/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Parse port from flag or env
	port := flag.String("port", "", "server port (default 8080)")
	flag.Parse()

	addr := ":8080"
	if *port != "" {
		addr = ":" + *port
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		addr = ":" + envPort
	}

	// Determine database path
	dbPath := "credits.db"

	// Initialize database
	database.Init(dbPath)

	// Setup Gin router
	r := gin.Default()

	// API routes
	api := r.Group("/api")
	{
		// Task management (parent)
		api.GET("/tasks", handlers.GetTasks)
		api.POST("/tasks", handlers.CreateTask)
		api.DELETE("/tasks/:id", handlers.DeleteTask)

		// Child tasks
		api.GET("/child/tasks", handlers.GetChildTasks)
		api.POST("/child/complete-task", handlers.CompleteTask)

		// Points
		api.GET("/points", handlers.GetPoints)
		api.POST("/points/adjust", handlers.AdjustPoints)
		api.GET("/points/history", handlers.GetPointHistory)

		// Redeem items
		api.GET("/redeem-items", handlers.GetRedeemItems)
		api.POST("/redeem-items", handlers.CreateRedeemItem)
		api.DELETE("/redeem-items/:id", handlers.DeleteRedeemItem)

		// Redeem
		api.POST("/redeem", handlers.RedeemItem)
	}

	// Static files
	staticDir := getStaticDir()
	r.StaticFile("/parent.html", filepath.Join(staticDir, "parent.html"))
	r.StaticFile("/child.html", filepath.Join(staticDir, "child.html"))

	// Redirect root to child page
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(staticDir, "child.html"))
	})

	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("  Child page:  http://localhost%s/child", addr)
	log.Printf("  Parent page: http://localhost%s/parent", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getStaticDir returns the path to the static directory
func getStaticDir() string {
	// Try common locations
	dirs := []string{
		"static",
		filepath.Join(".", "static"),
	}

	// Also try relative to executable
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Join(filepath.Dir(exe), "static"))
	}

	for _, d := range dirs {
		if _, err := os.Stat(filepath.Join(d, "child.html")); err == nil {
			return d
		}
	}

	return "static"
}
