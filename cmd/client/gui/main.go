package main

import (
	"log"
	"os"
)

func main() {
	f, _ := os.Create("/tmp/debug.log")
	defer f.Close()
	log.SetOutput(f)
	log.Println("Before Fyne")

	//a := app.New()
	//log.Println("After Fyne app.New()")
}
