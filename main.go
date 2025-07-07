package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Config contiene tutta la configurazione letta dall'ambiente.
type Config struct {
	ApiKey                string
	TelegramToken         string
	TelegramChatID        string
	NotificationThreshold float64
	CheckInterval         time.Duration
	BaseCurrency          string
	TargetCurrency        string
}

// loadConfig carica la configurazione dalle variabili d'ambiente.
func loadConfig() (*Config, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("la variabile d'ambiente API_KEY non è impostata")
	}

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("la variabile d'ambiente TELEGRAM_TOKEN non è impostata")
	}

	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if chatID == "" {
		return nil, fmt.Errorf("la variabile d'ambiente TELEGRAM_CHAT_ID non è impostata")
	}

	thresholdStr := os.Getenv("NOTIFICATION_THRESHOLD")
	if thresholdStr == "" {
		return nil, fmt.Errorf("la variabile d'ambiente NOTIFICATION_THRESHOLD non è impostata")
	}
	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		return nil, fmt.Errorf("valore non valido per NOTIFICATION_THRESHOLD: %w", err)
	}

	// Variabili opzionali con valori di default
	intervalStr := getEnv("CHECK_INTERVAL_MINUTES", "15")
	intervalMinutes, err := strconv.Atoi(intervalStr)
	if err != nil {
		intervalMinutes = 15
	}

	return &Config{
		ApiKey:                apiKey,
		TelegramToken:         token,
		TelegramChatID:        chatID,
		NotificationThreshold: threshold,
		CheckInterval:         time.Duration(intervalMinutes) * time.Minute,
		BaseCurrency:          getEnv("BASE_CURRENCY", "GBP"),
		TargetCurrency:        getEnv("TARGET_CURRENCY", "EUR"),
	}, nil
}

// getEnv è una funzione helper per leggere una var d'ambiente o usare un default.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// ApiResponse modella la risposta JSON dall'API di cambio.
type ApiResponse struct {
	Result          string             `json:"result"`
	ConversionRates map[string]float64 `json:"conversion_rates"`
	ErrorType       string             `json:"error-type"`
}

// sendTelegramNotification invia un messaggio all'utente.
func sendTelegramNotification(cfg *Config, message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.TelegramToken)
	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id":    {cfg.TelegramChatID},
		"text":       {message},
		"parse_mode": {"Markdown"},
	})

	if err != nil {
		return fmt.Errorf("errore durante l'invio della richiesta a Telegram: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram ha restituito un errore (%d): %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// checkAndNotify esegue il controllo e restituisce 'true' se deve fermarsi.
func checkAndNotify(cfg *Config) (stop bool) {
	url := fmt.Sprintf("https://v6.exchangerate-api.com/v6/%s/latest/%s", cfg.ApiKey, cfg.BaseCurrency)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Errore durante la richiesta HTTP: %v", err)
		return false
	}
	defer resp.Body.Close()

	var apiResponse ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		log.Printf("Errore durante la decodifica del JSON: %v", err)
		return false
	}

	if apiResponse.Result != "success" {
		log.Printf("L'API di cambio ha restituito un errore: %s", apiResponse.ErrorType)
		return false
	}

	tassoAttuale, ok := apiResponse.ConversionRates[cfg.TargetCurrency]
	if !ok {
		log.Printf("Errore: la valuta '%s' non è stata trovata.", cfg.TargetCurrency)
		return false
	}

	fmt.Printf("[%s] Tasso attuale: 1 %s = %.4f %s\n", time.Now().Format("2006-01-02 15:04:05"), cfg.BaseCurrency, tassoAttuale, cfg.TargetCurrency)

	if tassoAttuale > cfg.NotificationThreshold {
		log.Println("--- SOGLIA SUPERATA! --- Invio notifica su Telegram...")

		messaggio := fmt.Sprintf(
			"🔔 *Allerta Cambio %s/%s!*\n\nIl tasso ha superato la soglia di *%.4f*.\n\nValore attuale: `1 %s = %.4f %s`",
			cfg.BaseCurrency, cfg.TargetCurrency, cfg.NotificationThreshold, cfg.BaseCurrency, tassoAttuale, cfg.TargetCurrency,
		)

		if err := sendTelegramNotification(cfg, messaggio); err != nil {
			log.Printf("ERRORE: Impossibile inviare la notifica su Telegram: %v", err)
		} else {
			log.Println("Notifica inviata con successo. Lo script terminerà.")
		}
		return true
	}
	return false
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Errore di configurazione: %v", err)
	}

	log.Printf("Avvio del controllo del cambio %s/%s con notifiche su Telegram...", cfg.BaseCurrency, cfg.TargetCurrency)
	log.Printf("Soglia impostata: > %.4f %s", cfg.NotificationThreshold, cfg.TargetCurrency)
	log.Printf("Intervallo di controllo: %v", cfg.CheckInterval)

	if stop := checkAndNotify(cfg); stop {
		return
	}

	ticker := time.NewTicker(cfg.CheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		if stop := checkAndNotify(cfg); stop {
			break
		}
	}
}
