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
	currentCommands.commandHandlers["feeds"] = handlerfeeds
	currentCommands.commandHandlers["follow"] = handlerfollow
	currentCommands.commandHandlers["following"] = handlerfollowing
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
	s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{ID: uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		FeedID:    feed.ID,
		UserID:    currentUser.ID})
	return nil

}

func handlerfeeds(s *state, cmd command) error {
	listOfFeeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range listOfFeeds {
		fmt.Printf("Feed name = %s, Feed URL = %s, Assosiated user = %s", feed.Name, feed.Url, feed.Username.String)
	}
	return nil
}

func handlerfollow(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("expected URL to be passed in")
	}
	currentUser, err := s.db.GetUser(context.Background(), s.configuration.User)
	if err != nil {
		return errors.New("user id could not be found. Ensure current user is logged in coorectly")
	}
	feedId, err := s.db.GetFeedId(context.Background(), cmd.args[0])
	if err != nil {
		return errors.New("feed id could not be found. Ensure the feed URL is regestered usiing the addfeed command with the name and URL")
	}
	currentTime := time.Now()
	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{ID: uuid.New(),
		CreatedAt: currentTime,
		UpdatedAt: currentTime,
		FeedID:    feedId,
		UserID:    currentUser.ID})
	if err != nil {
		return errors.New("new relationship could not be made")
	}
	fmt.Printf("user: %s is now following feed: %s", feedFollow.FeedName, feedFollow.UserName)
	return nil
}

func handlerfollowing(s *state, cmd command) error {
	currentUser, err := s.db.GetUser(context.Background(), s.configuration.User)
	if err != nil {
		return errors.New("user id could not be found. Ensure current user is logged in coorectly")
	}
	followedFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)
	if err != nil {
		return errors.New("feeds could not be retrieved for the current user")
	}
	fmt.Printf("Current user: %s is following the following feeds: \n", currentUser.Name)
	for _, feed := range followedFeeds {
		fmt.Printf("%s\n", feed.FeedName)
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
