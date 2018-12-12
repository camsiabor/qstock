package main

import (
	"fmt"
	"reflect"
	"testing"
)

// https://colobu.com/

// TODO
// github.com/gopherjs/gopherjs

func TestOver(t *testing.T) {

	var a = make([]interface{}, 8);
	fmt.Println(len(a));
	fmt.Println(cap(a));



	var s interface{};
	var v = reflect.ValueOf(s);
	if (v.Kind() == reflect.Invalid) {

	}
	fmt.Println(v, v.Kind());
	//fmt.Println(time.Now().Unix());
}