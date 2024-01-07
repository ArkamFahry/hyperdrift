package models

import (
	"reflect"
	"testing"
	"time"

	"github.com/ArkamFahry/hyperdrift-storage/server/packages/apperr"
)

func TestNewEvent(t *testing.T) {
	t.Run("CreateEventInstance", func(t *testing.T) {
		name := "sampleEvent"
		data := map[string]any{
			"key1": "value1",
			"key2": 123,
		}
		createdAt := time.Now()

		expectedEvent := &CreateEvent{
			Id:        "event_someUniqueId",
			Name:      name,
			Data:      data,
			Producer:  "heyperdrift.storage",
			CreatedAt: createdAt,
		}

		newEvent := NewEvent(name, data)

		if newEvent.Name != expectedEvent.Name ||
			!reflect.DeepEqual(newEvent.Data, expectedEvent.Data) ||
			newEvent.Producer != expectedEvent.Producer {
			t.Errorf("Generated event does not match the expected event.\nExpected: %+v\nGot: %+v", expectedEvent, newEvent)
		}
	})
}

func TestEventValidation(t *testing.T) {
	t.Run("MissingId", func(t *testing.T) {
		eventMissingID := &CreateEvent{
			Name:      "sampleEvent",
			Data:      map[string]interface{}{"key1": "value1"},
			Producer:  "heyperdrift.storage",
			CreatedAt: time.Now(),
		}

		err := eventMissingID.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "id"
		expectedErrorMsg := "id is required"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("MissingName", func(t *testing.T) {
		eventMissingName := &CreateEvent{
			Id:        "event123",
			Data:      map[string]interface{}{"key1": "value1"},
			Producer:  "heyperdrift.storage",
			CreatedAt: time.Now(),
		}

		err := eventMissingName.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "name"
		expectedErrorMsg := "name is required"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("NameContainsWhiteSpace", func(t *testing.T) {
		eventNameWithWhiteSpace := &CreateEvent{
			Id:        "event123",
			Name:      "sample Event",
			Data:      map[string]interface{}{"key1": "value1"},
			Producer:  "heyperdrift.storage",
			CreatedAt: time.Now(),
		}

		err := eventNameWithWhiteSpace.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "name"
		expectedErrorMsg := "name should not contain any white spaces or tabs"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("MissingData", func(t *testing.T) {
		eventMissingData := &CreateEvent{
			Id:        "event123",
			Name:      "sampleEvent",
			Producer:  "heyperdrift.storage",
			CreatedAt: time.Now(),
		}

		err := eventMissingData.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "data"
		expectedErrorMsg := "data is required"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("InvalidProducer", func(t *testing.T) {
		eventInvalidProducer := &CreateEvent{
			Id:        "event123",
			Name:      "sampleEvent",
			Data:      map[string]interface{}{"key1": "value1"},
			Producer:  "hey per drift",
			CreatedAt: time.Now(),
		}

		err := eventInvalidProducer.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "producer"
		expectedErrorMsg := "producer should not contain any white spaces or tabs"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("MissingCreatedAt", func(t *testing.T) {
		eventMissingCreatedAt := &CreateEvent{
			Id:       "event123",
			Name:     "sampleEvent",
			Data:     map[string]interface{}{"key1": "value1"},
			Producer: "heyperdrift.storage",
		}

		err := eventMissingCreatedAt.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "created_at"
		expectedErrorMsg := "created_at is required"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("ValidEvent", func(t *testing.T) {
		validEvent := &CreateEvent{
			Id:        "event123",
			Name:      "sampleEvent",
			Data:      map[string]interface{}{"key1": "value1"},
			Producer:  "heyperdrift.storage",
			CreatedAt: time.Now(),
		}

		err := validEvent.Validate()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
