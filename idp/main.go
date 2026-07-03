package main

import (
	"fmt"
	"log"
	"net/http"
	"secure-fapi-zta-darkservices/idp/config"
	"secure-fapi-zta-darkservices/idp/handler"
)

func main() {
	// 1. Tải hoặc sinh cặp khóa ký ES256 cho IdP
	if err := config.LoadOrGenerateKeys("."); err != nil {
		log.Fatalf("LỖI: Không thể cấu hình khóa ký: %v", err)
	}

	// 2. Thiết lập router và handlers
	mux := http.NewServeMux()

	// OIDC Metadata Discovery
	mux.HandleFunc("/.well-known/openid-configuration", handler.DiscoveryHandler)

	// JWKS (JSON Web Key Set) - Cả hai đường dẫn để tương thích tối đa
	mux.HandleFunc("/.well-known/jwks.json", handler.JWKSHandler)
	mux.HandleFunc("/jwks", handler.JWKSHandler)

	// Luồng cấp mã Authorization Code
	mux.HandleFunc("/authorize", handler.AuthorizeHandler)

	// Luồng đổi token (DPoP + PKCE validation)
	mux.HandleFunc("/token", handler.TokenHandler)

	// Cấu hình cổng chạy mặc định là 8081
	port := "8081"
	fmt.Printf("Identity Provider (IdP) đang chạy trên cổng :%s...\n", port)
	fmt.Printf("- Discovery URI: http://localhost:%s/.well-known/openid-configuration\n", port)
	fmt.Printf("- JWKS URI:      http://localhost:%s/jwks\n", port)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("LỖI: Server crash: %v", err)
	}
}
