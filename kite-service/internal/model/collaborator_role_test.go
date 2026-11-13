package model

import "testing"

func TestCollaboratorRolePermissions(t *testing.T) {
	owner, admin, developer := AppCollaboratorRoleOwner, AppCollaboratorRoleAdmin, AppCollaboratorRoleDeveloper

	tests := []struct {
		name string
		got  bool
		want bool
	}{
		{"owner manages settings", owner.CanManageSettings(), true},
		{"admin manages settings", admin.CanManageSettings(), true},
		{"developer can't manage settings", developer.CanManageSettings(), false},

		{"owner invites admins", owner.CanInviteRole(admin), true},
		{"owner invites developers", owner.CanInviteRole(developer), true},
		{"admin can't invite admins", admin.CanInviteRole(admin), false},
		{"admin invites developers", admin.CanInviteRole(developer), true},
		{"developer can't invite", developer.CanInviteRole(developer), false},
		{"nobody invites owners", owner.CanInviteRole(owner), false},

		{"owner removes admins", owner.CanRemoveRole(admin), true},
		{"admin can't remove admins", admin.CanRemoveRole(admin), false},
		{"admin removes developers", admin.CanRemoveRole(developer), true},
		{"admin can't remove the owner", admin.CanRemoveRole(owner), false},
		{"developer can't remove", developer.CanRemoveRole(developer), false},

		{"only the owner changes roles", admin.CanChangeRoles(), false},
		{"only the owner deletes the app", admin.CanDeleteApp(), false},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}
