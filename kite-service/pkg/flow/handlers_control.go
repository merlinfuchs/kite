package flow

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeControlConditionCompare:     executeControlConditionCompare,
		FlowNodeTypeControlConditionUser:        executeControlConditionCompare,
		FlowNodeTypeControlConditionChannel:     executeControlConditionCompare,
		FlowNodeTypeControlConditionRole:        executeControlConditionCompare,
		FlowNodeTypeControlConditionItemCompare: executeControlConditionItemCompare,
		FlowNodeTypeControlConditionItemUser:    executeControlConditionItemUser,
		FlowNodeTypeControlConditionItemChannel: executeControlConditionItemUser,
		FlowNodeTypeControlConditionItemRole:    executeControlConditionItemUser,
		FlowNodeTypeControlConditionItemElse:    executeControlConditionItemElse,
		FlowNodeTypeControlErrorHandler:         executeControlErrorHandler,
		FlowNodeTypeControlLoop:                 executeControlLoop,
		FlowNodeTypeControlLoopEach:             executeControlLoopEach,
		FlowNodeTypeControlLoopEnd:              executeControlLoopEnd,
		FlowNodeTypeControlLoopExit:             executeControlLoopExit,
		FlowNodeTypeControlFlowEnd:              executeControlFlowEnd,
		FlowNodeTypeControlSleep:                executeControlSleep,
		FlowNodeTypeControlWebhookVerify:        executeControlWebhookVerify,
	})
}

func executeControlConditionCompare(n *CompiledFlowNode, ctx *FlowContext) error {
	baseValue, err := ctx.EvalTemplate(n.Data.ConditionBaseValue)
	if err != nil {
		return err
	}

	// The node state outlives a run of the condition, which can run again in
	// a loop, or inside one of its own branches through a cycle. Each run
	// starts without a match, and the outer run gets its state back after.
	state := ctx.GetNodeState(n.ID)
	outerBaseValue, outerItemMet := state.ConditionBaseValue, state.ConditionItemMet
	defer func() {
		state.ConditionBaseValue, state.ConditionItemMet = outerBaseValue, outerItemMet
	}()

	ctx.StoreNodeBaseValue(n, baseValue)
	state.ConditionItemMet = false

	var elseNode *CompiledFlowNode

	for _, child := range n.Children.Default {
		if child.Type == FlowNodeTypeControlConditionItemElse {
			elseNode = child
		} else {
			if err := child.ExecuteConditionItem(ctx); err != nil {
				return err
			}
		}
	}

	if elseNode != nil {
		// else node has to be executed last
		if err := elseNode.ExecuteConditionItem(ctx); err != nil {
			return err
		}
	}

	return nil
}

