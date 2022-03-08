package rabbitmq

import (
	"encoding/json"
)

const DEFAULT_RETRY_LIMIT = 3

type CommandRetry struct {
	Limit   int `json:"limit"`
	Attempt int `json:"attempt"`
}

type Command struct {
	Handler string       `json:"handler"`
	Data    interface{}  `json:"data"`
	Retry   CommandRetry `json:"retry,omitempty"`
}

func NewCommand() *Command {
	command := Command{}
	command.Retry.Limit = DEFAULT_RETRY_LIMIT
	command.Retry.Attempt = 0

	return &command
}

func (command *Command) ToJson() ([]byte, error) {
	return json.Marshal(command)
}
