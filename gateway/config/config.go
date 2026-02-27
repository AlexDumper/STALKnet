package config

// Config конфигурация gateway
type Config struct {
	Port          string
	AuthService   string
	UserService   string
	ChatService   string
	TaskService   string
	JWTSecret     string
	ReadTimeout   int
	WriteTimeout  int
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		AuthService:  getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		UserService:  getEnv("USER_SERVICE_URL", "http://localhost:8082"),
		ChatService:  getEnv("CHAT_SERVICE_URL", "http://localhost:8083"),
		TaskService:  getEnv("TASK_SERVICE_URL", "http://localhost:8084"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key"),
		ReadTimeout:  getEnvInt("READ_TIMEOUT", 30),
		WriteTimeout: getEnvInt("WRITE_TIMEOUT", 30),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
