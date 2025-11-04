/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package main

import (
	"embed"
	_ "embed"

	"carswellpress.com/trochilus/cmd"
	_ "carswellpress.com/trochilus/cmd/cron"
	_ "carswellpress.com/trochilus/cmd/exec"
	_ "carswellpress.com/trochilus/cmd/run"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

func main() {
	cmd.Execute(migrations)
}
