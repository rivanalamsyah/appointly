// Package pagination provides cursor-based and offset-based pagination helpers.
// Cursor-based pagination is preferred for large datasets and real-time feeds.
// Offset-based pagination is provided for backwards-compatible admin views.
package pagination

import (
	"encoding/base64"
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// OffsetParams holds offset-based pagination parameters.
type OffsetParams struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// Offset returns the SQL OFFSET value.
func (p OffsetParams) Offset() int {
	if p.Page <= 0 {
		return 0
	}
	return (p.Page - 1) * p.PerPage
}

// Limit returns the SQL LIMIT value.
func (p OffsetParams) Limit() int {
	if p.PerPage <= 0 {
		return DefaultPerPage
	}
	if p.PerPage > MaxPerPage {
		return MaxPerPage
	}
	return p.PerPage
}

// Normalize clamps page and per_page to valid ranges.
func (p *OffsetParams) Normalize() {
	if p.Page <= 0 {
		p.Page = DefaultPage
	}
	if p.PerPage <= 0 {
		p.PerPage = DefaultPerPage
	}
	if p.PerPage > MaxPerPage {
		p.PerPage = MaxPerPage
	}
}

// OffsetMeta holds pagination metadata for offset-based responses.
type OffsetMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// NewOffsetMeta constructs pagination metadata from params and total count.
func NewOffsetMeta(params OffsetParams, total int) OffsetMeta {
	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	if totalPages == 0 {
		totalPages = 1
	}
	return OffsetMeta{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// --- Cursor-based pagination -------------------------------------------------

// CursorParams holds cursor-based pagination parameters.
type CursorParams struct {
	Cursor string `json:"cursor"`
	Limit  int    `json:"limit"`
}

// Normalize clamps limit to valid ranges.
func (p *CursorParams) Normalize() {
	if p.Limit <= 0 {
		p.Limit = DefaultPerPage
	}
	if p.Limit > MaxPerPage {
		p.Limit = MaxPerPage
	}
}

// CursorMeta holds pagination metadata for cursor-based responses.
type CursorMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	HasNext    bool   `json:"has_next"`
	Limit      int    `json:"limit"`
}

// EncodeCursor encodes a value as a base64 cursor.
// Typically the cursor is the last item's ID or created_at timestamp.
func EncodeCursor(value string) string {
	return base64.URLEncoding.EncodeToString([]byte("cursor:" + value))
}

// DecodeCursor decodes a base64 cursor to its original value.
func DecodeCursor(cursor string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return "", fmt.Errorf("pagination: invalid cursor encoding: %w", err)
	}
	s := string(decoded)
	if !strings.HasPrefix(s, "cursor:") {
		return "", fmt.Errorf("pagination: invalid cursor format")
	}
	return strings.TrimPrefix(s, "cursor:"), nil
}

// EncodeIDCursor encodes an integer ID as a cursor.
func EncodeIDCursor(id int64) string {
	return EncodeCursor(strconv.FormatInt(id, 10))
}

// DecodeIDCursor decodes a cursor to an integer ID.
func DecodeIDCursor(cursor string) (int64, error) {
	s, err := DecodeCursor(cursor)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(s, 10, 64)
}
