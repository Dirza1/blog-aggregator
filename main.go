package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Dirza1/blog-aggregator/internal/config"
	"github.com/Dirza1/blog-aggregator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

func main() {
	readConfig := config.Read()
	currentState := state{configuration: &readConfig}
	currentCommands := commands{make(map[string]func(*state, command) error)}
	currentCommands.commandHandlers["login"] = handlerLogin
	currentCommands.commandHandlers["register"] = handlerRegister
	currentCommands.commandHandlers["reset"] = handlerReset
	currentCommands.commandHandlers["users"] = handlerUsers
	currentCommands.commandHandlers["agg"] = handlerAgg
	currentCommands.commandHandlers["addfeed"] = handleraddfeed
	db, err := sql.Open("postgres", "postgres://postgres:odin@localhost:5432/gator")
	if err != nil {
		os.Exit(1)
	}
	dbQueries := database.New(db)
	currentState.db = dbQueries
	userImput := os.Args
	if len(userImput) < 2 {
		fmt.Println("not enough arguments were given")
		os.Exit(1)
	}
	userCommand := command{
		name: userImput[1],
		args: userImput[2:],
	}
	err = currentCommands.run(&currentState, userCommand)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("expected username but none was given")
	}
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println("User not found!")
		os.Exit(1)
	}
	s.configuration.SetUser(cmd.args[0])
	fmt.Println("User set to: " + cmd.args[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("no username was given")
	}
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == nil {
		fmt.Println("user already exists")
		os.Exit(1)

	} else if errors.Is(err, sql.ErrNoRows) {
		currentTime := time.Now()
		user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{ID: uuid.New(),
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
			Name:      cmd.args[0]})
		if err != nil {
			return err
		}
		s.configuration.SetUser(user.Name)
		fmt.Println(user.Name)

	} else {
		return err
	}
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.Reset(context.Background())
	if err != nil {
		fmt.Println("error during database reset")
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Database reset")

	return nil
}

func handlerUsers(s *state, cmd command) error {
	userList, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Println("error during user retrtieval")
		fmt.Println(err)
		os.Exit(1)
	}
	for _, user := range userList {
		if user.Name == s.configuration.User {
			fmt.Println(user.Name + " (current)")
		} else {
			fmt.Println(user.Name)
		}

	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	url := "https://www.wagslane.dev/index.xml"
	data, err := fetchFeed(context.Background(), url)
	if err != nil {
		return err
	}
	fmt.Println(data)
	return nil
}

func handleraddfeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("not sufficient amounts of arguments given. Expect a name and a URL")
	}
	currentUser, err := s.db.GetUser(context.Background(), s.configuration.User)
	if err != nil {
		return err
	}
	feedName := cmd.args[0]
	feedURL := cmd.args[1]
	currentTime := time.Now()
	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{ID: uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		Name:      feedName,
		Url:       feedURL, UserID: currentUser.ID})
	if err != nil {
		return err
	}
	fmt.Printf("Feed created!\nName: %s\nID: %s\nURL: %s\nUser ID: %s\n",
		feed.Name, feed.ID, feed.Url, feed.UserID)
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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	madeStruct := RSSFeed{}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "gator")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode > 299 {
		return nil, errors.New("something went wrong with the fetiching of the data")
	}
	defer response.Body.Close()
	bites, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = xml.Unmarshal(bites, &madeStruct)
	if err != nil {
		return nil, err
	}

	madeStruct.Channel.Title = html.UnescapeString(madeStruct.Channel.Title)
	madeStruct.Channel.Description = html.UnescapeString(madeStruct.Channel.Description)
	for index := range madeStruct.Channel.Item {
		ptr := &madeStruct.Channel.Item[index]
		ptr.Description = html.UnescapeString(ptr.Description)
		ptr.Title = html.UnescapeString(ptr.Title)
	}

	return &madeStruct, nil
}

type state struct {
	configuration *config.Config
	db            *database.Queries
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandHandlers map[string]func(*state, command) error
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
