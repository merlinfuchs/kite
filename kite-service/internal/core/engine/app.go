package engine

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/sashabaranov/go-openai"
)

type App struct {
	sync.RWMutex

	id string

	config               EngineConfig
	appStore             store.AppStore
	logStore             store.LogStore
	usageStore           store.UsageStore
	messageStore         store.MessageStore
	messageInstanceStore store.MessageInstanceStore
	commandStore         store.CommandStore
	variableValueStore   store.VariableValueStore
	httpClient           *http.Client
	openaiClient         *openai.Client
	hasUndeployedChanges bool

	commands map[string]*Command
	// TODO?: Cache messages (LRUCache<*MessageInstance>)
	listeners map[string]*EventListener
}

func NewApp(
	config EngineConfig,
	id string,
	appStore store.AppStore,
	logStore store.LogStore,
	usageStore store.UsageStore,
	messageStore store.MessageStore,
	messageInstanceStore store.MessageInstanceStore,
	commandStore store.CommandStore,
	variableValueStore store.VariableValueStore,
	httpClient *http.Client,
	openaiClient *openai.Client,
) *App {
	return &App{
		id:                   id,
		config:               config,
		appStore:             appStore,
		logStore:             logStore,
		usageStore:           usageStore,
		messageStore:         messageStore,
		messageInstanceStore: messageInstanceStore,
		commandStore:         commandStore,
		variableValueStore:   variableValueStore,
		httpClient:           httpClient,
		commands:             make(map[string]*Command),
		listeners:            make(map[string]*EventListener),
		openaiClient:         openaiClient,
	}
}

func (a *App) AddCommand(cmd *model.Command) {
	a.Lock()
	defer a.Unlock()

	command, err := NewCommand(
		a.config,
		cmd,
		a.appStore,
		a.logStore,
		a.usageStore,
		a.messageStore,
		a.messageInstanceStore,
		a.variableValueStore,
		a.httpClient,
		a.openaiClient,
	)
	if err != nil {
		slog.With("error", err).Error("failed to create command")
		return
	}

	a.commands[cmd.ID] = command

	if !cmd.LastDeployedAt.Valid || cmd.LastDeployedAt.Time.Before(cmd.UpdatedAt) {
		a.hasUndeployedChanges = true
	}
}

func (a *App) RemoveDanglingCommands(commandIDs []string) {
	a.Lock()
	defer a.Unlock()

	commandIDMap := make(map[string]struct{}, len(commandIDs))
	for _, commandID := range commandIDs {
		commandIDMap[commandID] = struct{}{}
	}

	for cmdID := range a.commands {
		if _, ok := commandIDMap[cmdID]; !ok {
			delete(a.commands, cmdID)
			a.hasUndeployedChanges = true
		}
	}
}

func (a *App) AddEventListener(listener *model.EventListener) {
	a.Lock()
	defer a.Unlock()

	eventListener, err := NewEventListener(
		a.config,
		listener,
		a.appStore,
		a.logStore,
		a.usageStore,
		a.messageStore,
		a.messageInstanceStore,
		a.variableValueStore,
		a.httpClient,
		a.openaiClient,
	)
	if err != nil {
		slog.With("error", err).Error("failed to create event listener")
		return
	}

	a.listeners[listener.ID] = eventListener
}

func (a *App) RemoveDanglingEventListeners(listenerIDs []string) {
	a.Lock()
	defer a.Unlock()

	listenerIDMap := make(map[string]struct{}, len(listenerIDs))
	for _, listenerID := range listenerIDs {
		listenerIDMap[listenerID] = struct{}{}
	}

	for listenerID := range a.listeners {
		if _, ok := listenerIDMap[listenerID]; !ok {
			delete(a.listeners, listenerID)
			a.hasUndeployedChanges = true
		}
	}
}

func (a *App) createLogEntry(level model.LogLevel, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Create log entry which will be displayed in the dashboard
	err := a.logStore.CreateLogEntry(ctx, model.LogEntry{
		AppID:     a.id,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		slog.With("error", err).With("app_id", a.id).Error("Failed to create log entry from engine app")
	}
}

func (a *App) HandleEvent(appID string, session *state.State, event gateway.Event) {
	a.RLock()
	defer a.RUnlock()

	switch e := event.(type) {
	case *gateway.InteractionCreateEvent:
		switch d := e.Data.(type) {
		case *discord.CommandInteraction:
			fullName := getFullCommandName(d)
			for _, command := range a.commands {
				if command.cmd.Name == fullName {
					go command.HandleEvent(appID, session, event)
				}
			}
		case *discord.ButtonInteraction:
			messageID := e.Message.ID.String()
			messageInstnace, err := a.messageInstanceStore.MessageInstanceByDiscordMessageID(context.TODO(), messageID)
			if err != nil {
				if errors.Is(err, store.ErrNotFound) {
					return
				}

				slog.With("error", err).Error("failed to get message instance by discord message ID")
				return
			}

			instance, err := NewMessageInstance(
				a.config,
				a.id,
				messageInstnace,
				a.appStore,
				a.logStore,
				a.usageStore,
				a.messageStore,
				a.messageInstanceStore,
				a.variableValueStore,
				a.httpClient,
				a.openaiClient,
			)
			if err != nil {
				slog.With("error", err).Error("failed to create message instance")
				return
			}

			go instance.HandleEvent(appID, session, event)
		}
	default:
		eventType := model.EventTypeFromDiscordEventType(e.EventType())
		for _, listener := range a.listeners {
			if listener.listener.Source != model.EventSourceDiscord {
				continue
			}

			if listener.listener.Type != eventType {
				continue
			}

			listener.HandleEvent(appID, session, event)
		}
	}
}

func getFullCommandName(d *discord.CommandInteraction) string {
	fullName := d.Name
	for _, option := range d.Options {
		if option.Type == discord.SubcommandOptionType {
			fullName += " " + option.Name
			break
		} else if option.Type == discord.SubcommandGroupOptionType {
			fullName += " " + option.Name
			for _, subOption := range option.Options {
				fullName += " " + subOption.Name
			}
			break
		}
	}

	return fullName
}
