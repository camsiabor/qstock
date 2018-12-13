package main

import (
	"fmt"
	"github.com/camsiabor/qcom/global"
)

func daemon(g * global.G) {
	fmt.Printf("daemon %v", g);
}