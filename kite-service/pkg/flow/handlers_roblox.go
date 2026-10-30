package flow

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionRobloxUserGet:                 executeActionRobloxUserGet,
		FlowNodeTypeActionBloxlinkUserGet:               executeActionBloxlinkUserGet,
		FlowNodeTypeActionRobloxGroupMemberGet:          executeActionRobloxGroupMemberGet,
		FlowNodeTypeActionRobloxGroupMemberRoleEdit:     executeActionRobloxGroupRoleEdit,
		FlowNodeTypeActionRobloxGroupJoinRequestRespond: executeActionRobloxJoinRequestRespond,
	})
}

func executeActionRobloxUserGet(n *CompiledFlowNode, ctx *FlowContext) error {
	id, partial, err := evalRobloxUser(ctx, n)
	if err != nil {
		return err
	}

	var user *thing.RobloxUserValue
	if id != 0 {
		user, err = ctx.Roblox.UserByID(ctx, id)
		if errors.Is(err, provider.ErrNotFound) {
			user = partial
		} else if err != nil {
			return err
		}
	}

	storeRobloxUser(n, ctx, user)
	return n.ExecuteChildren(ctx)
}

// evalRobloxUser returns the ID of the user a Roblox block targets, or 0 if
// no user has the username. For usernames it also returns the user Roblox
// found, which leaves out the description, creation date and ban status.
func evalRobloxUser(ctx *FlowContext, n *CompiledFlowNode) (int64, *thing.RobloxUserValue, error) {
	target, err := ctx.EvalTemplate(n.Data.RobloxUserTarget)
	if err != nil {
		return 0, nil, err
	}
	if target.Type == thing.TypeRobloxUser {
		return target.RobloxUser().ID, nil, nil
	}
	value := strings.TrimSpace(target.String())

	if n.Data.RobloxLookupMode != RobloxLookupTypeName {
		id, err := parseRobloxID(value, "user")
		return id, nil, err
	}

	if value == "" {
		return 0, nil, errors.New("username is empty")
	}
	users, err := ctx.Roblox.UsersByUsername(ctx, value)
	if err != nil || len(users) == 0 {
		return 0, nil, err
	}
	return users[0].ID, &users[0], nil
}

func parseRobloxID(value string, what string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("must be a Roblox %s ID, got %q", what, value)
	}
	return id, nil
}

// storeRobloxUser stores the user with their avatar as the block's result, or
// null if there's no user.
func storeRobloxUser(n *CompiledFlowNode, ctx *FlowContext, user *thing.RobloxUserValue) {
	if user == nil {
		ctx.StoreNodeResult(n, thing.Null)
		return
	}

	// The avatar is extra, flows that only use the rest shouldn't fail when
	// Roblox can't give it.
	var err error
	user.AvatarURL, err = ctx.Roblox.UserAvatarURL(ctx, user.ID)
	if err != nil {
		ctx.Log.CreateLogEntry(ctx, provider.LogLevelWarn, fmt.Sprintf("%s returned no avatar: %v", blockDefinitions[n.Type].Title, err))
	}
	ctx.StoreNodeResult(n, thing.NewRobloxUser(*user))
}

func executeActionBloxlinkUserGet(n *CompiledFlowNode, ctx *FlowContext) error {
	userID, err := evalSnowflake(ctx, n.Data.UserTarget, "user")
	if err != nil {
		return err
	}
	guildID, err := n.targetGuildID(ctx)
	if err != nil {
		return err
	}
	if !guildID.IsValid() {
		return errors.New("needs a server, Bloxlink's API keys belong to one")
	}

	var res struct {
		RobloxID json.Number `json:"robloxID"`
	}
	path := fmt.Sprintf("/v4/public/guilds/%s/discord-to-roblox/%s", guildID, userID)
	err = integrationJSONRequest(ctx, integrations["bloxlink"], http.MethodGet, path, nil, &res)
	var integrationErr *integrationError
	switch {
	case errors.As(err, &integrationErr) && integrationErr.Status == http.StatusNotFound:
		// The member hasn't verified with Bloxlink.
	case err != nil:
		return err
	}

	var user *thing.RobloxUserValue
	if res.RobloxID != "" {
		id, err := res.RobloxID.Int64()
		if err != nil {
			return fmt.Errorf("Bloxlink returned an invalid Roblox ID %q", res.RobloxID)
		}
		user, err = ctx.Roblox.UserByID(ctx, id)
		if errors.Is(err, provider.ErrNotFound) {
			// The account was deleted. Its ID still tells the member apart
			// from one who didn't verify.
			user = &thing.RobloxUserValue{ID: id}
		} else if err != nil {
			return err
		}
	}

	storeRobloxUser(n, ctx, user)
	return n.ExecuteChildren(ctx)
}

