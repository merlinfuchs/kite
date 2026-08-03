package message

import "testing"

// A component's Discord custom ID is its flow source ID verbatim, and the
// engine routes any custom ID starting with ResumeCustomIDPrefix to a resume
// point lookup. A tenant authoring their own flow source IDs must therefore not
// be able to reach that namespace.

func componentMessage(flowSourceID string) MessageData {
	return MessageData{
		Components: []ComponentRowData{{
			Components: []ComponentData{{
				Type:         2,
				Label:        "click",
				FlowSourceID: flowSourceID,
			}},
		}},
	}
}

func TestMessageDataRejectsReservedComponentFlowSourceID(t *testing.T) {
	cases := []string{
		"resume:",
		"resume:rp-1",
		"resume:rp-1_0",
	}

	for _, flowSourceID := range cases {
		if err := componentMessage(flowSourceID).Validate(); err == nil {
			t.Errorf("flow source id %q was accepted, want rejected", flowSourceID)
		}
	}
}

func TestMessageDataRejectsReservedSelectOptionFlowSourceID(t *testing.T) {
	data := MessageData{
		Components: []ComponentRowData{{
			Components: []ComponentData{{
				Type: 3,
				Options: []ComponentSelectOptionData{{
					Label:        "option",
					FlowSourceID: "resume:rp-1_0",
				}},
			}},
		}},
	}

	if err := data.Validate(); err == nil {
		t.Error("reserved select option flow source id was accepted, want rejected")
	}
}

func TestMessageDataAcceptsOrdinaryFlowSourceIDs(t *testing.T) {
	cases := []string{
		"",
		"src-1",
		"abc123",
		"not-resume:rp-1",
		"RESUME:rp-1", // the prefix check is case sensitive, as is the router
	}

	for _, flowSourceID := range cases {
		if err := componentMessage(flowSourceID).Validate(); err != nil {
			t.Errorf("flow source id %q was rejected: %v", flowSourceID, err)
		}
	}
}

func TestMessageDataAcceptsEmptyData(t *testing.T) {
	if err := (MessageData{}).Validate(); err != nil {
		t.Errorf("empty message data was rejected: %v", err)
	}
}

// The encoder and decoder have to agree on the reserved prefix, otherwise the
// validation above guards a namespace nothing actually routes on.
func TestResumeCustomIDRoundTrip(t *testing.T) {
	component := CustomIDMessageComponentResumePoint("rp-1", 3)
	if !IsReservedCustomID(component) {
		t.Fatalf("%q is not recognised as reserved", component)
	}
	id, compID, ok := DecodeCustomIDMessageComponentResumePoint(component)
	if !ok || id != "rp-1" || compID != 3 {
		t.Errorf("decoded %q -> (%q, %d, %v), want (rp-1, 3, true)", component, id, compID, ok)
	}

	modal := CustomIDModalResumePoint("rp-2")
	if !IsReservedCustomID(modal) {
		t.Fatalf("%q is not recognised as reserved", modal)
	}
	if id, ok := DecodeCustomIDModalResumePoint(modal); !ok || id != "rp-2" {
		t.Errorf("decoded %q -> (%q, %v), want (rp-2, true)", modal, id, ok)
	}
}
