package mapper_test

import (
	"fmt"
	"testing"

	mapper "github.com/uttom-akash/lwmap"
)

type Gender uint8

const (
	Male Gender = iota + 1
	Female
	Other
)

type Address struct {
	Street string
}

type User struct {
	Name string
	Age  int
	Gender
	Address *Address
}

type AddressDTO struct {
	Street string
}

type UserDTO struct {
	Name    string
	UserAge int
	Gender  uint64
	Address *AddressDTO
}

func TestMapper(t *testing.T) {

	user := User{
		Name:   "World",
		Age:    30,
		Gender: Male,
		Address: &Address{
			Street: "Hello",
		},
	}

	config := mapper.NewConfiguration()

	mapper.Add[User, UserDTO](config, map[string]func(src User) any{
		"UserAge": func(src User) any {
			return src.Age
		},
	})

	got := ""
	target, err := mapper.Map[UserDTO](user, config)
	want := "Hello, World!"

	fmt.Println(target)

	if err != nil {
		t.Errorf("got %q want %q", got, want)
	}
}
