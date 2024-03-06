package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hasura/go-graphql-client"
	"github.com/jedib0t/go-pretty/v6/table"
	toml "github.com/pelletier/go-toml"
)

const (
	configFile = "config.toml"
	configDir  = ".config/lin"
	linearUrl  = "https://api.linear.app/graphql"
)

func getConfig() *toml.Tree {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(fmt.Errorf("could not retrieve homedir %w", err))
	}
	configContent, err := toml.LoadFile(strings.Join([]string{home, configDir, configFile}, "/"))
	if err != nil {
		fmt.Println(fmt.Errorf("could not open config file %w", err))
	}
	return configContent
}

type authenticatedTransport struct {
	apiToken string
}

func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("Authorization", t.apiToken)
	return http.DefaultTransport.RoundTrip(req)
}

func getMyId(client *graphql.Client) string {
	var query struct {
		Viewer struct {
			Id    string
			Name  string
			Email string
		}
	}

	err := client.Query(context.Background(), &query, nil)
	if err != nil {
		fmt.Println(fmt.Errorf("bad %w", err))
	}
	return query.Viewer.Id
}

type issues []struct {
	Id    string
	Title string
	Url   string
}

func getIssuesForUserId(client *graphql.Client, id string) issues {
	var query struct {
		User struct {
			Id             string
			Name           string
			AssignedIssues struct {
				Nodes issues
			}
		} `graphql:"user(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.String(id),
	}

	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		fmt.Println(fmt.Errorf("bad %w", err))
	}
	return query.User.AssignedIssues.Nodes
}

func getTable(issues issues) string {
	t := table.NewWriter()
	// t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Title", "URL"})
	for _, issue := range issues {
		t.AppendRows([]table.Row{{issue.Id, issue.Title, issue.Url}})
		t.AppendSeparator()
	}
	return t.Render()
}

func main() {
	config := getConfig()
	apiToken := config.Get("auth.api_token").(string)
	httpClient := &http.Client{Transport: &authenticatedTransport{apiToken: apiToken}}

	client := graphql.NewClient(linearUrl, httpClient).WithDebug(true)
	myId := getMyId(client)
	myIssues := getIssuesForUserId(client, myId)
	fmt.Println(getTable(myIssues))
}
