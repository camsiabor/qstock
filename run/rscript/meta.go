package rscript

import (
	"errors"
	"github.com/camsiabor/qcom/util"
	"strings"
)

type Meta struct {
	Name   string
	Hash   string
	Binary []byte
	Script string
	Lines  []string
	Map    map[string]interface{}
}

func (o *Meta) ToMap() map[string]interface{} {
	if o.Map == nil {
		o.Map = make(map[string]interface{})
	}
	o.Map["name"] = o.Name
	o.Map["hash"] = o.Hash
	o.Map["script"] = o.Script
	return o.Map
}

func (o *Meta) FromMap(m map[string]interface{}) error {
	o.Map = m
	o.Name = util.GetStr(m, "", "name")
	if len(o.Name) == 0 {
		return errors.New(" no name")
	}
	o.Script = util.GetStr(m, "", "script")
	o.Hash = util.GetStr(m, "", "hash")
	o.Lines = strings.Split(o.Script, "\n")
	return nil
}
