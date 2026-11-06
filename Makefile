# Переменные
APP_NAME := grif
BUILD_DIR := bin
MAIN_PATH := ./cmd/main.go

# Цель по умолчанию
.DEFAULT_GOAL := build

# .PHONY сообщает make, что это команды, а не файлы
.PHONY: build run test clean tidy lint run/live

# Сборка бинарного файла
build:
	@echo "==> Building binary..."
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

build_ish:
	GOOS=linux GOARCH=386 go build -ldflags="-s -w" -o $(BUILD_DIR)/app_ish $(MAIN_PATH)

ios:
	gomobile bind -v -target=ios -o ./mobile/ios/grifapp/grifapp/Mobile.xcframework github.com/SlavaNovikov/grif/mobile

android:
	gomobile bind -v  -target=android -androidapi 21 -o ./mobile/android/app/libs/grifapp.aar github.com/SlavaNovikov/grif/mobile

# Запуск приложения
run: build
	./$(BUILD_DIR)/$(APP_NAME)

# Запуск с горячей перезагрузкой
run/live:
	@if command -v air > /dev/null; then \
	    air; \
	else \
	    echo "Air не установлен. Установка..."; \
	    go install github.com/air-verse/air@latest; \
	    air; \
	fi

# Запуск тестов
test:
	go test -v -race ./...

# Очистка зависимостей и форматирование
tidy:
	go mod tidy
	go fmt ./...

# Проверка кода (требуется установленный golangci-lint)
lint:
	golangci-lint run

# Удаление артефактов сборки
clean:
	rm -rf $(BUILD_DIR)
