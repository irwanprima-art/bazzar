package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/irwan/bazzar/internal/config"
	"github.com/irwan/bazzar/internal/middleware"

	authRepo "github.com/irwan/bazzar/internal/modules/auth/repository"
	authUC "github.com/irwan/bazzar/internal/modules/auth/usecase"
	authHandler "github.com/irwan/bazzar/internal/modules/auth/transport"

	eventRepo "github.com/irwan/bazzar/internal/modules/event/repository"
	eventUC "github.com/irwan/bazzar/internal/modules/event/usecase"
	eventHandler "github.com/irwan/bazzar/internal/modules/event/transport"

	skuRepo "github.com/irwan/bazzar/internal/modules/sku/repository"
	skuUC "github.com/irwan/bazzar/internal/modules/sku/usecase"
	skuHandler "github.com/irwan/bazzar/internal/modules/sku/transport"

	invRepo "github.com/irwan/bazzar/internal/modules/inventory/repository"
	invUC "github.com/irwan/bazzar/internal/modules/inventory/usecase"
	invHandler "github.com/irwan/bazzar/internal/modules/inventory/transport"

	orderRepo "github.com/irwan/bazzar/internal/modules/order/repository"
	orderUsecase "github.com/irwan/bazzar/internal/modules/order/usecase"
	orderHandler "github.com/irwan/bazzar/internal/modules/order/transport"

	pickingUC "github.com/irwan/bazzar/internal/modules/picking/usecase"
	pickingHandler "github.com/irwan/bazzar/internal/modules/picking/transport"

	inboundRepo "github.com/irwan/bazzar/internal/modules/inbound/repository"
	inboundUC "github.com/irwan/bazzar/internal/modules/inbound/usecase"
	inboundHandler "github.com/irwan/bazzar/internal/modules/inbound/transport"
)

func main() {
	cfg := config.Load()
	db := config.ConnectDB(cfg)
	defer db.Close()

	// Run migrations
	runMigrations(db)

	// Auth middleware
	authMw := middleware.NewAuthMiddleware(cfg.JWTSecret)

	// ── Auth ──
	aRepo := authRepo.NewAuthRepository(db)
	aUC := authUC.NewAuthUsecase(aRepo, authMw)
	aHandler := authHandler.NewAuthHandler(aUC)

	// ── Event ──
	eRepo := eventRepo.NewEventRepository(db)
	eUC := eventUC.NewEventUsecase(eRepo)
	eHandler := eventHandler.NewEventHandler(eUC)

	// ── SKU ──
	sRepo := skuRepo.NewSKURepository(db)
	sUC := skuUC.NewSKUUsecase(sRepo)
	sHandler := skuHandler.NewSKUHandler(sUC)

	// ── Inventory ──
	iRepo := invRepo.NewInventoryRepository(db)
	iUC := invUC.NewInventoryUsecase(iRepo)
	iHandler := invHandler.NewInventoryHandler(iUC, eRepo)

	// ── Order ──
	oRepo := orderRepo.NewOrderRepository(db)
	oUC := orderUsecase.NewOrderUsecase(oRepo, sRepo, eRepo, iUC)
	oHandler := orderHandler.NewOrderHandler(oUC)

	// ── Picking & Handover ──
	pUC := pickingUC.NewPickingUsecase(oUC, sRepo, eRepo, iUC)
	pHandler := pickingHandler.NewPickingHandler(pUC)

	// ── Inbound ──
	ibRepo := inboundRepo.NewInboundRepository(db)
	ibUC := inboundUC.NewInboundUsecase(ibRepo, sRepo, eRepo, iUC)
	ibHandler := inboundHandler.NewInboundHandler(ibUC)

	// ── Seed Defaults ──
	aUC.EnsureDefaultAdmin(context.Background())
	eUC.EnsureDefaultEvent(context.Background())

	// ── Fiber App ──
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024, // 50MB
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	// ── Register Routes ──
	aHandler.RegisterRoutes(app, authMw)
	eHandler.RegisterRoutes(app, authMw)
	sHandler.RegisterRoutes(app, authMw)
	iHandler.RegisterRoutes(app, authMw)
	oHandler.RegisterRoutes(app, authMw)
	pHandler.RegisterRoutes(app, authMw)
	ibHandler.RegisterRoutes(app, authMw)

	// Health check
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "bazzar-makuku"})
	})

	// Serve frontend - detect path
	frontendPath := "./frontend"
	if _, err := os.Stat(frontendPath); os.IsNotExist(err) {
		frontendPath = "../frontend"
	}
	app.Static("/", frontendPath)

	// SPA fallback: serve index.html for non-API, non-file routes
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile(filepath.Join(frontendPath, "index.html"))
	})

	port := cfg.Port
	log.Printf("🚀 Bazzar Makuku server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func runMigrations(db *pgxpool.Pool) {
	migrationsDir := "./migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "../backend/migrations"
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			log.Println("⚠️ Migrations directory not found, skipping")
			return
		}
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Printf("⚠️ Could not read migrations: %v", err)
		return
	}

	ctx := context.Background()
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(migrationsDir, f.Name()))
		if err != nil {
			continue
		}
		_, execErr := db.Exec(ctx, string(content))
		if execErr != nil {
			errStr := execErr.Error()
			if !strings.Contains(errStr, "already exists") &&
				!strings.Contains(errStr, "duplicate key") &&
				!strings.Contains(errStr, "42P07") &&
				!strings.Contains(errStr, "42710") {
				log.Printf("⚠️ Migration %s: %v", f.Name(), execErr)
			}
		} else {
			log.Printf("✅ Applied migration: %s", f.Name())
		}
	}
}
