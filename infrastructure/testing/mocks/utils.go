package mocks

import "github.com/stretchr/testify/mock"

type (
	Mock interface {
		On(methodName string, arguments ...interface{}) *mock.Call
		AssertExpectations(t mock.TestingT) bool
	}

	Calls []struct {
		Method     string
		Args       []interface{}
		ReturnArgs []interface{}
	}
)
