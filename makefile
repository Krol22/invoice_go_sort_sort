-include .env

dev:
	ENV=development go run main.go

dev-production:
	ENV=production go run main.go

build:
	mkdir -p dist

	export FORWARDED_FROM_EMAIL
	export FORWARDED_TO_EMAIL
	export EMAIL
	export ICLOUD_PATH
	export API_KEY
	export ANTHROPIC_KEY
	export ANTHROPIC_VERSION
	export PUSHOVER_API_TOKEN
	export PUSHOVER_USER_KEY

	go build \
		-ldflags "\
		-X 'github.com/krol22/automate_firma/env.ForwardedFromEmail=${FORWARDED_FROM_EMAIL}' \
		-X 'github.com/krol22/automate_firma/env.ForwardedToEmail=${FORWARDED_TO_EMAIL}' \
		-X 'github.com/krol22/automate_firma/env.Email=${EMAIL}' \
		-X 'github.com/krol22/automate_firma/env.IcloudPath=${ICLOUD_PATH}' \
		-X 'github.com/krol22/automate_firma/env.ApiKey=${API_KEY}' \
		-X 'github.com/krol22/automate_firma/env.AnthropicKey=${ANTHROPIC_KEY}' \
		-X 'github.com/krol22/automate_firma/env.AnthropicVersion=${ANTHROPIC_VERSION}' \
		-X 'github.com/krol22/automate_firma/env.PushoverApiToken=${PUSHOVER_API_TOKEN}' \
		-X 'github.com/krol22/automate_firma/env.PushoverUserKey=${PUSHOVER_USER_KEY}'" \
	-o dist/automate_firma main.go

	go run scripts/generate_plist.go

check-env:
	@test -n "$(FORWARDED_FROM_EMAIL)" || (echo "FORWARDED_FROM_EMAIL is not set" && exit 1)
	@test -n "$(FORWARDED_TO_EMAIL)" || (echo "FORWARDED_TO_EMAIL is not set" && exit 1)
	@test -n "$(EMAIL)" || (echo "EMAIL is not set" && exit 1)
	@test -n "$(ICLOUD_PATH)" || (echo "ICLOUD_PATH is not set" && exit 1)
	@test -n "$(API_KEY)" || (echo "API_KEY is not set" && exit 1)
	@test -n "$(ANTHROPIC_KEY)" || (echo "ANTHROPIC_KEY is not set" && exit 1)
	@test -n "$(ANTHROPIC_VERSION)" || (echo "ANTHROPIC_VERSION is not set" && exit 1)
	@test -n "$(PUSHOVER_API_TOKEN)" || (echo "PUSHOVER_API_TOKEN is not set" && exit 1)
	@test -n "$(PUSHOVER_USER_KEY)" || (echo "PUSHOVER_USER_KEY is not set" && exit 1)

install: check-env build
	mkdir -p $(HOME)/.scripts/automate_firma
	cp dist/automate_firma $(HOME)/.scripts/automate_firma

	launchctl unload -w ~/Library/LaunchAgents/com.krol22.automate_firma.plist
	cp dist/com.krol22.automate_firma.plist ~/Library/LaunchAgents/com.krol22.automate_firma.plist
	launchctl load -w ~/Library/LaunchAgents/com.krol22.automate_firma.plist

