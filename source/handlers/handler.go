/**
This file is used only for DI.
*/

package handlers

import (
	"database/sql"

	"github.com/go-redis/redis/v8"
)

//Handler ...
type Handler struct {
	DB    *sql.DB
	Redis *redis.Client
}

//NewHandler ...
func NewHandler(db *sql.DB, redis *redis.Client) *Handler {
	return &Handler{
		DB:    db,
		Redis: redis,
	}
}
