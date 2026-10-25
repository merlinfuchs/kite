package engine

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
)

func TestWithTemplateParameters(t *testing.T) {
	sent := map[string]flow.FlowData{
		"verify": {
			Nodes:      []flow.FlowNode{{ID: "add_role", Data: flow.FlowNodeData{RoleTarget: "{{params.role_id}}"}}},
			Parameters: []flow.FlowParameter{{Key: "role_id", Type: flow.FlowParameterTypeRole, Value: "1"}},
		},
		// The button was removed from the template since.
		"old": {
			Parameters: []flow.FlowParameter{{Key: "role_id", Type: flow.FlowParameterTypeRole, Value: "1"}},
		},
	}
	template := map[string]flow.FlowData{
		"verify": {
			Nodes:      []flow.FlowNode{{ID: "changed_blocks"}},
			Parameters: []flow.FlowParameter{{Key: "role_id", Type: flow.FlowParameterTypeRole, Value: "2"}},
		},
	}

	res := withTemplateParameters(sent, template)
	assert.Equal(t, "2", res["verify"].ParameterValues()["role_id"])
	// The blocks stay the ones that were sent.
	assert.Equal(t, "add_role", res["verify"].Nodes[0].ID)
	assert.Equal(t, "1", res["old"].ParameterValues()["role_id"])
	// The sent message isn't changed.
	assert.Equal(t, "1", sent["verify"].ParameterValues()["role_id"])
}
