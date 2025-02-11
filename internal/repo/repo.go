package repo

import "github.com/jackc/pgx/v5"

type auth struct {
	db *pgx.Conn
}
type p2p struct {
	db *pgx.Conn
}
type shop struct {
	db *pgx.Conn
}
