/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package main

import (
	"carswellpress.com/cron-cowboy/cmd"
	_ "carswellpress.com/cron-cowboy/cmd/run"
)

func main() {
	cmd.Execute()
}
