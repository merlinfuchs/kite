package model

import "testing"

func TestSharedAppMemberCanGiveRole(t *testing.T) {
	mod := SharedAppMember{BotInServer: true, ManagesServer: true, HighestRole: 5}
	if !mod.CanGiveRole(4) {
		t.Error("a role below the highest one can't be given")
	}
	if mod.CanGiveRole(5) || mod.CanGiveRole(6) {
		t.Error("a role at or above the highest one can be given")
	}

	owner := SharedAppMember{BotInServer: true, ManagesServer: true, ServerOwner: true}
	if !owner.CanGiveRole(100) {
		t.Error("the server owner can't give every role")
	}
}