func executeControlConditionItemCompare(n *CompiledFlowNode, ctx *FlowContext) error {
	parent := n.FindDirectParentWithType(FlowNodeTypeControlConditionCompare)
	if parent == nil {
		return nil
	}

	parentState := ctx.GetNodeState(parent.ID)

	if parentState.ConditionItemMet && !parent.Data.ConditionAllowMultiple {
		// Another condition item has already been met
		return nil
	}

	// The value can be a secret, like the token a webhook sends, but only to
	// check equality. With a mode like starts with, or an expression around
	// the secret, whoever controls the base value could guess the secret a
	// character at a time. The base value can't be one, as it's stored with
	// the flow's state when the flow suspends. Ignoring case would let a token
	// with the wrong case through.
	mode := n.Data.ConditionItemMode
	hasSecrets := len(secretNames([]string{n.Data.ConditionItemValue})) > 0
	if mode != ComparsionModeEqual && mode != ComparsionModeNotEqual && hasSecrets {
		return fmt.Errorf("secrets can only be compared with equal or not equal")
	}
	if n.Data.ConditionItemIgnoreCase && hasSecrets {
		return fmt.Errorf("secrets can't be compared with ignore case")
	}
	secrets, err := newPlainFlowSecrets(ctx, n.Data.ConditionItemValue)
	if err != nil {
		return err
	}

	itemValue, err := secrets.EvalTemplate(ctx, n.Data.ConditionItemValue)
	if err != nil {
		return err
	}

	baseValue := parentState.ConditionBaseValue
	if n.Data.ConditionItemIgnoreCase {
		baseValue, itemValue = lowerCase(baseValue), lowerCase(itemValue)
	}

	var conditionMet bool
	switch mode {
	case ComparsionModeEqual:
		conditionMet = baseValue.Equals(&itemValue)
	case ComparsionModeNotEqual:
		conditionMet = !baseValue.Equals(&itemValue)
	case ComparsionModeGreaterThan:
		conditionMet = baseValue.GreaterThan(&itemValue)
	case ComparsionModeGreaterThanOrEqual:
		conditionMet = baseValue.GreaterThanOrEqual(&itemValue)
	case ComparsionModeLessThan:
		conditionMet = baseValue.LessThan(&itemValue)
	case ComparsionModeLessThanOrEqual:
		conditionMet = baseValue.LessThanOrEqual(&itemValue)
	case ComparsionModeContains:
		conditionMet = baseValue.Contains(&itemValue)
	case ComparsionModeStartsWith:
		conditionMet = baseValue.StartsWith(&itemValue)
	case ComparsionModeEndsWith:
		conditionMet = baseValue.EndsWith(&itemValue)
	// The values of the list are trimmed, so the base value is too.
	case ComparsionModeOneOf:
		conditionMet = slices.Contains(conditionListValues(itemValue, n.Data.ConditionItemIgnoreCase), strings.TrimSpace(baseValue.String()))
	case ComparsionModeNotOneOf:
		conditionMet = !slices.Contains(conditionListValues(itemValue, n.Data.ConditionItemIgnoreCase), strings.TrimSpace(baseValue.String()))
	case ComparsionModeContainsAny:
		base := baseValue.String()
		conditionMet = slices.ContainsFunc(conditionListValues(itemValue, n.Data.ConditionItemIgnoreCase), func(v string) bool {
			return strings.Contains(base, v)
		})
	}

	if conditionMet {
		parentState.ConditionItemMet = true
		return n.ExecuteChildren(ctx)
	}

	return nil
}

// lowerCase lowercases text to compare it ignoring case. Other values, like
// numbers or users, stay as they are, so they compare the same either way.
func lowerCase(v thing.Thing) thing.Thing {
	if s, ok := v.Value.(string); ok {
		return thing.NewString(strings.ToLower(s))
	}
	return v
}

// conditionListValues are the values of a list mode as text: the items of a
// placeholder that holds a list, like {{result('words')}}, or else one per line.
// Unlike in event filters they aren't separated by commas, as message text
// often contains them.
func conditionListValues(value thing.Thing, ignoreCase bool) []string {
	var values []string
	for _, item := range listOrLines(value) {
		if ignoreCase {
			item = lowerCase(item)
		}
		if v := strings.TrimSpace(item.String()); v != "" {
			values = append(values, v)
		}
	}
	return values
}

