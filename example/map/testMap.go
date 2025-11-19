package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/telcoltar/options/pkg/option"
)

type User struct {
	UserName *option.String
	Age      *option.Int[int]

	*option.Container[User]
}

func NewUser() *User {
	u := &User{
		UserName: option.NewString("userName"),
		Age:      option.NewInt[int]("age"),
	}
	u.Container = option.NewContainer("user", u)
	return u
}

type Data struct {
	Users  *option.Slice[*User]
	Groups *option.Map[*option.Slice[*User]]

	*option.Container[Data]
}

func NewUserSlice() *option.Slice[*User] {
	return option.NewSlice("", NewUser)
}

func NewData() *Data {
	d := &Data{
		Users:  option.NewSlice("users", NewUser),
		Groups: option.NewMap("groups", NewUserSlice),
	}
	d.Container = option.NewContainer("data", d)
	return d
}

type Opt struct {
	Mode *option.String

	*option.Container[Opt]
}

func NewOpt() *Opt {
	o := &Opt{
		Mode: option.NewString("mode").Enum("dev", "prod"),
	}
	o.Container = option.NewContainer("opt", o)
	return o

}

type Cfg struct {
	Data *option.Map[*Data]
	Opt  *Opt
	Test *option.Map[int]
	*option.Container[Cfg]
}

func main() {
	cfg := &Cfg{
		Data: option.NewMap("data", NewData),
		Opt:  NewOpt(),
		Test: option.NewMap[int]("test"),
	}
	cfg.Container = option.NewContainer("cfg", cfg)

	jsonSchema := cfg.JSONSchema()
	jsonSchemaBytes, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("example/map/schema.json", jsonSchemaBytes, 0600); err != nil {
		log.Fatal(err)
	}

	data, err := os.ReadFile("example/map/testInput.json")
	if err != nil {
		log.Fatal(err)
	}

	if err := cfg.Parse(data); err != nil {
		log.Fatal(err)
	}

	for key, value := range cfg.Data.Get() {
		println(key)
		for _, u := range value.Users.Get() {
			println(u.UserName.Get())
			println(u.Age.Get())
		}
	}
	println(cfg.Opt.Mode.Get())
}
