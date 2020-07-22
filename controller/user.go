package controller

import (
	"goshop/service-product/service/api"
)

type User struct {
	Base
}

func (m *User) Initialise() {

}

func (m *User) GetListQuery() {
	list, err := api.NewUser(m.Context).GetListQuery("hello")
	if err != nil {
		m.SetResponse(list, err)
		return
	}

	m.SetResponse(list)
}
