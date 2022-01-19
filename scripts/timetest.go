package main

import (
	"fmt"
	"os/exec"
	"time"
)

func main() {

	start := time.Now()

	cmd, err := exec.Command("/bin/sh", "R1.sh").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	output := string(cmd)
	fmt.Println(output)

	cmd1, err := exec.Command("/bin/sh", "qr1.sh").Output()
	if err != nil {
		fmt.Printf("error %s", err)
	}
	output1 := string(cmd1)
	fmt.Println(output1)

	elapsed := time.Since(start)
	fmt.Printf("Time taken %s", elapsed)
}
