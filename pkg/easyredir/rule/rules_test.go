package rule

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/mikelorant/easyredir-cli/pkg/easyredir"

	"github.com/MakeNowJust/heredoc"
	"github.com/gotidy/ptr"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	data string
}

func (m *mockClient) SendRequest(baseURL, path, method string, body io.Reader) (io.ReadCloser, error) {
	r := strings.NewReader(m.data)
	rc := io.NopCloser(r)
	return rc, nil
}

func TestListRules(t *testing.T) {
	type Args struct {
		options []func(*Options)
	}

	type Fields struct {
		data string
	}

	type Want struct {
		rules Rules
		err   string
	}

	tests := []struct {
		name   string
		args   Args
		fields Fields
		want   Want
	}{
		{
			name: "default",
			fields: Fields{
				data: `
					{
					  "data": [
					    {
					      "id": "abc-def",
					      "type": "rule",
					      "attributes": {
					        "forward_params": true,
					        "forward_path": true,
					        "response_type": "moved_permanently",
					        "source_urls": [
					          "abc.com",
					          "abc.com/123"
					        ],
					        "target_url": "otherdomain.com"
					      }
					    }
					  ],
					  "meta": {
					    "has_more": true
					  },
					  "links": {
					    "next": "/v1/rules?starting_after=abc-def",
					    "prev": "/v1/rules?ending_before=abc-def"
					  }
					}
				`,
			},
			want: Want{
				rules: Rules{
					Data: []Data{
						{
							ID:   "abc-def",
							Type: "rule",
							Attributes: Attributes{
								ForwardParams: ptr.Bool(true),
								ForwardPath:   ptr.Bool(true),
								ResponseType:  ptr.String("moved_permanently"),
								SourceURLs: []string{
									"abc.com",
									"abc.com/123",
								},
								TargetURL: ptr.String("otherdomain.com"),
							},
						},
					},
					Metadata: Metadata{
						HasMore: true,
					},
					Links: Links{
						Next: "/v1/rules?starting_after=abc-def",
						Prev: "/v1/rules?ending_before=abc-def",
					},
				},
			},
		}, {
			name: "with_source_filter",
			args: Args{
				options: []func(*Options){
					WithSourceFilter("https://www1.example.org"),
				},
			},
			fields: Fields{
				data: `
					{
					  "data": [
					    {
					      "id": "abc-def",
					      "type": "rule"
					    }
					  ]
					}
				`,
			},
			want: Want{
				rules: Rules{
					Data: []Data{
						{
							ID:   "abc-def",
							Type: "rule",
						},
					},
				},
			},
		}, {
			name: "with_target_filter",
			args: Args{
				options: []func(*Options){
					WithTargetFilter("https://www2.example.org"),
				},
			},
			fields: Fields{
				data: `
					{
					  "data": [
					    {
					      "id": "abc-def",
					      "type": "rule"
					    }
					  ]
					}
				`,
			},
			want: Want{
				rules: Rules{
					Data: []Data{
						{
							ID:   "abc-def",
							Type: "rule",
						},
					},
				},
			},
		}, {
			name: "with_both_source_target_filter",
			args: Args{
				options: []func(*Options){
					WithSourceFilter("https://www1.example.org"),
					WithTargetFilter("https://www2.example.org"),
				},
			},
			fields: Fields{
				data: `
					{
					  "data": [
					    {
					      "id": "abc-def",
					      "type": "rule"
					    }
					  ]
					}
				`,
			},
			want: Want{
				rules: Rules{
					Data: []Data{
						{
							ID:   "abc-def",
							Type: "rule",
						},
					},
				},
			},
		}, {
			name: "with_limit",
			args: Args{
				options: []func(*Options){
					WithLimit(1),
				},
			},
			fields: Fields{
				data: `
					{
					  "data": [
					    {
					      "id": "abc-def",
					      "type": "rule"
					    }
					  ]
					}
				`,
			},
			want: Want{
				rules: Rules{
					Data: []Data{
						{
							ID:   "abc-def",
							Type: "rule",
						},
					},
				},
			},
		}, {
			name: "error_invalid_json",
			fields: Fields{
				data: "notjson",
			},
			want: Want{
				rules: Rules{
					Data: []Data{},
				},
				err: "unable to get json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &easyredir.Easyredir{
				Client: &mockClient{
					data: tt.fields.data,
				},
				Config: &easyredir.Config{},
			}

			got, err := ListRules(e, tt.args.options...)
			if tt.want.err != "" {
				assert.NotNil(t, err)
				td.CmpContains(t, err, tt.want.err)
				return
			}
			assert.Nil(t, err)
			td.Cmp(t, got, tt.want.rules)
		})
	}
}

