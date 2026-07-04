package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"gateway/internal/api"
	"gateway/internal/audit"
	"gateway/internal/middleware"
	"gateway/internal/policy"
	"gateway/internal/telemetry"
	"gateway/internal/ziti"
)

func main() {
	// Cấu hình log ghi ra cả Console và file gateway.log để Promtail thu thập
	logFile, err := os.OpenFile("gateway.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.SetOutput(io.MultiWriter(os.Stdout, logFile))
	}

	log.Println("Starting Secure FAPI-ZTA Dark API Gateway...")

	// Khởi tạo Policy Engine (PDP)
	if err := policy.LoadPolicies(); err != nil {
		log.Fatalf("POLICY ERROR: Failed to load policies: %v", err)
	}

	// 1. Tải cấu hình từ biến môi trường
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "host=localhost port=5432 user=app_user password=app_secure_password_2026 dbname=fapi_db sslmode=disable"
	}

	jwksURL := os.Getenv("IDP_JWKS_URL")
	if jwksURL == "" {
		jwksURL = "http://localhost:8081/jwks"
	}

	zitiIdentityPath := os.Getenv("ZITI_IDENTITY_PATH")
	if zitiIdentityPath == "" {
		zitiIdentityPath = "docker/identities/gateway-dev.json"
	}

	zitiServiceName := os.Getenv("ZITI_SERVICE_NAME")
	if zitiServiceName == "" {
		zitiServiceName = "financial-ledger-service"
	}

	useZitiStr := strings.TrimSpace(os.Getenv("USE_ZITI"))
	useZiti := !strings.EqualFold(useZitiStr, "false")
	enforceZiti := useZiti || strings.EqualFold(strings.TrimSpace(os.Getenv("ENFORCE_ZITI")), "true")

	// 2. Khởi tạo Database client và kết nối
	dbClient, err := audit.NewDBClient(dbURL)
	if err != nil {
		log.Fatalf("DATABASE ERROR: Failed to connect to database: %v", err)
	}
	defer dbClient.Close()
	log.Println("Connected to PostgreSQL database successfully.")

	// 3. Khởi tạo Auth Middleware & API Handlers
	authMiddleware := middleware.NewAuthMiddleware(jwksURL, enforceZiti)
	handlers := api.NewAPIHandlers(dbClient)

	// 4. Định nghĩa Router và Middleware Chain
	mux := http.NewServeMux()

	// Giao dịch chuyển khoản: Chuyển quyền quyết định qua PDP và thực thi tại PEP
	mux.Handle("/api/transfer", authMiddleware.SecureAPI(
		authMiddleware.EnforcePolicy(http.HandlerFunc(handlers.CreateTransferHandler)),
	))

	// Truy vấn số dư
	mux.Handle("/api/balance", authMiddleware.SecureAPI(
		authMiddleware.EnforcePolicy(http.HandlerFunc(handlers.GetBalanceHandler)),
	))

	// Tra cứu nhật ký audit ledger
	mux.Handle("/api/audit-logs", authMiddleware.SecureAPI(
		authMiddleware.EnforcePolicy(http.HandlerFunc(handlers.GetAuditLogsHandler)),
	))

	// Endpoint giám sát an ninh và hiệu năng (Prometheus Exporter)
	mux.HandleFunc("/metrics", telemetry.ServeMetrics)

	// 5. Cấu hình Server lắng nghe kết nối
	// Tiêm net.Conn vào Request Context để Middleware mTLS đối chiếu danh tính
	server := &http.Server{
		Handler: mux,
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(ctx, middleware.ConnKey, c)
		},
	}

	var listener net.Listener

	if useZiti {
		log.Printf("Connecting to OpenZiti overlay network using identity: %s...", zitiIdentityPath)
		zCtx, err := ziti.NewZitiContext(zitiIdentityPath)
		if err != nil {
			log.Fatalf("ZITI CONFIG ERROR: %v", err)
		}
		defer zCtx.Close()

		log.Printf("Binding and listening on OpenZiti Dark Service: '%s'...", zitiServiceName)
		listener, err = zCtx.Listen(zitiServiceName)
		if err != nil {
			log.Fatalf("ZITI LISTEN ERROR: %v", err)
		}
		log.Println("Ziti Dark Gateway is successfully bound. Zero inbound TCP ports open on Internet!")
	} else {
		// Fallback sang chạy TCP Local phục vụ debug/testing nhanh
		localPort := os.Getenv("PORT")
		if localPort == "" {
			localPort = "8080"
		}
		addr := ":" + localPort
		log.Printf("Fallback mode: listening on standard local TCP address %s...", addr)
		listener, err = net.Listen("tcp", addr)
		if err != nil {
			log.Fatalf("TCP LISTEN ERROR: %v", err)
		}
	}

	log.Println("API Gateway is online and ready to serve requests.")
	if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
		log.Fatalf("SERVER SERVE ERROR: %v", err)
	}
}
