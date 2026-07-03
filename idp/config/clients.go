package config

// RegisteredClients maps client IDs to their pre-shared secrets.
var RegisteredClients = map[string]string{
	"client-alice": "alice-secure-secret-2026",
	"client-bob":   "bob-secure-secret-2026",
	"client-evil":  "evil-secure-secret-2026",
}
