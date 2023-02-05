package main

import (
	finalsocks "bitbucket.org/lunelabs/final-socks"
	"fmt"
	"os"
)

func main() {
	noAuth := finalsocks.NoAuthOption()

	if err := finalsocks.ListenAndServe(":8888", nil, noAuth); err != nil {
		fmt.Println(err)

		os.Exit(1)
	}
}
