package rule

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mikelorant/easyredir-cli/pkg/easyredir/option"
)

type ClientAPI interface {
	SendRequest(baseURL, path, method string, body io.Reader) (io.ReadCloser, error)
}

type ConfigAPI interface {
	BaseURL() string
}

func ListRulesPaginator(cl ClientAPI, cfg ConfigAPI, opts ...func(*option.Options)) (r Rules, err error) {
	r = Rules{
		Data: []Data{},
	}

	rules := Rules{}
	for {
		optsWithPage := opts
		if rules.HasMore() {
			optsWithPage = append(optsWithPage, rules.NextPage())
		}

		rules, err = ListRules(cl, cfg, optsWithPage...)
		if err != nil {
			return r, fmt.Errorf("unable to get a rules page: %w", err)
		}
		r.Data = append(r.Data, rules.Data...)
		if !rules.HasMore() {
			break
		}
	}

	return r, nil
}

func ListRules(cl ClientAPI, cfg ConfigAPI, opts ...func(*option.Options)) (r Rules, err error) {
	options := &option.Options{}
	for _, o := range opts {
		o(options)
	}

	pathQuery := buildListRules(options)
	reader, err := cl.SendRequest(cfg.BaseURL(), pathQuery, http.MethodGet, nil)
	if err != nil {
		return r, fmt.Errorf("unable to send request: %w", err)
	}

	if err := decodeJSON(reader, &r); err != nil {
		return r, fmt.Errorf("unable to get json: %w", err)
	}

	return r, nil
}

func (r *Rules) NextPage() func(o *option.Options) {
	return func(o *option.Options) {
		o.Pagination.StartingAfter = strings.Split(r.Links.Next, "=")[1]
	}
}

func (r *Rules) HasMore() bool {
	return r.Metadata.HasMore
}

func buildListRules(opts *option.Options) string {
	var sb strings.Builder
	var params []string

	fmt.Fprint(&sb, "/rules")

	if opts.Pagination.StartingAfter != "" {
		params = append(params, fmt.Sprintf("starting_after=%v", opts.Pagination.StartingAfter))
	}

	if opts.Pagination.EndingBefore != "" {
		params = append(params, fmt.Sprintf("ending_before=%v", opts.Pagination.EndingBefore))
	}

	if opts.SourceFilter != "" {
		params = append(params, fmt.Sprintf("sq=%v", opts.SourceFilter))
	}

	if opts.TargetFilter != "" {
		params = append(params, fmt.Sprintf("tq=%v", opts.TargetFilter))
	}

	if opts.Limit != 0 {
		params = append(params, fmt.Sprintf("limit=%v", opts.Limit))
	}

	if len(params) != 0 {
		fmt.Fprintf(&sb, "?%v", strings.Join(params, "&"))
	}

	return sb.String()
}

func decodeJSON(r io.ReadCloser, v interface{}) error {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf("unable to json decode: %w", err)
	}
	r.Close()

	return nil
}
