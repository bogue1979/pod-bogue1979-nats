package babashka

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/jackpal/bencode-go"
)

type Message struct {
	Op   string
	Id   string
	Args string
	Var  string
}

type Namespace struct {
	Name string `bencode:"name"`
	Vars []Var  `bencode:"vars"`
}

type Var struct {
	Name string `bencode:"name"`
	Code string `bencode:"code,omitempty"`
}

type DescribeResponse struct {
	Format     string      `bencode:"format"`
	Namespaces []Namespace `bencode:"namespaces"`
}

type InvokeResponse struct {
	Id     string   `bencode:"id"`
	Value  string   `bencode:"value"` // stringified json response
	Status []string `bencode:"status"`
}

type ErrorResponse struct {
	Id        string   `bencode:"id"`
	Status    []string `bencode:"status"`
	ExMessage string   `bencode:"ex-message"`
	ExData    string   `bencode:"ex-data"`
}

func (m *Message) Arguments() (inputList []json.RawMessage, err error) {
	if err := json.Unmarshal([]byte(m.Args), &inputList); err != nil {
		return nil, err
	}
	return inputList, nil
}

func ReadMessage() (*Message, error) {
	reader := bufio.NewReader(os.Stdin)
	message := &Message{}
	if err := bencode.Unmarshal(reader, &message); err != nil {
		return nil, err
	}
	return message, nil
}

func writeResponse(response any) error {
	writer := bufio.NewWriter(os.Stdout)
	if err := bencode.Marshal(writer, response); err != nil {
		return err
	}
	writer.Flush()
	return nil
}

func WriteDescribeResponse(describeResponse *DescribeResponse) {
	writeResponse(*describeResponse)
}

func WriteNotDoneInvokeResponse(inputMessage *Message, value any) error {
	if value == nil {
		return nil
	}
	resultValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	response := InvokeResponse{
		Id:     inputMessage.Id,
		Status: []string{},
		Value:  string(resultValue),
	}
	writeResponse(response)
	return nil
}

func WriteInvokeResponse(inputMessage *Message, value any) error {
	if value == nil {
		return nil
	}
	resultValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	response := InvokeResponse{
		Id:     inputMessage.Id,
		Status: []string{"done"},
		Value:  string(resultValue),
	}
	writeResponse(response)
	return nil
}

func WriteErrorResponse(inputMessage *Message, err error) {
	errorResponse := ErrorResponse{
		Id:        inputMessage.Id,
		Status:    []string{"done", "error"},
		ExMessage: err.Error(),
	}
	writeResponse(errorResponse)
}
