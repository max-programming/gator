package main

import (
	"fmt"
	"log"
	"os"

	"github.com/max-programming/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmds map[string]func(*state, command) error
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	s := state{&cfg}
	cmds := commands{
		make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)

	args := os.Args
	if len(args) < 2 {
		log.Fatal("no command provided")
	}

	cmdName := args[1]
	cmd := command{name: cmdName, args: args[2:]}

	err = cmds.run(&s, cmd)
	if err != nil {
		log.Fatal(err)
	}
}

func handlerLogin(s *state, cmd command) error {
	if cmd.name != "login" {
		return fmt.Errorf("invalid command")
	}
	if len(cmd.args) < 1 {
		return fmt.Errorf("username is required")
	}
	username := cmd.args[0]
	err := s.cfg.SetUser(username)
	if err != nil {
		return err
	}
	fmt.Printf("Logged in as %s\n", username)
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	return c.cmds[cmd.name](s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}
