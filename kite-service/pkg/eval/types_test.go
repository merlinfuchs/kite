package eval

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"slices"
	"testing"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

var update = flag.Bool("update", false, "update the generated files")

// The editor reads the types of what expressions can use from this file, run
// `go test ./pkg/eval -update` to update it.
const typesFile = "../../../kite-web/src/lib/flow/expressionTypes.gen.json"

// typeNames are the types the editor knows, by the names it uses.
var typeNames = map[reflect.Type]string{
	reflect.TypeOf(UserEnv{}):         "user",
	reflect.TypeOf(MemberEnv{}):       "member",
	reflect.TypeOf(RoleEnv{}):         "role",
	reflect.TypeOf(ChannelEnv{}):      "channel",
	reflect.TypeOf(GuildEnv{}):        "guild",
	reflect.TypeOf(MessageEnv{}):      "message",
	reflect.TypeOf(AttachmentEnv{}):   "attachment",
	reflect.TypeOf(EmojiEnv{}):        "emoji",
	reflect.TypeOf(HTTPResponseEnv{}): "http_response",
	reflect.TypeOf(RobloxUserEnv{}):   "roblox_user",
}

type typeField struct {
	Name string `json:"name"`
	// Type is string, int, float, bool, any or one of the named types.
	Type string `json:"type"`
	List bool   `json:"list,omitempty"`
	// Call is set for fields that are functions, like result('id').data().
	Call bool `json:"call,omitempty"`
	// NeedsServer is set for member fields that fail for a member from a
	// block result, as it doesn't know its server.
	NeedsServer bool   `json:"needs_server,omitempty"`
	Description string `json:"description,omitempty"`

	scope string
}

type expressionTypes struct {
	// Things are the types of block results by their thing type.
	Things map[thing.Type]string  `json:"things"`
	Types  map[string][]typeField `json:"types"`
}

func describeTypes(t *testing.T) expressionTypes {
	res := expressionTypes{
		Things: map[thing.Type]string{},
		Types:  map[string][]typeField{},
	}

	for _, v := range []thing.Thing{
		thing.NewDiscordMessage(discord.Message{}),
		thing.NewDiscordUser(discord.User{}),
		thing.NewDiscordMember(discord.Member{}),
		thing.NewDiscordChannel(discord.Channel{}),
		thing.NewDiscordGuild(discord.Guild{}),
		thing.NewDiscordRole(discord.Role{}),
		thing.NewRobloxUser(thing.RobloxUserValue{}),
		thing.NewHTTPResponse(thing.HTTPResponseValue{}),
	} {
		res.Things[v.Type] = typeName(t, reflect.TypeOf(NewThingEnv(v)))
	}

	for typ, name := range typeNames {
		fields := describeFields(t, typ)
		if l, ok := reflect.New(typ).Interface().(lazyFielder); ok {
			fields = append(fields, describeFields(t, reflect.TypeOf(l.lazyFields()))...)
		}
		// The member fields of a user are only filled when it's a member.
		if typ == reflect.TypeOf(UserEnv{}) {
			fields = slices.DeleteFunc(fields, func(f typeField) bool { return f.scope != "" })
		}
		res.Types[name] = fields
	}
	return res
}

func describeFields(t *testing.T, typ reflect.Type) []typeField {
	var res []typeField
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		if f.Anonymous {
			res = append(res, describeFields(t, f.Type)...)
			continue
		}
		name, ok := f.Tag.Lookup("expr")
		if !ok || !f.IsExported() {
			continue
		}

		field := typeField{
			Name:        name,
			Description: f.Tag.Get("desc"),
			NeedsServer: f.Tag.Get("scope") == "server",
			scope:       f.Tag.Get("scope"),
		}
		ft := f.Type
		if ft.Kind() == reflect.Func {
			field.Call = true
			ft = ft.Out(0)
		}
		if ft.Kind() == reflect.Slice {
			field.List = true
			ft = ft.Elem()
		}
		field.Type = typeName(t, ft)
		res = append(res, field)
	}
	return res
}

func typeName(t *testing.T, typ reflect.Type) string {
	t.Helper()

	switch typ.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int64, reflect.Uint64:
		return "int"
	case reflect.Float64:
		return "float"
	case reflect.Bool:
		return "bool"
	case reflect.Interface:
		return "any"
	case reflect.Pointer:
		typ = typ.Elem()
	}
	name, ok := typeNames[typ]
	if !ok {
		t.Fatalf("%s has no name", typ)
	}
	return name
}

func TestExpressionTypesFile(t *testing.T) {
	data, err := json.MarshalIndent(describeTypes(t), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')

	if *update {
		if err := os.WriteFile(typesFile, data, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	current, err := os.ReadFile(typesFile)
	if err != nil || string(current) != string(data) {
		t.Fatalf("%s is out of date, run `go test ./pkg/eval -update`", typesFile)
	}
}

// Fields that lazyField computes are only known to the editor from
// lazyFields, so every one of them has to be computed.
func TestLazyFieldsAreComputed(t *testing.T) {
	for typ := range typeNames {
		l, ok := reflect.New(typ).Interface().(lazyFielder)
		if !ok {
			continue
		}
		for _, field := range describeFields(t, reflect.TypeOf(l.lazyFields())) {
			v, ok, err := l.lazyField(field.Name)
			if !ok {
				t.Errorf("%s.%s isn't computed", typ.Name(), field.Name)
			} else if err == nil {
				vt := reflect.TypeOf(v)
				if field.List {
					vt = vt.Elem()
				}
				if got := typeName(t, vt); got != field.Type {
					t.Errorf("%s.%s is a %s, not a %s", typ.Name(), field.Name, got, field.Type)
				}
			}
		}
	}
}
