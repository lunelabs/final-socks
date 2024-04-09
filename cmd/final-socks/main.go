package main

import (
	"fmt"
	"os"

	finalsocks "github.com/lunelabs/final-socks"
)

func main() {
	//noAuth := finalsocks.NoAuthOption()
	passAuth := finalsocks.UserPassAuth("user", "pass")

	if err := finalsocks.ListenAndServe(":8888", nil, passAuth); err != nil {
		fmt.Println(err)

		os.Exit(1)
	}
}
