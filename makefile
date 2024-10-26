dev:
	ENV=development go run main.go

dev-production:
	ENV=production go run main.go

build:
	mkdir -p dist

	go build -o dist/automate_firma main.go
	go run scripts/generate_plist.go

install: build
	mkdir -p $(HOME)/.scripts/automate_firma
	cp dist/automate_firma $(HOME)/.scripts/automate_firma
	cp dist/com.krol22.automate_firma.plist ~/Library/LaunchAgents/com.krol22.automate_firma.plist

	launchctl unload -w ~/Library/LaunchAgents/com.krol22.automate_firma.plist
	launchctl load -w ~/Library/LaunchAgents/com.krol22.automate_firma.plist

