package commands

import (
	"fmt"
	"reflect"
)

type CommandRegistry struct {
	commands map[string]Command
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{commands: make(map[string]Command)}
}

func (r *CommandRegistry) Register(command Command) {
	r.commands[command.Name()] = command
}
func (r *CommandRegistry) Get(commandName string) (Command, error) {
	cmd, exists := r.commands[commandName]
	if !exists {
		return nil, fmt.Errorf("command not found: %s", commandName)
	}
	return cmd, nil
}

func AutoRegisterCommands(registry *CommandRegistry, repos map[string]interface{}) error {
	// Obtenha todos os tipos exportados no pacote de repositórios
	for _, repo := range repos {
		repoValue := reflect.ValueOf(repo)
		repoType := reflect.TypeOf(repo).Elem()

		for i := 0; i < repoType.NumField(); i++ {
			field := repoType.Field(i)
			if field.Type.Implements(reflect.TypeOf((*Command)(nil)).Elem()) {
				command := repoValue.Field(i).Interface().(Command)
				registry.Register(command)
			}
		}
	}
	return nil
}
