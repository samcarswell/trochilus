/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package main

import (
	_ "embed"

	"carswellpress.com/trochilus/cmd"
	_ "carswellpress.com/trochilus/cmd/cron"
	_ "carswellpress.com/trochilus/cmd/exec"
	_ "carswellpress.com/trochilus/cmd/run"
)

//go:embed sql/schema.sql
var sqlSchema string

func main() {
	cmd.Execute(sqlSchema)
}
