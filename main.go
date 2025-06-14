package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/Dirza1/blog-aggregator/internal/config"
	_ "github.com/lib/pq"
)

func main() {
	readConfig := config.Read()
	currentState := state{configuration: &readConfig}
	currentCommands := commands{make(map[string]func(*state, command) error)}
	currentCommands.commandHandlers["login"] = handlerLogin
	userImput := os.Args
	if len(userImput) < 2 {
		fmt.Println("not enough arguments were given")
		os.Exit(1)
	}
	userCommand := command{
		name: userImput[1],
		args: userImput[2:],
	}
	err := currentCommands.run(&currentState, userCommand)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("expected username but none was given")
	}
	s.configuration.SetUser(cmd.args[0])
	fmt.Println("User set to: " + cmd.args[0])
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	function, exists := c.commandHandlers[cmd.name]
	if exists {
		err := function(s, cmd)
		if err != nil {
			return err
		}
	} else {
		return errors.New("command does not exist")
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandHandlers[name] = f
}

type state struct {
	configuration *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandHandlers map[string]func(*state, command) error
}
