docker run -d \
	--name go-auth-postgres \
	-p 5433:5432 \
	-e POSTGRES_DB=auth \
	-e POSTGRES_USER=auth \
	-e POSTGRES_PASSWORD=my-secret-password \
	-v $PWD/pgdata/:/var/lib/postgresql/data/ \
	-v $PWD/deploy/db/:/docker-entrypoint-initdb.d/ \
	postgres:14
