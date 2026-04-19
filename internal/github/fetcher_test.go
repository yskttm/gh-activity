package github

import (
	"encoding/json"
	"fmt"
	"testing"
)

// mockRESTClient は Get 呼び出しに対して順番に応答を返す
type mockRESTClient struct {
	responses []searchResponse
	index     int
}

func (m *mockRESTClient) Get(path string, resp interface{}) error {
	if m.index >= len(m.responses) {
		return fmt.Errorf("unexpected Get call #%d for path: %s", m.index+1, path)
	}
	data, err := json.Marshal(m.responses[m.index])
	if err != nil {
		return err
	}
	m.index++
	return json.Unmarshal(data, resp)
}

func item(id int64, number int, repoURL string) SearchItem {
	return SearchItem{ID: id, Number: number, RepositoryURL: repoURL}
}

func TestUniqueByID(t *testing.T) {
	items := []SearchItem{
		{ID: 1}, {ID: 2}, {ID: 1}, {ID: 3}, {ID: 2},
	}
	got := UniqueByID(items)
	if len(got) != 3 {
		t.Errorf("expected 3 unique items, got %d", len(got))
	}
	seen := map[int64]bool{}
	for _, item := range got {
		if seen[item.ID] {
			t.Errorf("duplicate ID %d found", item.ID)
		}
		seen[item.ID] = true
	}
}

func TestSortItems(t *testing.T) {
	items := []SearchItem{
		item(1, 10, "https://api.github.com/repos/org/b"),
		item(2, 5, "https://api.github.com/repos/org/a"),
		item(3, 20, "https://api.github.com/repos/org/a"),
	}
	SortItems(items)

	if items[0].RepositoryURL != "https://api.github.com/repos/org/a" || items[0].Number != 5 {
		t.Errorf("unexpected first item: repo=%s number=%d", items[0].RepositoryURL, items[0].Number)
	}
	if items[1].Number != 20 {
		t.Errorf("expected second item number=20, got %d", items[1].Number)
	}
	if items[2].RepositoryURL != "https://api.github.com/repos/org/b" {
		t.Errorf("unexpected third item repo: %s", items[2].RepositoryURL)
	}
}

func TestSearch_NoResults(t *testing.T) {
	mock := &mockRESTClient{
		responses: []searchResponse{
			{TotalCount: 0, Items: nil},
		},
	}
	client := newClientWithREST(mock)
	items, err := client.Search("author:octocat type:pr", "merged", "2024-01-01", "2024-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}

func TestSearch_SinglePage(t *testing.T) {
	mock := &mockRESTClient{
		responses: []searchResponse{
			{TotalCount: 3, Items: nil}, // checkCount
			{TotalCount: 3, Items: []SearchItem{{ID: 1}, {ID: 2}, {ID: 3}}}, // fetchPage(1)
		},
	}
	client := newClientWithREST(mock)
	items, err := client.Search("author:octocat type:pr", "merged", "2024-01-01", "2024-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
}

func TestSearch_MultiPage(t *testing.T) {
	page1 := make([]SearchItem, 100)
	for i := range page1 {
		page1[i] = SearchItem{ID: int64(i + 1)}
	}
	page2 := []SearchItem{{ID: 101}, {ID: 102}}

	mock := &mockRESTClient{
		responses: []searchResponse{
			{TotalCount: 102},         // checkCount
			{Items: page1},            // fetchPage(1) → 100件、ループ継続
			{Items: page2},            // fetchPage(2) → 2件、ループ終了
		},
	}
	client := newClientWithREST(mock)
	items, err := client.Search("author:octocat type:pr", "merged", "2024-01-01", "2024-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 102 {
		t.Errorf("expected 102 items, got %d", len(items))
	}
}

func TestSearch_SplitPeriod(t *testing.T) {
	// 2024-01-01..2024-01-03 (3日間) で1001件 → 分割される
	// 分割後: 2024-01-01..2024-01-02 と 2024-01-03..2024-01-03
	mock := &mockRESTClient{
		responses: []searchResponse{
			{TotalCount: 1001},                                    // checkCount(全期間) → 分割トリガー
			{TotalCount: 2},                                       // checkCount(前半)
			{Items: []SearchItem{{ID: 1}, {ID: 2}}},              // fetchPage(前半, 1)
			{TotalCount: 2},                                       // checkCount(後半)
			{Items: []SearchItem{{ID: 3}, {ID: 4}}},              // fetchPage(後半, 1)
		},
	}
	client := newClientWithREST(mock)
	items, err := client.Search("author:octocat type:pr", "merged", "2024-01-01", "2024-01-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 4 {
		t.Errorf("expected 4 items, got %d", len(items))
	}
}

func TestSearch_SplitCannotSplitFurther(t *testing.T) {
	// 1日しかない範囲で1001件 → Warning を出して1000件上限でフェッチ
	mock := &mockRESTClient{
		responses: []searchResponse{
			{TotalCount: 1001},                        // checkCount → 分割試みる
			{Items: make([]SearchItem, 100)},          // fetchPage(1) → 100件
			{Items: make([]SearchItem, 100)},          // fetchPage(2)
			{Items: make([]SearchItem, 100)},          // fetchPage(3)
			{Items: make([]SearchItem, 100)},          // fetchPage(4)
			{Items: make([]SearchItem, 100)},          // fetchPage(5)
			{Items: make([]SearchItem, 100)},          // fetchPage(6)
			{Items: make([]SearchItem, 100)},          // fetchPage(7)
			{Items: make([]SearchItem, 100)},          // fetchPage(8)
			{Items: make([]SearchItem, 100)},          // fetchPage(9)
			{Items: make([]SearchItem, 100)},          // fetchPage(10)
			{Items: []SearchItem{}},                   // fetchPage(11) → 0件、終了
		},
	}
	client := newClientWithREST(mock)
	// 1日の範囲 (diff_days=0 or 1)
	items, err := client.Search("author:octocat type:pr", "merged", "2024-01-01", "2024-01-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1000 {
		t.Errorf("expected 1000 items (capped), got %d", len(items))
	}
}
