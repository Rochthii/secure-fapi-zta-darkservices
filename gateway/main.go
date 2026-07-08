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
	"gateway/internal/pdpclient"
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

	pdpAddr := os.Getenv("PDP_ADDR")
	if pdpAddr == "" {
		pdpAddr = "localhost:50051"
	}

	pdpUseZiti := strings.EqualFold(strings.TrimSpace(os.Getenv("PDP_USE_ZITI")), "true")
	pdpZitiIdentity := os.Getenv("PDP_ZITI_IDENTITY_PATH")
	if pdpZitiIdentity == "" {
		pdpZitiIdentity = os.Getenv("ZITI_IDENTITY_PATH")
		if pdpZitiIdentity == "" {
			pdpZitiIdentity = "docker/identities/gateway-dev.json"
		}
	}
	pdpZitiServiceName := os.Getenv("PDP_ZITI_SERVICE_NAME")
	if pdpZitiServiceName == "" {
		pdpZitiServiceName = "policy-decision-service"
	}

	pdpCfg := pdpclient.Config{
		Addr:             pdpAddr,
		TLSCertFile:      os.Getenv("PDP_TLS_CERT"),
		TLSKeyFile:       os.Getenv("PDP_TLS_KEY"),
		TLSCAFile:        os.Getenv("PDP_TLS_CA"),
		FailOpen:         strings.EqualFold(strings.TrimSpace(os.Getenv("PDP_FAIL_OPEN")), "true"),
		UseZiti:          pdpUseZiti,
		ZitiIdentityFile: pdpZitiIdentity,
		ZitiServiceName:  pdpZitiServiceName,
	}
	pdpClient, err := pdpclient.New(pdpCfg)
	if err != nil {
		log.Fatalf("PDP ERROR: Failed to connect to Policy Decision Point at %s: %v", pdpAddr, err)
	}
	defer pdpClient.Close()
	if pdpUseZiti {
		log.Printf("Connected to Policy Decision Point (PDP) via OpenZiti Dark Service: %s", pdpZitiServiceName)
	} else {
		log.Printf("Connected to Policy Decision Point (PDP) via standard network: %s", pdpAddr)
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
	auditSecret := os.Getenv("AUDIT_SECRET")
	if auditSecret == "" {
		auditSecret = "test-audit-secret-key-2026" // Default fallback for local dev
	}
	dbClient, err := audit.NewDBClient(dbURL, auditSecret)
	if err != nil {
		log.Fatalf("DATABASE ERROR: Failed to connect to database: %v", err)
	}
	defer dbClient.Close()
	log.Println("Connected to PostgreSQL database successfully.")

	// 3. Khởi tạo Auth Middleware với gRPC PDP Client
	authMiddleware := middleware.NewAuthMiddleware(jwksURL, enforceZiti, pdpClient)
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
