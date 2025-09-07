package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
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

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
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
	cmds.register("agg", handleAgg)

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

func handleAgg(s *state, cmd command) error {
	if cmd.name != "agg" {
		return fmt.Errorf("invalid command")
	}
	rssFeed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}

	fmt.Printf("Title: %s\n", rssFeed.Channel.Title)
	fmt.Printf("Link: %s\n", rssFeed.Channel.Link)
	fmt.Printf("Description: %s\n", rssFeed.Channel.Description)

	fmt.Println("Items")
	for _, item := range rssFeed.Channel.Item {
		fmt.Printf("Item Title: %s\n", item.Title)
		fmt.Printf("Item Link: %s\n", item.Link)
		fmt.Printf("Item Description: %s\n", item.Description)
		fmt.Printf("Item Publish Date: %s\n", item.PubDate)
	}

	return nil
}

func (c *commands) run(s *state, cmd command) error {
	return c.cmds[cmd.name](s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rssFeed RSSFeed
	err = xml.Unmarshal(bodyBytes, &rssFeed)
	if err != nil {
		return nil, err
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	for idx, item := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[idx].Title = html.UnescapeString(item.Title)
		rssFeed.Channel.Item[idx].Description = html.UnescapeString(item.Description)
	}

	return &rssFeed, nil
}
