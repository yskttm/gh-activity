package github

import (
	"fmt"
	"net/url"
	"os"
	"sort"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

type SearchItem struct {
	ID            int64  `json:"id"`
	Number        int    `json:"number"`
	Title         string `json:"title"`
	HTMLURL       string `json:"html_url"`
	RepositoryURL string `json:"repository_url"`
	State         string `json:"state"`
	Draft         bool   `json:"draft"`
	Comments      int    `json:"comments"`
	CreatedAt     string `json:"created_at"`
	ClosedAt      string `json:"closed_at"`
	User          struct {
		Login string `json:"login"`
	} `json:"user"`
	Assignees []struct {
		Login string `json:"login"`
	} `json:"assignees"`
	Labels []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Milestone *struct {
		Title string `json:"title"`
	} `json:"milestone"`
	PullRequest *struct {
		MergedAt string `json:"merged_at"`
	} `json:"pull_request"`
}

type restClient interface {
	Get(path string, resp interface{}) error
}

type Client struct {
	rest          restClient
	pageSleepTime time.Duration
}

func NewClient() (*Client, error) {
	client, err := api.DefaultRESTClient()
	if err != nil {
		return nil, err
	}
	return &Client{rest: client, pageSleepTime: time.Second}, nil
}

func newClientWithREST(rest restClient) *Client {
	return &Client{rest: rest, pageSleepTime: 0}
}

type searchResponse struct {
	TotalCount int          `json:"total_count"`
	Items      []SearchItem `json:"items"`
}

func (c *Client) checkCount(query string) (int, error) {
	var resp searchResponse
	params := url.Values{"q": {query}, "per_page": {"1"}}
	if err := c.rest.Get("search/issues?"+params.Encode(), &resp); err != nil {
		return 0, err
	}
	return resp.TotalCount, nil
}

func (c *Client) fetchPage(query string, page int) ([]SearchItem, error) {
	var resp searchResponse
	params := url.Values{"q": {query}, "per_page": {"100"}, "page": {fmt.Sprintf("%d", page)}}
	if err := c.rest.Get("search/issues?"+params.Encode(), &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) fetchAllPages(query string) ([]SearchItem, error) {
	var all []SearchItem
	for page := 1; ; page++ {
		items, err := c.fetchPage(query, page)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if len(items) < 100 {
			break
		}
		time.Sleep(c.pageSleepTime)
	}
	return all, nil
}

func (c *Client) fetchWithSplit(queryBase, dateField, rangeStart, rangeEnd string) ([]SearchItem, error) {
	query := queryBase + " " + dateField + ":" + rangeStart + ".." + rangeEnd

	total, err := c.checkCount(query)
	if err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, nil
	}
	if total <= 1000 {
		return c.fetchAllPages(query)
	}

	start, err := time.Parse("2006-01-02", rangeStart)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", rangeEnd)
	if err != nil {
		return nil, err
	}

	diffDays := int(end.Sub(start).Hours() / 24)
	if diffDays <= 1 {
		fmt.Fprintf(os.Stderr, "Warning: %s..%s has %d items but cannot split further. Capped at 1000.\n", rangeStart, rangeEnd, total)
		return c.fetchAllPages(query)
	}

	mid := start.AddDate(0, 0, diffDays/2)
	midNext := mid.AddDate(0, 0, 1)

	first, err := c.fetchWithSplit(queryBase, dateField, rangeStart, mid.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	second, err := c.fetchWithSplit(queryBase, dateField, midNext.Format("2006-01-02"), rangeEnd)
	if err != nil {
		return nil, err
	}

	return UniqueByID(append(first, second...)), nil
}

func (c *Client) Search(queryBase, dateField, start, end string) ([]SearchItem, error) {
	return c.fetchWithSplit(queryBase, dateField, start, end)
}

func UniqueByID(items []SearchItem) []SearchItem {
	seen := make(map[int64]bool)
	result := make([]SearchItem, 0, len(items))
	for _, item := range items {
		if !seen[item.ID] {
			seen[item.ID] = true
			result = append(result, item)
		}
	}
	return result
}

func SortItems(items []SearchItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].RepositoryURL != items[j].RepositoryURL {
			return items[i].RepositoryURL < items[j].RepositoryURL
		}
		return items[i].Number < items[j].Number
	})
}