func executeControlConditionItemUser(n *CompiledFlowNode, ctx *FlowContext) error {
	parent := n.FindDirectParentWithType(
		FlowNodeTypeControlConditionUser,
		FlowNodeTypeControlConditionChannel,
		FlowNodeTypeControlConditionRole,
	)
	if parent == nil {
		return nil
	}

	parentState := ctx.GetNodeState(parent.ID)

	if parentState.ConditionItemMet && !parent.Data.ConditionAllowMultiple {
		// Another condition item has already been met
		return nil
	}

	itemValue, err := ctx.EvalTemplate(n.Data.ConditionItemValue)
	if err != nil {
		return err
	}

	baseValue := parentState.ConditionBaseValue

	var conditionMet bool
	switch n.Data.ConditionItemMode {
	case ComparsionModeEqual:
		conditionMet = baseValue.Equals(&itemValue)
	case ComparsionModeNotEqual:
		conditionMet = !baseValue.Equals(&itemValue)
	case ComparsionModeHasRole:
		member := baseValue.DiscordMember()
		if !member.User.ID.IsValid() {
			// TODO?: fetch member by id from discord?
			return nil
		}
		conditionMet = slices.Contains(member.RoleIDs, discord.RoleID(itemValue.Int()))
	case ComparsionModeNotHasRole:
		member := baseValue.DiscordMember()
		if !member.User.ID.IsValid() {
			// TODO?: fetch member by id from discord?
			return nil
		}
		conditionMet = !slices.Contains(member.RoleIDs, discord.RoleID(itemValue.Int()))
	case ComparsionModeHasPermission:
		member := baseValue.DiscordMember()
		if !member.User.ID.IsValid() {
			// TODO?: fetch member by id from discord?
			return nil
		}

		roles, err := ctx.Discord.GuildRoles(ctx, ctx.Data.GuildID())
		if err != nil {
			return err
		}

		var permission discord.Permissions
		for _, role := range roles {
			if slices.Contains(member.RoleIDs, role.ID) {
				permission |= role.Permissions
			}
		}

		itemPermissions := discord.Permissions(itemValue.Int())
		conditionMet = permission&itemPermissions == itemPermissions
	case ComparsionModeNotHasPermission:
		member := baseValue.DiscordMember()
		if !member.User.ID.IsValid() {
			// TODO?: fetch member by id from discord?
			return nil
		}

		roles, err := ctx.Discord.GuildRoles(ctx, ctx.Data.GuildID())
		if err != nil {
			return err
		}

		var permission discord.Permissions
		for _, role := range roles {
			if slices.Contains(member.RoleIDs, role.ID) {
				permission |= role.Permissions
			}
		}

		itemPermissions := discord.Permissions(itemValue.Int())
		conditionMet = permission&itemPermissions != itemPermissions
	}

	if conditionMet {
		parentState.ConditionItemMet = true
		return n.ExecuteChildren(ctx)
	}

	return nil
}

func executeControlConditionItemElse(n *CompiledFlowNode, ctx *FlowContext) error {
	parent := n.FindDirectParentWithType(
		FlowNodeTypeControlConditionCompare,
		FlowNodeTypeControlConditionUser,
		FlowNodeTypeControlConditionChannel,
		FlowNodeTypeControlConditionRole,
		FlowNodeTypeControlConditionAIChoice,
	)
	if parent == nil {
		return nil
	}

	parentState := ctx.GetNodeState(parent.ID)

	if parentState.ConditionItemMet {
		// Another condition item has already been met
		return nil
	}

	return n.ExecuteChildren(ctx)
}

func executeControlErrorHandler(n *CompiledFlowNode, ctx *FlowContext) error {
	if err := n.ExecuteChildren(ctx); err != nil {
		return n.handleError(ctx, err)
	}
	return nil
}

func executeControlLoop(n *CompiledFlowNode, ctx *FlowContext) error {
	loopCount, err := ctx.EvalTemplate(n.Data.LoopCount)
	if err != nil {
		return err
	}

	eachNode := n.FindDirectChildWithType(FlowNodeTypeControlLoopEach)
	endNode := n.FindDirectChildWithType(FlowNodeTypeControlLoopEnd)

	// Like for conditions, each run of the loop starts without an exit, and
	// an outer run through a cycle gets its state back after.
	nodeState := ctx.GetNodeState(n.ID)
	outerExited := nodeState.LoopExited
	defer func() { nodeState.LoopExited = outerExited }()
	nodeState.LoopExited = false

	if eachNode != nil {
		for i := 0; i < int(loopCount.Int()); i++ {
			if nodeState.LoopExited || ctx.ended {
				break
			}

			if err := eachNode.Execute(ctx); err != nil {
				return err
			}
		}
	}

	if endNode != nil {
		if err := endNode.Execute(ctx); err != nil {
			return err
		}
	}

	return nil
}

