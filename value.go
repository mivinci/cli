package cli

import (
	"fmt"
	"net"
	"strconv"
)

type Value interface {
	Set(string) error
	String() string
	Type() string
}

type stringValue string

func (v *stringValue) Set(s string) error {
	*v = stringValue(s)
	return nil
}

func (v stringValue) String() string {
	return string(v)
}

func (v stringValue) Type() string {
	return "string"
}

func String(def string) Value {
	return (*stringValue)(&def)
}

type intValue int

func (v *intValue) Set(s string) error {
	i, err := strconv.Atoi(s)
	*v = intValue(i)
	return err
}

func (v intValue) String() string {
	return strconv.Itoa(int(v))
}

func (v intValue) Type() string {
	return "int"
}

func Int(def int) Value {
	return (*intValue)(&def)
}

type boolValue bool

func (v *boolValue) Set(s string) error {
	b, err := strconv.ParseBool(s)
	*v = boolValue(b)
	return err
}

func (v boolValue) String() string {
	return strconv.FormatBool(bool(v))
}

func (v boolValue) Type() string {
	return "bool"
}

func Bool(def bool) Value {
	return (*boolValue)(&def)
}

type ipValue net.IP

func (v *ipValue) Set(s string) error {
	if v == nil {
		v = &ipValue{}
	}
	if ip := net.ParseIP(s); ip != nil {
		*v = ipValue(ip)
		return nil
	}
	return fmt.Errorf("invalid IP address: %s", s)
}

func (v ipValue) String() string {
	ip := net.IP(v)
	return ip.String()
}

func (ipValue) Type() string {
	return "IP"
}

func IP(s string) Value {
	v := ipValue(net.ParseIP(s))
	return &v
}
