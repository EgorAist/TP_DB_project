FROM golang:1.13-stretch AS builder

WORKDIR /usr/src/app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN ls
RUN go build -v -work -o TP_DB_project cmd/main.go
RUN ls cmd/

FROM ubuntu:18.04
ENV PGVER 10
ENV PORT 5000
ENV POSTGRES_HOST localhost
ENV POSTGRES_PORT 5432
ENV POSTGRES_DB tp_forum
ENV POSTGRES_USER forum_user
ENV POSTGRES_PASSWORD 1221
EXPOSE $PORT

RUN apt-get -y update && apt-get install -y --no-install-recommends apt-utils
RUN apt-get install -y postgresql-$PGVER

USER postgres

RUN pwd
RUN ls

COPY /init.sql .

RUN service postgresql start &&\
    psql --command "CREATE USER forum_user WITH SUPERUSER PASSWORD '1221';" &&\
    createdb -O forum_user tp_forum &&\
    psql tp_forum -a -f init.sql &&\
    service postgresql stop

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

RUN ls /usr
COPY --from=builder /usr/src/app/TP_DB_project .
CMD service postgresql start && ./TP_DB_project
