package comms_test

import (
	"encoding/json"
	"testing"

	"github.com/yookoala/botgame-playground/playground/comms"
)

func TestNewSimpleMessage(t *testing.T) {
	// Create a new simple message
	m := comms.NewSimpleMessage("123", "test")

	// Check the session ID
	if expected, actual := "123", m.SessionID(); expected != actual {
		t.Errorf("session ID is not correct. expected %#v, got %#v", expected, actual)
	}

	// Check the message type
	if expected, actual := "test", m.Type(); expected != actual {
		t.Errorf("type is not correct. expected %#v, got %#v", expected, actual)
	}

	// Unmarshal into a map and check the JSON values
	b, _ := json.Marshal(m)
	tm := make(map[string]interface{})
	json.Unmarshal(b, &tm)
	if expected, actual := "123", tm["sessionID"]; expected != actual {
		t.Errorf("session ID is not correct. expected %#v, got %#v", expected, actual)
	}
	if expected, actual := "test", tm["type"]; expected != actual {
		t.Errorf("type is not correct. expected %#v, got %#v", expected, actual)
	}
}

func TestNewSimpleMessage_MarshalJSON(t *testing.T) {
	// Create a new simple message
	m := comms.NewSimpleMessage("123", "test")

	// Check the session ID
	if expected, actual := "123", m.SessionID(); expected != actual {
		t.Errorf("session ID is not correct. expected %#v, got %#v", expected, actual)
	}

	// Check the message type
	if expected, actual := "test", m.Type(); expected != actual {
		t.Errorf("type is not correct. expected %#v, got %#v", expected, actual)
	}

	// Unmarshal into a map and check the JSON values
	b, _ := json.Marshal(m)
	tm := make(map[string]interface{})
	json.Unmarshal(b, &tm)
	if expected, actual := "123", tm["sessionID"]; expected != actual {
		t.Errorf("session ID is not correct. expected %#v, got %#v", expected, actual)
	}
	if expected, actual := "test", tm["type"]; expected != actual {
		t.Errorf("type is not correct. expected %#v, got %#v", expected, actual)
	}
}

func TestNewSimpleMessage_UnmarshalJSON(t *testing.T) {
	m := comms.NewSimpleMessage("", "")
	json.Unmarshal([]byte(`{"sessionID":"123", "userID": "abc", "type":"test"}`), m)

	// Check the session ID
	if expected, actual := "123", m.SessionID(); expected != actual {
		t.Errorf("session ID is not correct. expected %#v, got %#v", expected, actual)
	}

	// Check the message type
	if expected, actual := "test", m.Type(); expected != actual {
		t.Errorf("type is not correct. expected %#v, got %#v", expected, actual)
	}

}

func TestMessage_ReadDataTo(t *testing.T) {
	// Create a new simple message
	m := comms.MustMessage(comms.NewMessageFromJSONString(`{"sessionID":"123", "type":"test", "data": {"key": "some value"}}`))
	v := &struct {
		Key string `json:"key"`
	}{}
	m.ReadDataTo(v)

	// Check the data read
	if expected, actual := "some value", v.Key; expected != actual {
		t.Errorf("data is not correct. expected %#v, got %#v", expected, actual)
	}
}

func TestMessage_WriteDataFrom(t *testing.T) {
	// Create a new simple message
	m := comms.NewSimpleMessage("123", "test")
	v := struct {
		Key string `json:"key"`
	}{
		Key: "some value",
	}
	m.WriteDataFrom(v)

	j, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}

	result := make(map[string]interface{})
	err = json.Unmarshal(j, &result)
	if err != nil {
		t.Fatal(err)
	}

	// Check the data written
	if _, ok := result["data"]; !ok {
		t.Errorf("data is not found")
	}
	if _, ok := result["data"].(map[string]interface{})["key"]; !ok {
		t.Errorf("data.key is not found")
	}
	if expected, actual := "some value", result["data"].(map[string]interface{})["key"]; expected != actual {
		t.Errorf("data is not correct. expected %#v, got %#v", expected, actual)
	}
}

func TestNewGreeting(t *testing.T) {
	// Create a new greeting message
	m := comms.NewGreeting("123")

	// Check the session ID
	if expected, actual := "123", m.SessionID(); expected != actual {
		t.Errorf("session ID is not correct. expected %#v, got %#v", expected, actual)
	}

	// Check the message type
	if expected, actual := "greeting", m.Type(); expected != actual {
		t.Errorf("type is not correct. expected %#v, got %#v", expected, actual)
	}
}
