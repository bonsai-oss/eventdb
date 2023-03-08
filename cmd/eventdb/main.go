package main

import (
	"os"

	"github.com/alecthomas/kingpin/v2"

	"github.com/bonsai-oss/eventdb/internal/mode"
)

func main() {
	app := kingpin.New(os.Args[0], "eventdb")

	server := &mode.Server{}
	serverCmd := app.Command("serve", "starts the eventdb server").Action(server.Run)
	serverCmd.Flag("database.name", "name of the database").Default("eventdb").Envar("DB_NAME").StringVar(&server.Database.Database)
	serverCmd.Flag("database.user", "username of the database").Default("postgres").Envar("DB_USER").StringVar(&server.Database.Username)
	serverCmd.Flag("database.password", "password of the database").Default("test123").Envar("DB_PASSWORD").StringVar(&server.Database.Password)
	serverCmd.Flag("database.host", "address of the database").Envar("DB_HOST").Required().StringVar(&server.Database.Host)
	serverCmd.Flag("web.listen-address", "address listening on").Default(":8080").Envar("WEB_LISTEN_ADDRESS").StringVar(&server.ListenAddress)

	kingpin.MustParse(app.Parse(os.Args[1:]))
}