// robloxGroupRole is a role of a Roblox group, as Open Cloud returns it.
type robloxGroupRole struct {
	Path        string `json:"path"`
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Rank        int    `json:"rank"`
}

type robloxGroupMembership struct {
	Path string `json:"path"`
	// The highest-ranked role, which isn't always in Roles.
	Role  string   `json:"role"`
	Roles []string `json:"roles"`
}

// maxRobloxRolePages bounds the requests for a group's roles, 20 a page.
const maxRobloxRolePages = 10

// robloxOpenCloudRequest sends a request to Open Cloud, with the path
// relative to /cloud/v2.
func robloxOpenCloudRequest(ctx *FlowContext, method string, path string, body any, res any) error {
	return integrationJSONRequest(ctx, integrations["roblox_open_cloud"], method, "/"+path, body, res)
}

func evalRobloxGroup(ctx *FlowContext, n *CompiledFlowNode) (int64, error) {
	group, err := ctx.EvalTemplate(n.Data.RobloxGroupID)
	if err != nil {
		return 0, err
	}
	return parseRobloxID(group.String(), "group")
}

// robloxUserFilter limits a list of Open Cloud to the user.
func robloxUserFilter(userID int64) string {
	return "maxPageSize=1&filter=" + url.QueryEscape(fmt.Sprintf("user == 'users/%d'", userID))
}

// robloxGroupMember returns the membership of the user in the group, or nil
// if they aren't in it.
func robloxGroupMember(ctx *FlowContext, groupID int64, userID int64) (*robloxGroupMembership, error) {
	var res struct {
		GroupMemberships []robloxGroupMembership `json:"groupMemberships"`
	}
	err := robloxOpenCloudRequest(ctx, http.MethodGet, fmt.Sprintf("groups/%d/memberships?%s", groupID, robloxUserFilter(userID)), nil, &res)
	if err != nil || len(res.GroupMemberships) == 0 {
		return nil, err
	}

	member := res.GroupMemberships[0]
	if member.Role != "" && !slices.Contains(member.Roles, member.Role) {
		member.Roles = append(member.Roles, member.Role)
	}
	return &member, nil
}

func robloxGroupRoles(ctx *FlowContext, groupID int64) ([]robloxGroupRole, error) {
	var roles []robloxGroupRole
	pageToken := ""
	for range maxRobloxRolePages {
		path := fmt.Sprintf("groups/%d/roles?maxPageSize=20", groupID)
		if pageToken != "" {
			path += "&pageToken=" + url.QueryEscape(pageToken)
		}

		var res struct {
			GroupRoles    []robloxGroupRole `json:"groupRoles"`
			NextPageToken string            `json:"nextPageToken"`
		}
		if err := robloxOpenCloudRequest(ctx, http.MethodGet, path, nil, &res); err != nil {
			return nil, err
		}
		roles = append(roles, res.GroupRoles...)
		if res.NextPageToken == "" {
			return roles, nil
		}
		pageToken = res.NextPageToken
	}
	return nil, fmt.Errorf("the group has more than %d roles", len(roles))
}

// findRobloxGroupRole finds a role by its name, then by its rank, then by
// its ID. Ranks go up to 255, IDs are bigger.
func findRobloxGroupRole(roles []robloxGroupRole, value string) (robloxGroupRole, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return robloxGroupRole{}, errors.New("role is empty")
	}
	for _, role := range roles {
		if strings.EqualFold(role.DisplayName, value) {
			return role, nil
		}
	}
	for _, role := range roles {
		if strconv.Itoa(role.Rank) == value || role.ID == value {
			return role, nil
		}
	}
	return robloxGroupRole{}, fmt.Errorf("the group has no role named, ranked or with the ID %q", value)
}

// robloxGroupMemberResult is the result of the group member blocks: the
// member's roles, highest first.
func robloxGroupMemberResult(userID int64, rolePaths []string, roles []robloxGroupRole) thing.Thing {
	var held []robloxGroupRole
	for _, role := range roles {
		if slices.Contains(rolePaths, role.Path) {
			held = append(held, role)
		}
	}
	slices.SortStableFunc(held, func(a, b robloxGroupRole) int { return b.Rank - a.Rank })

	items := make([]thing.Thing, len(held))
	for i, role := range held {
		items[i] = thing.NewObject(map[string]thing.Thing{
			"id":   thing.NewString(role.ID),
			"name": thing.NewString(role.DisplayName),
			"rank": thing.NewInt(role.Rank),
		})
	}
	res := map[string]thing.Thing{
		"user_id": thing.NewString(strconv.FormatInt(userID, 10)),
		"roles":   thing.NewArray(items),
		"role":    thing.NewString(""),
		"rank":    thing.NewInt(0),
	}
	if len(held) > 0 {
		res["role"] = thing.NewString(held[0].DisplayName)
		res["rank"] = thing.NewInt(held[0].Rank)
	}
	return thing.NewObject(res)
}

