package handler

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"secure-fapi-zta-darkservices/idp/config"
)

type JWKKey struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Kid string `json:"kid"`
}

type JWKSResponse struct {
	Keys []JWKKey `json:"keys"`
}

// JWKSHandler công khai Public Key của IdP để các dịch vụ khác verify token
func JWKSHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	pubKey := config.AppConfig.PublicKey
	
	// Trích xuất tọa độ X, Y của khóa Elliptic Curve P-256
	xBytes := pubKey.X.Bytes()
	yBytes := pubKey.Y.Bytes()

	// Độ rộng của tọa độ P-256 phải là 32 bytes (256 bits). Nếu ngắn hơn, cần pad thêm byte 0 ở đầu.
	// Để đơn giản và chính xác:
	xPad := make([]byte, 32)
	yPad := make([]byte, 32)
	copy(xPad[32-len(xBytes):], xBytes)
	copy(yPad[32-len(yBytes):], yBytes)

	jwk := JWKKey{
		Kty: "EC",
		Use: "sig",
		Alg: "ES256",
		Crv: "P-256",
		X:   base64.RawURLEncoding.EncodeToString(xPad),
		Y:   base64.RawURLEncoding.EncodeToString(yPad),
		Kid: "idp-signing-key", // Khóa tĩnh định danh khóa ký của IdP
	}

	response := JWKSResponse{
		Keys: []JWKKey{jwk},
	}

	json.NewEncoder(w).Encode(response)
}
