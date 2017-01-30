package main

import (
	"flag"
	"fmt"
	"time"

	"math/rand"

	"github.com/jzelinskie/geddit"
)

var random *rand.Rand
var upvoteCount int64

func main() {
	userPtr := flag.String("user", "", "provide username for reddit account")
	passPtr := flag.String("pass", "", "provide password for reddit account")
	subPtr := flag.String("sub", "", "provide subreddit (name only leave off /r/)")
	flag.Parse()

	if *userPtr == "" || *passPtr == "" || *subPtr == "" {
		flag.Usage()
		return
	}

	random = rand.New(rand.NewSource(time.Now().UnixNano()))

	session, err := geddit.NewLoginSession(
		*userPtr,
		*passPtr,
		"Mozilla/5.0 (Android 4.4; Mobile; rv:41.0) Gecko/41.0 Firefox/41.0",
	)
	if err != nil {
		panic("Failed to authenticate: " + err.Error())
	}

	var timestamp time.Time
	var iterations int64
	var upvotes []*geddit.Submission
	for {
		if iterations > 0 && time.Now().Sub(timestamp) < 60*time.Second {
			fmt.Println("Not many people are posting, sleeping for 30 seconds...")
			time.Sleep(30 * time.Second)
		}
		timestamp = time.Now()
		iterations += 1

		fmt.Printf("Round %d - UpVote Count %d...\n", iterations, upvoteCount)
		var err error
		if len(upvotes) == 0 {
			fmt.Println("Querying Up Votes...")
			upvotes, err = GetUpvotes(session)
			if err != nil {
				fmt.Printf("Failed to get upvoted posts: %s\n", err.Error())
				continue
			}
		}
		fmt.Println("Querying New Posts...")
		posts, err := GetAllPosts(session, *subPtr)
		if err != nil {
			fmt.Printf("Failed to get all posts: %s\n", err.Error())
			continue
		}

		UpvoteSubmissions(session, posts, upvotes)
	}
}

func GetUpvotes(s *geddit.LoginSession) ([]*geddit.Submission, error) {
	items, err := s.MyLiked(geddit.NewSubmissions, "")
	if err != nil {
		return nil, err
	}
	return items, nil
}

func GetAllPosts(s *geddit.LoginSession, sub string) ([]*geddit.Submission, error) {
	opts := geddit.ListingOptions{
		Limit: 25,
		After: "",
	}
	items, err := s.SubredditSubmissions(sub, geddit.NewSubmissions, opts)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func UpvoteSubmissions(s *geddit.LoginSession, posts, upvotes []*geddit.Submission) {
	pLen := len(posts)
	for i, p := range posts {
		found := false
		for _, u := range upvotes {
			if p.ID == u.ID {
				found = true
				break
			}
		}
		if found {
			continue
		}
		err := s.Vote(p, geddit.UpVote)
		if err != nil {
			fmt.Printf("Failed to upvote [%s] %s\n", p.ID, p.Title)
			continue
		}

		if random.Int()%4 == 3 {
			fmt.Println("Randomly checking some comments...")
			_, _ = s.Comments(p)
			delay := random.Int()%3000 + 500
			fmt.Printf("Adding %d milliseconds of entropy..\n", delay)
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
		upvotes = append(upvotes, p)
		upvoteCount += 1
		fmt.Printf("Upvoted [%s] %s [%d/%d]\n", p.ID, p.Title, i+1, pLen)
		delay := random.Int()%10000 + 500
		fmt.Printf("Adding %d milliseconds of entropy..\n", delay)
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}
