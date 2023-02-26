package main

import (
	"fmt"
	finalsocks "github.com/lunelabs/final-socks"
	"os"
)

func main() {
	//noAuth := finalsocks.NoAuthOption()
	passAuth := finalsocks.UserPassAuth("user", "pass")

	if err := finalsocks.ListenAndServe(":8888", nil, passAuth); err != nil {
		fmt.Println(err)

		os.Exit(1)
	}
}
