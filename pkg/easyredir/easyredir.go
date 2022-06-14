package easyredir

import (
	"github.com/mikelorant/easyredir-cli/pkg/easyredir/client"
	"github.com/mikelorant/easyredir-cli/pkg/easyredir/host"
	"github.com/mikelorant/easyredir-cli/pkg/easyredir/rule"
)

type Easyredir struct {
	Client *client.Client
}

func New(apiKey, apiSecret string) *Easyredir {
	return &Easyredir{
		Client: client.New(&client.Config{
			APIKey:    apiKey,
			APISecret: apiSecret,
		}),
	}
}

func (c *Easyredir) ListRules() (r rule.Rules, err error) {
	return rule.ListRulesPaginator(c.Client)
}

func (c *Easyredir) ListHosts() (h host.Hosts, err error) {
	return host.ListHostsPaginator(c.Client)
}

func (c *Easyredir) GetHost(id string) (h host.Host, err error) {
	return host.GetHost(c.Client, id)
}