func executeControlLoopEach(n *CompiledFlowNode, ctx *FlowContext) error {
	return n.ExecuteChildren(ctx)
}

func executeControlLoopEnd(n *CompiledFlowNode, ctx *FlowContext) error {
	return n.ExecuteChildren(ctx)
}

func executeControlLoopExit(n *CompiledFlowNode, ctx *FlowContext) error {
	// Mark all parent loops as exited
	parentLoops := n.FindAllParentsWithType(FlowNodeTypeControlLoop)
	for _, loop := range parentLoops {
		ctx.GetNodeState(loop.ID).LoopExited = true
	}

	return nil
}

func executeControlFlowEnd(n *CompiledFlowNode, ctx *FlowContext) error {
	ctx.ended = true
	return nil
}

func executeControlSleep(n *CompiledFlowNode, ctx *FlowContext) error {
	sleepSeconds, err := ctx.EvalTemplate(n.Data.SleepDurationSeconds)
	if err != nil {
		return err
	}

	// Checked before converting, huge values overflow time.Duration.
	// Negative values would overflow into a huge duration too.
	seconds := max(sleepSeconds.Float(), 0)
	if seconds > maxSleepDuration.Seconds() {
		return fmt.Errorf("sleep can't be longer than %d days", int(maxSleepDuration.Hours()/24))
	}
	duration := time.Duration(seconds) * time.Second

	// A resumed flow only runs what comes after the sleep block, so a loop
	// couldn't continue with its next iteration.
	if duration > durableSleepThreshold && !n.inLoop() {
		return n.sleepDurable(ctx, duration)
	}

	deadline, ok := ctx.Deadline()
	if ok && time.Now().Add(duration).After(deadline) {
		return &FlowError{
			Code:    FlowNodeErrorTimeout,
			Message: "sleep would exceed deadline",
		}
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(duration):
		return n.ExecuteChildren(ctx)
	}
}

const (
	// durableSleepThreshold is the longest sleep that keeps the flow running.
	// Longer sleeps end the execution and resume the flow from the database.
	durableSleepThreshold = 5 * time.Second
	maxSleepDuration      = 30 * 24 * time.Hour
	// maxDurableSleeps bounds how often one flow can resume from a durable
	// sleep. Each resume starts with fresh execution limits.
	maxDurableSleeps = 10
)

func (n *CompiledFlowNode) sleepDurable(ctx *FlowContext, duration time.Duration) error {
	if ctx.DurableSleeps >= maxDurableSleeps {
		return fmt.Errorf("a flow can wait at most %d times for longer than %s", maxDurableSleeps, durableSleepThreshold)
	}

	// The execution ends here, so the auto defer would never fire. Discord
	// shows "This interaction failed" for interactions without a response.
	if err := n.deferUnanswered(ctx); err != nil {
		return err
	}

	return ctx.suspendTimer(n.ID, time.Now().UTC().Add(duration))
}

// ResumeAfterSleep continues a flow that suspended in the sleep block n. The
// blocks that led to the sleep don't run again, so errors are handled here
// like they would have been by them.
func (n *CompiledFlowNode) ResumeAfterSleep(ctx *FlowContext) error {
	if err := ctx.startOperation(0); err != nil {
		return traceError(n, err)
	}
	defer ctx.endOperation()

	err := n.ExecuteChildren(ctx)

	// Like normal execution, an error goes to the nearest Error Handler, and
	// an error in its error branch to the next one.
	for _, handler := range n.enclosingErrorHandlers() {
		if err == nil {
			break
		}
		err = handler.handleError(ctx, err)
	}

	if err != nil {
		createDefaultErrorResponse(ctx, err)
	}
	return err
}

