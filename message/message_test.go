package message

import (
	"reflect"
	"testing"

	demo "github.com/ahmadmuzakkir/demo-gokit-kafka-grpc"
)

type nullLog struct {
}

func (n nullLog) Log(a ...interface{}) error {
	return nil
}

type mockMessageRepo struct {
	OnSave func(msg demo.Message) error
	OnGet  func(limit int) ([]demo.Message, error)
}

func (m *mockMessageRepo) Save(msg demo.Message) error {
	return m.OnSave(msg)
}

func (m *mockMessageRepo) Get(limit int) ([]demo.Message, error) {
	return m.OnGet(limit)
}

func TestMessage(t *testing.T) {
	mockRepo := &mockMessageRepo{}

	var actualSave *demo.Message
	mockRepo.OnSave = func(msg demo.Message) error {
		actualSave = &msg
		return nil
	}

	expGet := []demo.Message{
		{
			ID:       "1",
			Msg:      "Hello",
			Username: "Alice",
		},
	}
	mockRepo.OnGet = func(limit int) ([]demo.Message, error) {
		return expGet, nil
	}

	svc := NewService(mockRepo, &nullLog{})

	actualGet, err := svc.Get(1)
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(expGet, actualGet) {
		t.Errorf("mismatch want: %+v got: %+v", expGet, actualGet)
		return
	}

	expSave := demo.Message{
		ID:       "2",
		Msg:      "Morning",
		Username: "Bob",
	}

	err = svc.Send(expSave)
	if err != nil {
		t.Error(err)
		return
	}

	if actualSave == nil {
		t.Error("save is never called")
		return
	}

	if !reflect.DeepEqual(expSave, *actualSave) {
		t.Errorf("mismatch want: %+v got: %+v", expGet, *actualSave)
		return
	}
}
