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

	"github.com/max-programming/gator/internal/config"
	"github.com/max-programming/gator/internal/database"

	"github.com/google/uuid"
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
	cmds.register("addfeed", middlewareLoggedIn(handleAddFeed))
	cmds.register("feeds", handleFeeds)
	cmds.register("follow", middlewareLoggedIn(handleFollow))
	cmds.register("following", middlewareLoggedIn(handleFollowing))

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

func handleAddFeed(s *state, cmd command, user database.User) error {
	if cmd.name != "addfeed" {
		return fmt.Errorf("invalid command")
	}
	if len(cmd.args) < 2 {
		return fmt.Errorf("feed name and url is required")
	}

	feedName := cmd.args[0]
	feedUrl := cmd.args[1]

	feed, err := s.db.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			Name:      feedName,
			Url:       feedUrl,
			UserID:    user.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf(
		"Added feed %s\nID: %s\nURL: %s\nCreated At: %s\nUpdated At: %s\n",
		feed.Name, feed.ID, feed.Url, feed.CreatedAt.Local().String(), feed.UpdatedAt.Local().String(),
	)

	_, err = s.db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			UserID:    user.ID,
			FeedID:    feed.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	)
	if err != nil {
		return err
	}

	fmt.Println("Feed Followed!")

	return nil
}

func handleFeeds(s *state, cmd command) error {
	if cmd.name != "feeds" {
		return fmt.Errorf("invalid command")
	}

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf(
			"Name: %s\nURL: %s\nUser Name: %s",
			feed.Name, feed.Url, feed.Username,
		)
	}

	return nil
}

func handleFollow(s *state, cmd command, user database.User) error {
	if cmd.name != "follow" {
		return fmt.Errorf("invalid command")
	}
	if len(cmd.args) < 1 {
		return fmt.Errorf("url is required")
	}

	feedUrl := cmd.args[0]

	feed, err := s.db.GetFeedByURL(context.Background(), feedUrl)
	if err != nil {
		return err
	}

	feed_follow, err := s.db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			UserID:    user.ID,
			FeedID:    feed.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf(
		"Feed Name: %s\nUser Name: %s\n",
		feed_follow.FeedName, feed_follow.UserName,
	)

	return nil
}

func handleFollowing(s *state, cmd command, user database.User) error {
	if cmd.name != "following" {
		return fmt.Errorf("invalid command")
	}

	feed_follows, err := s.db.GetFeedFollowsForUser(
		context.Background(),
		user.ID,
	)
	if err != nil {
		return err
	}

	for _, feed_follow := range feed_follows {
		fmt.Printf(
			"Feed Name: %s\n",
			feed_follow.FeedName,
		)
	}

	return nil
}

func middlewareLoggedIn(
	handler func(s *state, cmd command, user database.User) error,
) func(*state, command) error {
	return func(s *state, c command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}

		return handler(s, c, user)
	}
}

func (c *commands) run(s *state, cmd command) error {
	f, exists := c.cmds[cmd.name]
	if !exists {
		return fmt.Errorf("invalid command")
	}
	return f(s, cmd)
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
