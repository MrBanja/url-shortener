prepare:
	@cp .env.example .env

up:
	@docker-compose up -d

up-dev:
	@docker-compose up -d --build

stop:
	@docker-compose stop