// handleError runs the error branch of the Error Handler n.
func (n *CompiledFlowNode) handleError(ctx *FlowContext, err error) error {
	ctx.StoreNodeResult(n, thing.NewString(err.Error()))
	return n.ExecuteChildrenByHandle(ctx, "error")
}

// enclosingErrorHandlers returns the Error Handlers that n runs under, nearest
// first. Those are the ones that reach n through their default branch rather
// than their error branch.
func (n *CompiledFlowNode) enclosingErrorHandlers() []*CompiledFlowNode {
	var res []*CompiledFlowNode
	for _, handler := range n.FindAllParentsWithType(FlowNodeTypeControlErrorHandler) {
		if n.runsUnder(handler.Children.Default) {
			res = append(res, handler)
		}
	}
	return res
}

// inLoop reports whether n runs as part of a loop iteration in its execution.
func (n *CompiledFlowNode) inLoop() bool {
	for _, each := range n.FindAllParentsWithType(FlowNodeTypeControlLoopEach) {
		if n.runsUnder(each.Children.Default) {
			return true
		}
	}
	return false
}

// executeControlWebhookVerify continues with the valid or the invalid branch,
// depending on the signature of the webhook request. It only fails when its
// settings are wrong, never because of the request.
func executeControlWebhookVerify(n *CompiledFlowNode, ctx *FlowContext) error {
	event, ok := ctx.Data.Event().(*webhook.Event)
	if !ok {
		return errors.New("only webhook event listeners can verify a webhook request")
	}

	preset := n.Data.WebhookVerifyPreset
	key, err := webhookVerifyKey(n, ctx, preset)
	if err != nil {
		return err
	}

	var valid bool
	if preset == "" || preset == webhook.PresetCustom {
		valid, err = verifyCustomSignature(n, ctx, key)
	} else {
		valid, err = webhook.VerifyPreset(preset, key, event)
	}
	if err != nil {
		return err
	}

	if valid {
		return n.ExecuteChildrenByHandle(ctx, "valid")
	}
	return n.ExecuteChildrenByHandle(ctx, "invalid")
}

// webhookVerifyKey is the key the signature is checked with. Presets that
// know it ignore one left from another preset.
func webhookVerifyKey(n *CompiledFlowNode, ctx *FlowContext, preset string) (string, error) {
	switch preset {
	case webhook.PresetERLC:
		// Built in.
		return "", nil
	case webhook.PresetDiscord:
		// Discord signs with the app's own key.
		return ctx.Discord.AppVerifyKey(ctx)
	}

	// An expression could use a part of the key, which is easy to guess.
	secrets, err := newPlainFlowSecrets(ctx, n.Data.WebhookVerifyKey)
	if err != nil {
		return "", err
	}
	key, err := secrets.EvalTemplateRaw(ctx, n.Data.WebhookVerifyKey)
	if err != nil {
		return "", err
	}
	return key.String(), nil
}

func verifyCustomSignature(n *CompiledFlowNode, ctx *FlowContext, key string) (bool, error) {
	// The request decides what the templates see, so they can fail for it,
	// like a function called on a header it left out. That's only invalid,
	// unless the template itself is. The log helps to find mistakes.
	evalFailed := func(err error) (bool, error) {
		var compileErr *eval.CompileError
		if errors.As(err, &compileErr) {
			return false, err
		}
		ctx.Log.CreateLogEntry(ctx, provider.LogLevelWarn, fmt.Sprintf("Webhook signature is invalid: %v", err))
		return false, nil
	}

	message, err := ctx.EvalTemplateRaw(n.Data.WebhookVerifyMessage)
	if err != nil {
		return evalFailed(err)
	}
	signature, err := ctx.EvalTemplate(n.Data.WebhookVerifySignature)
	if err != nil {
		return evalFailed(err)
	}

	return webhook.VerifySignature(
		n.Data.WebhookVerifyAlgorithm,
		key,
		message.String(),
		signature.String(),
		n.Data.WebhookVerifySignatureEncoding,
	)
}
