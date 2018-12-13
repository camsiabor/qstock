package main

import (
	"fmt"
	"github.com/camsiabor/qcom/global"
	"github.com/pkg/errors"
)

func daemon(g * global.G) {
	fmt.Printf("daemon %v", g);
	panic(errors.New("daemon process not implement yet!"));
}