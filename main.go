package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/max-programming/gator/internal/config"
	"github.com/max-programming/gator/internal/database"

	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
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

	db, err := sql.Open("postgres", cfg.DBUrl)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	s := state{
		db:  dbQueries,
		cfg: &cfg,
	}
	cmds := commands{
		make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handleRegister)
	cmds.register("reset", handleReset)
	cmds.register("users", handleUsers)

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
	user, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("Logged in as %s\n", username)
	return nil
}

func handleRegister(s *state, cmd command) error {
	if cmd.name != "register" {
		return fmt.Errorf("invalid command")
	}
	if len(cmd.args) < 1 {
		return fmt.Errorf("username is required")
	}
	username := cmd.args[0]
	user, err := s.db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        uuid.New(),
			Name:      username,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	)
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf(
		"Logged in as %s\nID: %s\nCreated At: %s\nUpdated At: %s",
		user.Name, user.ID, user.CreatedAt.Local().String(), user.UpdatedAt.Local().String(),
	)
	return nil
}

func handleReset(s *state, cmd command) error {
	if cmd.name != "reset" {
		return fmt.Errorf("invalid command")
	}
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("Successfully deleted all users")
	return nil
}

func handleUsers(s *state, cmd command) error {
	if cmd.name != "users" {
		return fmt.Errorf("invalid command")
	}
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, user := range users {
		username := user.Name
		if s.cfg.CurrentUserName == username {
			username += " (current)"
		}
		fmt.Printf("* %s\n", username)
	}
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	return c.cmds[cmd.name](s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}
