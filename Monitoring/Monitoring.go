package main

import (
	"fmt"
	"log"
	"os/exec"
)

func main() {
	// Read date unix system
	date, err := exec.Command("date").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The date is %s\n", date)

	// Read mem unix system
	mem, err := exec.Command("free", "-m").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Memory usage:\n")
	fmt.Printf("%s\n", mem)

	// Read disk unix system
	disk, err := exec.Command("df", "-h").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Disk usage:\n")
	fmt.Printf("%s\n", disk)
}
