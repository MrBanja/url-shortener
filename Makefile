prepare:
	@cp .env.example .env

up:
	@docker-compose up -d

stop:
	@docker-compose stop