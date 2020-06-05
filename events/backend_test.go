package events

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff(t *testing.T) {
	var b1, b2 Backends
	json.Unmarshal([]byte(`
{
	"backend-factory-octoback-octoback": {
		"loadBalancer": {"method":"wrr"},
		"servers": {
			"server-factory-octoback-octoback-1-702c4d92a31ff8a6a1fbb1b2c3bd234f": {
				"url":"http://172.16.12.2:5000",
				"weight": 1
			}
		}
	}
}
	`), &b1)
	json.Unmarshal([]byte(`
{
	"backend-factory-octoback-octoback": {
		"loadBalancer": {"method":"wrr"},
		"servers": {
			"server-factory-octoback-octoback-1-702c4d92a31ff8a6a1fbb1b2c3bd234f": {
				"url":"http://172.16.12.2:5000",
				"weight": 1
			}
		}
	},
	"backend-bearstech-beartwint-app": {
		"servers": {
			"server-bearstech-beartwint-app-1-ccf905ee93d0-49619247143d792671192b72fb66af9b": {
				"url": "http://172.16.16.2:9000",
				"weight": 1}
				},
		"loadBalancer": {"method":"wrr"}
	}
}
	`), &b2)
	b := Diff(b1, b2)
	fmt.Println(b)
	assert.Len(t, b, 1)
	_, ok := b["backend-bearstech-beartwint-app"]
	assert.True(t, ok)

	b = Diff(b2, b1)
	fmt.Println(b)
	assert.Len(t, b, 1)
	v, ok := b["backend-bearstech-beartwint-app"]
	assert.True(t, ok)
	assert.Nil(t, v)
}