func TestBuildListRules(t *testing.T) {
	type Args struct {
		options *Options
	}

	type Want struct {
		pathQuery string
	}

	tests := []struct {
		name string
		args Args
		want Want
	}{
		{
			name: "no_options",
			args: Args{
				options: &Options{},
			},
			want: Want{
				pathQuery: "/rules",
			},
		}, {
			name: "starting_after",
			args: Args{
				options: &Options{
					pagination: Pagination{
						startingAfter: "96b30ce8-6331-4c18-ae49-4155c3a2136c",
					},
				},
			},
			want: Want{
				pathQuery: "/rules?starting_after=96b30ce8-6331-4c18-ae49-4155c3a2136c",
			},
		}, {
			name: "ending_before",
			args: Args{
				options: &Options{
					pagination: Pagination{
						endingBefore: "c6312a3c5514-94ea-81c4-1336-8ec03b69",
					},
				},
			},
			want: Want{
				pathQuery: "/rules?ending_before=c6312a3c5514-94ea-81c4-1336-8ec03b69",
			},
		}, {
			name: "source_filter",
			args: Args{
				options: &Options{
					sourceFilter: "http://www1.example.org",
				},
			},
			want: Want{
				pathQuery: "/rules?sq=http://www1.example.org",
			},
		}, {
			name: "target_filter",
			args: Args{
				options: &Options{
					targetFilter: "http://www2.example.org",
				},
			},
			want: Want{
				pathQuery: "/rules?tq=http://www2.example.org",
			},
		}, {
			name: "source_target_filter",
			args: Args{
				options: &Options{
					sourceFilter: "http://www1.example.org",
					targetFilter: "http://www2.example.org",
				},
			},
			want: Want{
				pathQuery: "/rules?sq=http://www1.example.org&tq=http://www2.example.org",
			},
		}, {
			name: "limit",
			args: Args{
				options: &Options{
					limit: 100,
				},
			},
			want: Want{
				pathQuery: "/rules?limit=100",
			},
		}, {
			name: "all",
			args: Args{
				options: &Options{
					sourceFilter: "http://www1.example.org",
					targetFilter: "http://www2.example.org",
					limit:        100,
					pagination: Pagination{
						startingAfter: "96b30ce8-6331-4c18-ae49-4155c3a2136c",
					},
				},
			},
			want: Want{
				pathQuery: "/rules?starting_after=96b30ce8-6331-4c18-ae49-4155c3a2136c&sq=http://www1.example.org&tq=http://www2.example.org&limit=100",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildListRules(tt.args.options)
			td.Cmp(t, got, tt.want.pathQuery)
		})
	}
}

type mockPaginatorClient struct {
	idx  int
	data string
}

func (m *mockPaginatorClient) SendRequest(baseURL, path, method string, body io.Reader) (io.ReadCloser, error) {
	data := strings.NewReader(m.data)
	docs := make(map[int]interface{})
	dec := json.NewDecoder(data)

	i := 0
	for {
		var doc map[string]interface{}
		err := dec.Decode(&doc)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("unable to decode json page %v: %w", i, err)
		}
		docs[i] = doc
		i++
	}

	b, err := json.Marshal(docs[m.idx])
	if err != nil {
		return nil, fmt.Errorf("unable to encode json page %v: %w", i, err)
	}

	r := strings.NewReader(string(b))
	rc := io.NopCloser(r)

	m.idx++

	return rc, nil
}

