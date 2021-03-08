package connections

import (
	"database/sql"
	"os"

	//postgres driver
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

//Postgres database connection
func Postgres() *sql.DB {
	dsn := os.Getenv("DATABASE_URL")

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logrus.Fatal("Could not connect to database " + err.Error())
	} else {
		logrus.Info("Connected to database")
	}

	return db
}
