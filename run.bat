@echo on
docker compose down
docker compose up --build --force-recreate
docker compose down
