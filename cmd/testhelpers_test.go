package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	gh "github.com/yskttm/gh-activity/internal/github"
)

type mockREST struct {
	responses []mockResponse
	index     int
}

type mockResponse struct {
	TotalCount int
	Items      []gh.SearchItem
}

func (m *mockREST) Get(path string, resp interface{}) error {
	if m.index >= len(m.responses) {
		return fmt.Errorf("unexpected Get call #%d for path: %s", m.index+1, path)
	}
	r := m.responses[m.index]
	m.index++
	type wire struct {
		TotalCount int             `json:"total_count"`
		Items      []gh.SearchItem `json:"items"`
	}
	data, err := json.Marshal(wire(r))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, resp)
}

func newMockFactory(responses []mockResponse) func() (*gh.Client, error) {
	return func() (*gh.Client, error) {
		return gh.NewClientWithREST(&mockREST{responses: responses}), nil
	}
}

func newTestRoot() *cobra.Command {
	root := &cobra.Command{Use: "test"}
	root.PersistentFlags().String("format", "table", "output format: table or csv")
	return root
}
