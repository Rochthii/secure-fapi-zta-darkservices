package handler

import (
	"encoding/json"
	"net/http"
	"secure-fapi-zta-darkservices/idp/config"
)

type DiscoveryMetadata struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	JwksURI                           string   `json:"jwks_uri"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IDTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	DPoPSigningAlgValuesSupported     []string `json:"dpop_signing_alg_values_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
}

// DiscoveryHandler trả về tài liệu cấu hình Discovery theo chuẩn OIDC
func DiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	baseURI := config.AppConfig.Issuer
	metadata := DiscoveryMetadata{
		Issuer:                            baseURI,
		AuthorizationEndpoint:             baseURI + "/authorize",
		TokenEndpoint:                     baseURI + "/token",
		JwksURI:                           baseURI + "/jwks",
		ResponseTypesSupported:            []string{"code"},
		SubjectTypesSupported:             []string{"public"},
		IDTokenSigningAlgValuesSupported:  []string{"ES256"},
		CodeChallengeMethodsSupported:     []string{"S256"},  // FAPI 2.0 chỉ chấp nhận S256
		DPoPSigningAlgValuesSupported:     []string{"ES256"}, // Bắt buộc cho DPoP
		TokenEndpointAuthMethodsSupported: []string{"private_key_jwt", "tls_client_auth"},
	}

	json.NewEncoder(w).Encode(metadata)
}
