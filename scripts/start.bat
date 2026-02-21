@echo off
REM one-click launcher for the go-ride stack (postgres, redis, kafka, api)

echo >> building and starting go-ride infrastructure...
docker-compose up --build -d

echo.
echo >> containers launched. waiting for api health...
echo >> api:    http://localhost:3000
echo >> pg:     localhost:5432
echo >> redis:  localhost:6379
echo >> kafka:  localhost:9092
echo.
echo >> run 'docker-compose logs -f api' to tail the api logs.
