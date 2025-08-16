/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package main

import (
	_ "embed"

	"carswellpress.com/cron-cowboy/cmd"
	_ "carswellpress.com/cron-cowboy/cmd/cron"
	_ "carswellpress.com/cron-cowboy/cmd/exec"
	_ "carswellpress.com/cron-cowboy/cmd/run"
)

//go:embed sql/schema.sql
var sqlSchema string

func main() {
	cmd.Execute(sqlSchema)
}
