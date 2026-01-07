package config

import "os"

type Config struct {
	Port      string
	MongoURI  string
	MongoDB   string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		MongoURI:  getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:   getEnv("MONGO_DB", "casos_de_codigo"),
		JWTSecret: getEnv("JWT_SECRET", "sua_chave_secreta_aqui_mude_em_producao"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
