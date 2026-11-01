package bundle

import "encoding/json"

const aiGeneratedThreshold = 0.8

// AIUse classifies the recorded AI contribution to a bundle's flows. A bundle
// is generated when AI created at least 80% of its blocks, and assisted when
// AI contributed to a smaller share. Empty means no AI contribution was
// recorded.
func AIUse(content json.RawMessage) string {
	var bundle struct {
		Commands []struct {
			FlowSource flowSource `json:"flow_source"`
		} `json:"commands"`
		EventListeners []struct {
			FlowSource flowSource `json:"flow_source"`
		} `json:"event_listeners"`
		Messages []struct {
			FlowSources map[string]flowSource `json:"flow_sources"`
		} `json:"messages"`
	}
	if json.Unmarshal(content, &bundle) != nil {
		return ""
	}

	var total, generated, assisted int
	add := func(flow flowSource) {
		for _, node := range flow.Nodes {
			if node.Type == "note" {
				continue
			}
			total++
			switch node.Data["ai_contribution"] {
			case "generated":
				generated++
			case "assisted":
				assisted++
			}
		}
	}
	for _, command := range bundle.Commands {
		add(command.FlowSource)
	}
	for _, listener := range bundle.EventListeners {
		add(listener.FlowSource)
	}
	for _, message := range bundle.Messages {
		for _, flow := range message.FlowSources {
			add(flow)
		}
	}

	if total == 0 || generated+assisted == 0 {
		return ""
	}
	if float64(generated)/float64(total) >= aiGeneratedThreshold {
		return "generated"
	}
	return "assisted"
}

// AIUseDisclosure combines recorded Flow AI edits with the creator's
// disclosure of external AI use. Recorded contributions can raise, but never
// lower, the disclosed level.
func AIUseDisclosure(content json.RawMessage, declared string) string {
	detected := AIUse(content)
	if detected == "generated" || declared == "generated" {
		return "generated"
	}
	if detected == "assisted" || declared == "assisted" {
		return "assisted"
	}
	if declared == "none" || declared == "unknown" {
		return declared
	}
	return detected
}

type flowSource struct {
	Nodes []struct {
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	} `json:"nodes"`
}
