module github.com/kalbasit/sqlc-multi-db/example

go 1.25.5

tool github.com/kalbasit/sqlc-multi-db

require (
	github.com/go-sql-driver/mysql v1.9.2
	github.com/jackc/pgx/v5 v5.7.4
	github.com/mattn/go-sqlite3 v1.14.28
)

require (
	filippo.io/edwards25519 v1.1.1 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/kalbasit/sqlc-multi-db v0.0.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/mod v0.33.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/text v0.21.0 // indirect
	golang.org/x/tools v0.42.0 // indirect
	mvdan.cc/gofumpt v0.9.2 // indirect
)

replace github.com/kalbasit/sqlc-multi-db => ../
