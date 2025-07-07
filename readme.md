# Currency Watcher

Currency Watcher is a simple Go application that monitors the exchange rate between two currencies and sends a Telegram notification when a specified threshold is exceeded.

## Features

- Periodically checks the exchange rate between two configurable currencies.
- Sends a Telegram message if the rate exceeds a configurable threshold.
- Easy configuration via environment variables.

## Requirements

- Go 1.24 or later
- A valid [ExchangeRate-API](https://www.exchangerate-api.com/) key
- A Telegram bot token and chat ID

## Configuration

Set the following environment variables before running the application:

- `API_KEY`: Your ExchangeRate-API key (**required**)
- `TELEGRAM_TOKEN`: Your Telegram bot token (**required**)
- `TELEGRAM_CHAT_ID`: The chat ID to send notifications to (**required**)
- `NOTIFICATION_THRESHOLD`: The exchange rate threshold for notifications (**required**, e.g. `1.15`)
- `CHECK_INTERVAL_MINUTES`: (Optional) How often to check the rate, in minutes (default: `15`)
- `BASE_CURRENCY`: (Optional) The base currency code (default: `GBP`)
- `TARGET_CURRENCY`: (Optional) The target currency code (default: `EUR`)

## Usage

1. Clone the repository:
```bash
git clone https://github.com/yourusername/moneyChangeNotifier.git cd moneyChangeNotifier
```
2. Set the required environment variables.
3. Run the application:
```bash
go run main.go
```