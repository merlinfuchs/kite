package js

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

// A module is the code of a whole app. Its top level registers handlers of
// commands, events and components, with discord.command, discord.on and
// discord.component, and declares the stored variables it uses with
// variables.define. Calls fail at the top level.
//
// A module loads once for each version of it, and every run of a handler
// starts from a snapshot of the instance after that, so globals start with
// what the top level set and don't carry over from one run to the next.
// Modules can't rely on what globals do between runs, as that will change.

// Manifest is what a module registers when it loads.
type Manifest struct {
	Commands   []ManifestCommand  `json:"commands"`
	Events     []string           `json:"events"`
	Components []string           `json:"components"`
	Variables  []ManifestVariable `json:"variables"`
}

type ManifestCommand struct {
	// Name is the name of the command, or of a subcommand after the names of
	// its command and group, like "level show".
	Name string `json:"name"`
	// Data is the object the module passed with it, like the description
	// and options in Discord's format.
	Data json.RawMessage `json:"data"`
}

type ManifestVariable struct {
	Name   string `json:"name"`
	Scoped bool   `json:"scoped"`
}

type ModuleOptions struct {
	Code string
	// Log is like in Options.
	Log func(level LogLevel, message string)
	// Host is like in Options. Loading doesn't use it, as calls fail while
	// a module loads. Handlers of interactions also call
	// interaction.respond with Discord's interaction response, like
	// { type: 4, data: { content: "Hi" } }, and interaction.follow_up with
	// a message, which the host answers for the interaction of the run.
	Host func(name string) HostFunc

	// Replace ModuleTimeout, ModuleLoadTimeout and MaxModuleWallTime in
	// tests, like in Options.
	timeout  time.Duration
	wallTime time.Duration
}

type ModuleCallKind string

const (
	ModuleCallCommand   ModuleCallKind = "command"
	ModuleCallEvent     ModuleCallKind = "event"
	ModuleCallComponent ModuleCallKind = "component"
)

// ModuleCall is what a module handles: a command, event or component, by
// the name the module registered it with.
type ModuleCall struct {
	Kind ModuleCallKind `json:"kind"`
	Name string         `json:"name"`
	// Interaction is Discord's interaction, for commands and components.
	Interaction json.RawMessage `json:"interaction,omitempty"`
	// Event is the data of Discord's gateway event, for events.
	Event json.RawMessage `json:"event,omitempty"`
	// Data is what the module saved with the component, for components.
	Data json.RawMessage `json:"data,omitempty"`
}

// A module loads in a function, which the prelude calls. The function starts
// on its first line so line numbers match. Handlers run with __kiteDispatch.
const (
	moduleHead   = "__kiteModule(function () {"
	moduleTail   = "\n})"
	moduleCaller = "__kiteDispatch"
)

// LoadModule runs the top level of a module and returns what it registered.
// The Manifest is shared with other callers and must not be changed.
func LoadModule(ctx context.Context, opts ModuleOptions) (*Manifest, error) {
	m, err := modules.load(ctx, opts)
	if err != nil {
		return nil, err
	}
	return m.manifest, nil
}

// RunModule runs the handler of call. It loads the module first if it isn't
// loaded yet, see moduleCache.
func RunModule(ctx context.Context, opts ModuleOptions, call ModuleCall) error {
	m, err := modules.load(ctx, opts)
	if err != nil {
		return err
	}
	in, err := json.Marshal(call)
	if err != nil {
		return fmt.Errorf("failed to encode input: %w", err)
	}
	if len(in) > MaxInputLength {
		return &Error{Message: fmt.Sprintf("the input is more than %d MB", MaxInputLength>>20)}
	}
	_, _, err = runScript(ctx, script{
		limits:   moduleLimits,
		eval:     moduleCaller,
		source:   m.source,
		input:    in,
		snapshot: m.snapshot,
		log:      opts.Log,
		host:     opts.Host,
		timeout:  opts.timeout,
		wallTime: opts.wallTime,
	})
	return err
}

// moduleCacheSize caps the memory of the snapshots moduleCache keeps. A
// small module's is about 1.4 MB, one of MaxModuleLength about 7 MB.
const moduleCacheSize = 512 << 20

var modules = &moduleCache{max: moduleCacheSize}

// moduleCache keeps the snapshots of the instances of loaded modules by the
// hash of their code, so each version of a module only loads once while
// it's used. Every run of a handler starts from its snapshot. The modules
// used the longest time ago are dropped once the snapshots are over max.
type moduleCache struct {
	max int

	mu    sync.Mutex
	size  int
	lru   list.List // of *loadedModule, the most recently used first
	items map[[32]byte]*list.Element
	// Runs of a module that isn't loaded yet wait for one load, or until
	// their context is done. They get its error, and what its top level logs
	// only goes to the first.
	loads singleflight.Group
}

