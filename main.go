package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hasura/go-graphql-client"
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
	Id         string
	Identifier string
	Title      string
	Url        string
}

func getLinearClient() *graphql.Client {
	config := getConfig()
	apiToken := config.Get("auth.api_token").(string)
	httpClient := &http.Client{Transport: &authenticatedTransport{apiToken: apiToken}}
	return graphql.NewClient(linearUrl, httpClient).WithDebug(true)
}

func getMyIssuesCmd() tea.Msg {
	linearClient := getLinearClient()
	myId := getMyId(linearClient)
	return getIssuesForUserId(linearClient, myId)
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

// func getTable(issues issues) string {
// 	t := table.NewWriter()
// 	// t.SetOutputMirror(os.Stdout)
// 	t.AppendHeader(table.Row{"Title", "Identifier"})
// 	for _, issue := range issues {
// 		t.AppendRows([]table.Row{{text.WrapSoft(issue.Title, 55), issue.Identifier}})
// 		t.AppendSeparator()
// 	}
// 	return t.Render()
// }

//	func getShortUrl(url string) string {
//		return url[0:9]
//	}
type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func initialModel() model {
	return model{}
	// 	choices: []string,
	//
	// 	// A map which indicates which choices are selected. We're using
	// 	// the  map like a mathematical set. The keys refer to the indexes
	// 	// of the `choices` slice, above.
	// 	selected: make(map[int]struct{}),
	// }
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return getMyIssuesCmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a key press?
	case issues:
		for _, issue := range msg {
			m.choices = append(m.choices, fmt.Sprintf("%s\t%s", issue.Title, issue.Identifier))
		}
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	s := "Select an issue\n"
	s += "Title\tIdentifier\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = "x" // selected!
		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
	// myId := getMyId(client)
	// myIssues := getIssuesForUserId(client, myId)
	// fmt.Println(myIssues)
}
