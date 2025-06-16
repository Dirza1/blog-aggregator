package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
