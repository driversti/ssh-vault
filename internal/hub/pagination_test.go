package hub

import (
	"testing"
)

func TestCalcPagination(t *testing.T) {
	tests := []struct {
		name       string
		totalItems int
		page       int
		pageSize   int
		wantPage   int
		wantTotal  int
		wantPrev   bool
		wantNext   bool
		wantPrevP  int
		wantNextP  int
	}{
		{
			name:       "zero items",
			totalItems: 0, page: 1, pageSize: 20,
			wantPage: 1, wantTotal: 0,
		},
		{
			name:       "one item",
			totalItems: 1, page: 1, pageSize: 20,
			wantPage: 1, wantTotal: 1,
		},
		{
			name:       "exactly one page",
			totalItems: 20, page: 1, pageSize: 20,
			wantPage: 1, wantTotal: 1,
		},
		{
			name:       "two pages first page",
			totalItems: 21, page: 1, pageSize: 20,
			wantPage: 1, wantTotal: 2,
			wantNext: true, wantNextP: 2,
		},
		{
			name:       "two pages second page",
			totalItems: 21, page: 2, pageSize: 20,
			wantPage: 2, wantTotal: 2,
			wantPrev: true, wantPrevP: 1,
		},
		{
			name:       "three pages middle page",
			totalItems: 45, page: 2, pageSize: 20,
			wantPage: 2, wantTotal: 3,
			wantPrev: true, wantPrevP: 1,
			wantNext: true, wantNextP: 3,
		},
		{
			name:       "clamp page below 1",
			totalItems: 45, page: 0, pageSize: 20,
			wantPage: 1, wantTotal: 3,
			wantNext: true, wantNextP: 2,
		},
		{
			name:       "clamp negative page",
			totalItems: 45, page: -5, pageSize: 20,
			wantPage: 1, wantTotal: 3,
			wantNext: true, wantNextP: 2,
		},
		{
			name:       "clamp page beyond last",
			totalItems: 45, page: 999, pageSize: 20,
			wantPage: 3, wantTotal: 3,
			wantPrev: true, wantPrevP: 2,
		},
		{
			name:       "last page partial",
			totalItems: 45, page: 3, pageSize: 20,
			wantPage: 3, wantTotal: 3,
			wantPrev: true, wantPrevP: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pd := calcPagination(tt.totalItems, tt.page, tt.pageSize)

			if pd.CurrentPage != tt.wantPage {
				t.Errorf("CurrentPage = %d, want %d", pd.CurrentPage, tt.wantPage)
			}
			if pd.TotalPages != tt.wantTotal {
				t.Errorf("TotalPages = %d, want %d", pd.TotalPages, tt.wantTotal)
			}
			if pd.TotalItems != tt.totalItems {
				t.Errorf("TotalItems = %d, want %d", pd.TotalItems, tt.totalItems)
			}
			if pd.HasPrev != tt.wantPrev {
				t.Errorf("HasPrev = %v, want %v", pd.HasPrev, tt.wantPrev)
			}
			if pd.HasNext != tt.wantNext {
				t.Errorf("HasNext = %v, want %v", pd.HasNext, tt.wantNext)
			}
			if pd.PrevPage != tt.wantPrevP {
				t.Errorf("PrevPage = %d, want %d", pd.PrevPage, tt.wantPrevP)
			}
			if pd.NextPage != tt.wantNextP {
				t.Errorf("NextPage = %d, want %d", pd.NextPage, tt.wantNextP)
			}
		})
	}
}

func TestBuildPageItems(t *testing.T) {
	tests := []struct {
		name    string
		current int
		total   int
		want    []PageItem
	}{
		{
			name:    "single page returns nil",
			current: 1, total: 1,
			want: nil,
		},
		{
			name:    "three pages all shown",
			current: 2, total: 3,
			want: []PageItem{
				{Number: 1}, {Number: 2, IsActive: true}, {Number: 3},
			},
		},
		{
			name:    "seven pages all shown",
			current: 4, total: 7,
			want: []PageItem{
				{Number: 1}, {Number: 2}, {Number: 3},
				{Number: 4, IsActive: true},
				{Number: 5}, {Number: 6}, {Number: 7},
			},
		},
		{
			name:    "ten pages current at start",
			current: 1, total: 10,
			want: []PageItem{
				{Number: 1, IsActive: true}, {Number: 2},
				{IsGap: true},
				{Number: 10},
			},
		},
		{
			name:    "ten pages current at end",
			current: 10, total: 10,
			want: []PageItem{
				{Number: 1},
				{IsGap: true},
				{Number: 9}, {Number: 10, IsActive: true},
			},
		},
		{
			name:    "ten pages current in middle",
			current: 5, total: 10,
			want: []PageItem{
				{Number: 1},
				{IsGap: true},
				{Number: 4}, {Number: 5, IsActive: true}, {Number: 6},
				{IsGap: true},
				{Number: 10},
			},
		},
		{
			name:    "ten pages current at 3 no leading gap",
			current: 3, total: 10,
			want: []PageItem{
				{Number: 1}, {Number: 2}, {Number: 3, IsActive: true}, {Number: 4},
				{IsGap: true},
				{Number: 10},
			},
		},
		{
			name:    "ten pages current at 8 no trailing gap",
			current: 8, total: 10,
			want: []PageItem{
				{Number: 1},
				{IsGap: true},
				{Number: 7}, {Number: 8, IsActive: true}, {Number: 9},
				{Number: 10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPageItems(tt.current, tt.total)

			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d\ngot:  %+v\nwant: %+v", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("item[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
