package js

import (
	"container/list"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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
	_, err = runScript(ctx, script{
		limits:   moduleLimits,
		eval:     moduleCaller,
		head:     moduleHead,
		code:     opts.Code,
		tail:     moduleTail,
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
	// Runs of a module that isn't loaded yet wait for one load. They get
	// its error, and what its top level logs only goes to the first.
	loads singleflight.Group
}

type loadedModule struct {
	key      [32]byte
	snapshot []byte
	manifest *Manifest
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
	c.size += len(m.snapshot)
	for c.size > c.max && c.lru.Len() > 1 {
		old := c.lru.Remove(c.lru.Back()).(*loadedModule)
		delete(c.items, old.key)
		c.size -= len(old.snapshot)
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
	v, err, _ := c.loads.Do(string(key[:]), func() (any, error) {
		if m := c.get(key); m != nil {
			return m, nil
		}
		m := &loadedModule{key: key}
		out, err := runScript(ctx, script{
			limits:   moduleLoadLimits,
			eval:     moduleHead + opts.Code + moduleTail,
			head:     moduleHead,
			code:     opts.Code,
			tail:     moduleTail,
			input:    []byte("null"),
			log:      opts.Log,
			timeout:  opts.timeout,
			wallTime: opts.wallTime,
			keep:     func(snapshot []byte) { m.snapshot = snapshot },
		})
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(out, &m.manifest); err != nil {
			return nil, fmt.Errorf("failed to decode manifest: %w", err)
		}
		c.add(m)
		return m, nil
	})
	if err != nil {
		return nil, err
	}
	return v.(*loadedModule), nil
}
