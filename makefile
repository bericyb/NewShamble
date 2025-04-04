# Name of your docker-compose service and volume
SERVICE_NAME=Roshamble_postgres
VOLUME_NAME=roshamble_postgres_data  # change this to match your actual volume
COMPOSE=docker-compose

# Reset Postgres completely (container + volume)
reset-db:
	@echo "Stopping and removing PostgreSQL container and volume..."
	$(COMPOSE) down -v
	docker volume rm -f $(VOLUME_NAME) || true
	$(COMPOSE) up -d --build
	@echo "PostgreSQL has been reset!"

up: 
	@echo "Starting PostgreSQL container..."
	$(COMPOSE) up -d --build
	@echo "PostgreSQL is running!"
	air