type loadedModule struct {
	key      [32]byte
	snapshot []byte
	manifest *Manifest
	// source is the code as it loaded, which errors point to.
	source string
}

func (c *moduleCache) get(key [32]byte) *loadedModule {
	c.mu.Lock()
	defer c.mu.Unlock()
	el, ok := c.items[key]
	if !ok {
		return nil
	}
	c.lru.MoveToFront(el)
	return el.Value.(*loadedModule)
}

func (c *moduleCache) add(m *loadedModule) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.items == nil {
		c.items = make(map[[32]byte]*list.Element)
	}
	if _, ok := c.items[m.key]; ok {
		return
	}
	c.items[m.key] = c.lru.PushFront(m)
	c.size += len(m.snapshot) + len(m.source)
	for c.size > c.max && c.lru.Len() > 1 {
		old := c.lru.Remove(c.lru.Back()).(*loadedModule)
		delete(c.items, old.key)
		c.size -= len(old.snapshot) + len(old.source)
	}
}

func (c *moduleCache) load(ctx context.Context, opts ModuleOptions) (*loadedModule, error) {
	if len(opts.Code) > MaxModuleLength {
		return nil, &Error{Message: fmt.Sprintf("the module is %d KB, more than the limit of %d KB", len(opts.Code)>>10, MaxModuleLength>>10)}
	}
	key := sha256.Sum256([]byte(opts.Code))
	if m := c.get(key); m != nil {
		return m, nil
	}
	// The load goes on if the caller that started it gives up, as others
	// may wait for it. Its limits end it.
	loadCtx := context.WithoutCancel(ctx)
	ch := c.loads.DoChan(string(key[:]), func() (any, error) {
		if m := c.get(key); m != nil {
			return m, nil
		}
		m := &loadedModule{key: key, source: moduleHead + opts.Code + moduleTail}
		out, snapshot, err := runScript(loadCtx, script{
			limits:   moduleLoadLimits,
			source:   m.source,
			input:    []byte("null"),
			load:     true,
			log:      opts.Log,
			timeout:  opts.timeout,
			wallTime: opts.wallTime,
		})
		if err != nil {
			return nil, err
		}
		if len(out) > MaxManifestLength {
			return nil, &Error{Message: fmt.Sprintf("what the module registers is %d KB, more than the limit of %d KB", len(out)>>10, MaxManifestLength>>10)}
		}
		m.snapshot = snapshot
		m.manifest = &Manifest{}
		if err := json.Unmarshal(out, m.manifest); err != nil {
			return nil, fmt.Errorf("failed to decode manifest: %w", err)
		}
		// The prelude checks all of this, but the code can get around it.
		if err := m.manifest.validate(); err != nil {
			return nil, err
		}
		c.add(m)
		return m, nil
	})
	select {
	case r := <-ch:
		if r.Err != nil {
			return nil, r.Err
		}
		return r.Val.(*loadedModule), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

var (
	eventNameRe     = regexp.MustCompile(`^[a-z_]+$`)
	componentNameRe = regexp.MustCompile(`^[\w-]{1,32}$`)
)

func (m *Manifest) validate() error {
	seen := map[string]bool{}
	unique := func(kind, name string) error {
		if name == "" || seen[kind+" "+name] {
			return fmt.Errorf("invalid manifest: %s %q is empty or registered twice", kind, name)
		}
		seen[kind+" "+name] = true
		return nil
	}
	for _, c := range m.Commands {
		if err := unique("command", c.Name); err != nil {
			return err
		}
		if len(c.Data) == 0 || c.Data[0] != '{' {
			return fmt.Errorf("invalid manifest: the options of command %q aren't an object", c.Name)
		}
	}
	for _, e := range m.Events {
		if err := unique("event", e); err != nil {
			return err
		}
		if !eventNameRe.MatchString(e) {
			return fmt.Errorf("invalid manifest: invalid event %q", e)
		}
	}
	for _, c := range m.Components {
		if err := unique("component", c); err != nil {
			return err
		}
		if !componentNameRe.MatchString(c) {
			return fmt.Errorf("invalid manifest: invalid component %q", c)
		}
	}
	for _, v := range m.Variables {
		if err := unique("variable", v.Name); err != nil {
			return err
		}
	}
	return nil
}

func loadCallMessage(name string) string {
	return name + " can't be used at the top level of a module. Call it inside a handler instead."
}
