package mapper_test

import (
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
	Address []Address
	PrimaryAddress *Address
}

type AddressDTO struct {
	Street string
}

type UserDTO struct {
	Name    string
	UserAge int
	Gender  uint64
	Address []AddressDTO
	PrimaryAddress AddressDTO
}

func TestMapper(t *testing.T) {

	user := User{
		Name:   "World",
		Age:    30,
		Gender: Male,
		Address: []Address{
			{
				Street: "Hello",
			},
			{
				Street: "World",
			},
		},
		PrimaryAddress: &Address{
			Street: "Stuttgart",
		},
	}

	config := mapper.NewConfiguration()

	mapper.Add[User, UserDTO](config, map[string]func(src User) any{
		"UserAge": func(src User) any {
			return src.Age
		},
	})

	got := ""
	target, err := mapper.Map[[]UserDTO]([]User{user}, config)
	want := "Hello, World!"

	t.Error(target)

	if err != nil {
		t.Errorf("got %q want %q", got, want)
	}
}
