package easyredir

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/gotidy/ptr"
	"github.com/maxatome/go-testdeep/td"
	"github.com/stretchr/testify/assert"
)

type mockClient struct {
	filename string
}

func (m *mockClient) sendRequest(baseURL, path, method string, body io.Reader) (io.ReadCloser, error) {
	fh, err := os.Open(m.filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}
	return fh, nil
}

func TestListRules(t *testing.T) {
	type Args struct {
		options []func(*RulesOptions)
	}

	type Fields struct {
		filename string
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
				filename: "testdata/rules_list_default.json",
			},
			want: Want{
				rules: Rules{
					Data: []RuleData{
						{
							ID:   ptr.String("abc-def"),
							Type: ptr.String("rule"),
							Attributes: RuleAttributes{
								ForwardParams: ptr.Bool(true),
								ForwardPath:   ptr.Bool(true),
								ResponseType:  ptr.String("moved_permanently"),
								SourceURLs: []*string{
									ptr.String("abc.com"),
									ptr.String("abc.com/123"),
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
				options: []func(*RulesOptions){
					WithSourceFilter("https://www1.example.org"),
				},
			},
			fields: Fields{
				filename: "testdata/rules_list_minimal.json",
			},
			want: Want{
				rules: Rules{
					Data: []RuleData{
						{
							ID:   ptr.String("abc-def"),
							Type: ptr.String("rule"),
						},
					},
				},
			},
		}, {
			name: "with_target_filter",
			args: Args{
				options: []func(*RulesOptions){
					WithTargetFilter("https://www2.example.org"),
				},
			},
			fields: Fields{
				filename: "testdata/rules_list_minimal.json",
			},
			want: Want{
				rules: Rules{
					Data: []RuleData{
						{
							ID:   ptr.String("abc-def"),
							Type: ptr.String("rule"),
						},
					},
				},
			},
		}, {
			name: "with_both_source_target_filter",
			args: Args{
				options: []func(*RulesOptions){
					WithSourceFilter("https://www1.example.org"),
					WithTargetFilter("https://www2.example.org"),
				},
			},
			fields: Fields{
				filename: "testdata/rules_list_minimal.json",
			},
			want: Want{
				rules: Rules{
					Data: []RuleData{
						{
							ID:   ptr.String("abc-def"),
							Type: ptr.String("rule"),
						},
					},
				},
			},
		}, {
			name: "with_limit",
			args: Args{
				options: []func(*RulesOptions){
					WithLimit(1),
				},
			},
			fields: Fields{
				filename: "testdata/rules_list_minimal.json",
			},
			want: Want{
				rules: Rules{
					Data: []RuleData{
						{
							ID:   ptr.String("abc-def"),
							Type: ptr.String("rule"),
						},
					},
				},
			},
		}, {
			name: "error_invalid_json",
			fields: Fields{
				filename: "testdata/rules_list_invalid.json",
			},
			want: Want{
				rules: Rules{
					Data: []RuleData{},
				},
				err: "unable to get json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Easyredir{
				client: &mockClient{
					filename: tt.fields.filename,
				},
				config: &Config{},
			}

			got, err := e.ListRules(tt.args.options...)
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

func TestListRulesPathQuery(t *testing.T) {
	type Args struct {
		options *RulesOptions
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
				options: &RulesOptions{},
			},
			want: Want{
				pathQuery: "/rules",
			},
		}, {
			name: "starting_after",
			args: Args{
				options: &RulesOptions{
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
				options: &RulesOptions{
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
				options: &RulesOptions{
					sourceFilter: "http://www1.example.org",
				},
			},
			want: Want{
				pathQuery: "/rules?sq=http://www1.example.org",
			},
		}, {
			name: "target_filter",
			args: Args{
				options: &RulesOptions{
					targetFilter: "http://www2.example.org",
				},
			},
			want: Want{
				pathQuery: "/rules?tq=http://www2.example.org",
			},
		}, {
			name: "source_target_filter",
			args: Args{
				options: &RulesOptions{
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
				options: &RulesOptions{
					limit: 100,
				},
			},
			want: Want{
				pathQuery: "/rules?limit=100",
			},
		}, {
			name: "all",
			args: Args{
				options: &RulesOptions{
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
			got := listRulesPathQuery(tt.args.options)
			td.Cmp(t, got, tt.want.pathQuery)
		})
	}
}
