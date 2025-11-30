/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package main

import (
	"embed"

	"github.com/samcarswell/trochilus/cmd"
	_ "github.com/samcarswell/trochilus/cmd/exec"
	_ "github.com/samcarswell/trochilus/cmd/job"
	_ "github.com/samcarswell/trochilus/cmd/run"
)

//go:embed db/migrations/*.sql
var migrations embed.FS

func main() {
	cmd.Execute(migrations)
}
