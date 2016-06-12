package user

import (
	"encoding/json"
	"errors"

	"qianno.xie/superservice/superserviced/storage"
)

type User struct {
	Name     string
	Password string
}

func Add(name, password string) error {
	bolt := storage.GetBolt("user")
	defer bolt.Close()
	user := User{
		Name:     name,
		Password: password,
	}
	return bolt.Put(name, user)
}

func Delete(name string) error {
	bolt := storage.GetBolt("user")
	defer bolt.Close()
	return bolt.Delete(name)
}

func Update(name, password string) error {
	bolt := storage.GetBolt("user")
	defer bolt.Close()
	user := User{
		Name:     name,
		Password: password,
	}
	return bolt.Update(name, user)
}

func Verify(name, password string) (bool, error) {
	bolt := storage.GetBolt("user")
	defer bolt.Close()
	kv, err := bolt.Get(name)
	if err != nil {
		return false, err
	}

	var user User
	err = json.Unmarshal(kv.Value, &user)
	if err != nil {
		return false, err
	}

	if user.Password != password {
		return false, errors.New("password wrong.")
	}

	return true, nil
}
