FROM ubuntu:latest

COPY build/main .
COPY start.sh .
COPY wait-for.sh .
COPY app.env .
COPY db/migration/. ./migration/.
RUN apt-get update && apt-get install -y netcat curl
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.14.1/migrate.linux-amd64.tar.gz | tar xvz
RUN mv migrate.linux-amd64 /usr/bin/migrate
RUN which migrate


CMD ["./build/main"]
