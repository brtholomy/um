package next

import (
	"fmt"
	"os/exec"
)

func Last() {
	ls, err := exec.Command("ls").Output()
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("ls\n%s\n", string(ls))
}