func TestListRulesPaginator(t *testing.T) {
	type Args struct {
		options []func(*Options)
	}

	type Fields struct {
		data string
	}

	type Want struct {
		rules Rules
		err   string
	}

	tests := []struct {
		name   string
		args   Args
		fields Fields
		want   Want
	}{
		{
			name: "one",
			fields: Fields{
				data: `
					{
					  "data": [
					    {
					      "id": "abc-def",
					      "type": "rule"
					    }
					  ]
					}
				`,
			},
			want: Want{
				rules: Rules{
					Data: []Data{
						{
							ID:   "abc-def",
							Type: "rule",
						},
					},
					Metadata: Metadata{
						HasMore: false,
					},
				},
			},
		},
		{
			name: "many",
			fields: Fields{
				data: `
					{
					  "data": [
					    {
					      "id": "abc-def",
					      "type": "rule"
					    }
					  ],
					  "meta": {
						  "has_more": true
					  },
					  "links": {
						  "next": "/v1/rules?starting_after=abc-def"
					  }
					}
					{
					  "data": [
					    {
					      "id": "bcd-efg",
					      "type": "rule"
					    }
					  ]
					}
				`,
			},
			want: Want{
				rules: Rules{
					Data: []Data{
						{
							ID:   "abc-def",
							Type: "rule",
						}, {
							ID:   "bcd-efg",
							Type: "rule",
						},
					},
				},
			},
		},
		{
			name: "none",
			want: Want{
				rules: Rules{
					Data: []Data{},
				},
			},
		},
		{
			name: "invalid_page",
			fields: Fields{
				data: `
					{
					  "data": [
					    {
					      "id": "abc-def",
					      "type": "rule"
					    }
					  ],
					  "meta": {
						  "has_more": true
					  },
					  "links": {
						  "next": "/v1/rules?starting_after=abc-def"
					  }
					}
					{ notjson }
				`,
			},
			want: Want{
				rules: Rules{
					Data: []Data{
						{
							ID:   "abc-def",
							Type: "rule",
						},
					},
				},
				err: "unable to get a rules page",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &easyredir.Easyredir{
				Client: &mockPaginatorClient{
					data: tt.fields.data,
				},
				Config: &easyredir.Config{},
			}

			got, err := ListRulesPaginator(e, tt.args.options...)
			if tt.want.err != "" {
				assert.NotNil(t, err)
				td.CmpContains(t, err, tt.want.err)
				return
			}
			assert.Nil(t, err)
			td.Cmp(t, got, tt.want.rules)
		})
	}
}

func TestRulesDataStringer(t *testing.T) {
	tests := []struct {
		name string
		give Data
		want string
	}{
		{
			name: "minimal",
			give: Data{
				ID:   "abc-def",
				Type: "rule",
			},
			want: heredoc.Doc(`
				id: abc-def
				type: rule
			`),
		},
		{
			name: "typical",
			give: Data{
				ID:   "abc-def",
				Type: "rule",
				Attributes: Attributes{
					ForwardParams: ptr.Bool(true),
					ForwardPath:   ptr.Bool(true),
					ResponseType:  ptr.String("moved_permanently"),
					SourceURLs: []string{
						"http://www1.example.org",
						"http://www2.example.org",
					},
					TargetURL: ptr.String("http://www3.example.org"),
				},
			},
			want: heredoc.Doc(`
				id: abc-def
				type: rule
				attributes:
				  forward_params: true
				  forward_path: true
				  response_type: moved_permanently
				  source_urls:
				  - http://www1.example.org
				  - http://www2.example.org
				  target_url: http://www3.example.org
			`),
		},
		{
			name: "empty",
			give: Data{},
			want: heredoc.Doc(`
				id: ""
				type: ""
			`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.give
			td.CmpString(t, got, strings.ReplaceAll(tt.want, "\t", "    "))
		})
	}
}

func TestRulesStringer(t *testing.T) {
	tests := []struct {
		name string
		give Rules
		want string
	}{
		{
			name: "minimal",
			give: Rules{
				Data: []Data{
					{
						ID:   "abc-def",
						Type: "rule",
					},
				},
			},
			want: heredoc.Doc(`
				id: abc-def
				type: rule

				Total: 1
			`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.give
			td.CmpString(t, got, strings.ReplaceAll(tt.want, "\t", "    "))
		})
	}
}