func executeActionRobloxGroupMemberGet(n *CompiledFlowNode, ctx *FlowContext) error {
	groupID, err := evalRobloxGroup(ctx, n)
	if err != nil {
		return err
	}
	userID, _, err := evalRobloxUser(ctx, n)
	if err != nil {
		return err
	}

	var member *robloxGroupMembership
	if userID != 0 {
		member, err = robloxGroupMember(ctx, groupID, userID)
		if err != nil {
			return err
		}
	}

	res := thing.Null
	if member != nil {
		roles, err := robloxGroupRoles(ctx, groupID)
		if err != nil {
			return err
		}
		res = robloxGroupMemberResult(userID, member.Roles, roles)
	}

	ctx.StoreNodeResult(n, res)
	return n.ExecuteChildren(ctx)
}

func executeActionRobloxGroupRoleEdit(n *CompiledFlowNode, ctx *FlowContext) error {
	groupID, err := evalRobloxGroup(ctx, n)
	if err != nil {
		return err
	}
	userID, _, err := evalRobloxUser(ctx, n)
	if err != nil {
		return err
	}
	if userID == 0 {
		return errors.New("no Roblox user has that username")
	}
	roleValue, err := ctx.EvalTemplate(n.Data.RobloxGroupRole)
	if err != nil {
		return err
	}

	member, err := robloxGroupMember(ctx, groupID, userID)
	if err != nil {
		return err
	}
	if member == nil {
		return fmt.Errorf("Roblox user %d isn't in the group", userID)
	}
	roles, err := robloxGroupRoles(ctx, groupID)
	if err != nil {
		return err
	}
	role, err := findRobloxGroupRole(roles, roleValue.String())
	if err != nil {
		return err
	}

	held := member.Roles
	switch n.Data.RobloxGroupRoleMode {
	case RobloxGroupRoleModeAdd:
		if !slices.Contains(held, role.Path) {
			err = robloxOpenCloudRequest(ctx, http.MethodPost, member.Path+":assignRole", map[string]string{"role": role.Path}, nil)
			held = append(held, role.Path)
		}
	case RobloxGroupRoleModeRemove:
		if slices.Contains(held, role.Path) {
			err = robloxOpenCloudRequest(ctx, http.MethodPost, member.Path+":unassignRole", map[string]string{"role": role.Path}, nil)
			held = slices.DeleteFunc(held, func(p string) bool { return p == role.Path })
		}
	case RobloxGroupRoleModeSet, "":
		// Roblox deprecated this for assigning and unassigning roles one by
		// one, but it's the only request that replaces all of them, including
		// the hidden base role every member has.
		err = robloxOpenCloudRequest(ctx, http.MethodPatch, member.Path, map[string]string{"role": role.Path}, nil)
		held = []string{role.Path}
	default:
		return fmt.Errorf("unknown mode %q", n.Data.RobloxGroupRoleMode)
	}
	if err != nil {
		return err
	}

	ctx.StoreNodeResult(n, robloxGroupMemberResult(userID, held, roles))
	return n.ExecuteChildren(ctx)
}

// robloxJoinRequest returns the path of the user's pending request to join
// the group, or "" if they have none.
func robloxJoinRequest(ctx *FlowContext, groupID int64, userID int64) (string, error) {
	var res struct {
		GroupJoinRequests []struct {
			Path string `json:"path"`
		} `json:"groupJoinRequests"`
	}
	err := robloxOpenCloudRequest(ctx, http.MethodGet, fmt.Sprintf("groups/%d/join-requests?%s", groupID, robloxUserFilter(userID)), nil, &res)
	if err != nil || len(res.GroupJoinRequests) == 0 {
		return "", err
	}
	return res.GroupJoinRequests[0].Path, nil
}

func executeActionRobloxJoinRequestRespond(n *CompiledFlowNode, ctx *FlowContext) error {
	action := cmp.Or(n.Data.RobloxJoinRequestAction, RobloxJoinRequestActionAccept)
	if action != RobloxJoinRequestActionAccept && action != RobloxJoinRequestActionDecline {
		return fmt.Errorf("unknown action %q", action)
	}
	groupID, err := evalRobloxGroup(ctx, n)
	if err != nil {
		return err
	}
	userID, _, err := evalRobloxUser(ctx, n)
	if err != nil {
		return err
	}

	var path string
	if userID != 0 {
		path, err = robloxJoinRequest(ctx, groupID, userID)
		if err != nil {
			return err
		}
	}
	if path != "" {
		err := robloxOpenCloudRequest(ctx, http.MethodPost, path+":"+string(action), map[string]string{}, nil)
		if err != nil {
			return err
		}
	}

	// False if the user has no pending join request, like when they already
	// joined.
	ctx.StoreNodeResult(n, thing.NewBool(path != ""))
	return n.ExecuteChildren(ctx)
}
