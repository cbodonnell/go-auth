build:
	go build -o ./bin/auth

serve:
	./bin/auth

postgres:
	./scripts/docker_postgres.sh

