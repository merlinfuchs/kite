package flow

import (
	"fmt"
	"slices"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
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
		FlowNodeTypeControlSleep:                executeControlSleep,
	})
}

func executeControlConditionCompare(n *CompiledFlowNode, ctx *FlowContext) error {
	baseValue, err := ctx.EvalTemplate(n.Data.ConditionBaseValue)
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeBaseValue(n, baseValue)

	var elseNode *CompiledFlowNode

	for _, child := range n.Children.Default {
		if child.Type == FlowNodeTypeControlConditionItemElse {
			elseNode = child
		} else {
			if err := child.Execute(ctx); err != nil {
				return traceError(n, err)
			}
		}
	}

	if elseNode != nil {
		// else node has to be executed last
		if err := elseNode.Execute(ctx); err != nil {
			return traceError(n, err)
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

	itemValue, err := ctx.EvalTemplate(n.Data.ConditionItemValue)
	if err != nil {
		return traceError(n, err)
	}

	baseValue := parentState.ConditionBaseValue

	var conditionMet bool
	switch n.Data.ConditionItemMode {
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
	}

	if conditionMet {
		parentState.ConditionItemMet = true
		return n.ExecuteChildren(ctx)
	}

	return nil
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
		return traceError(n, err)
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
			return traceError(n, err)
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
			return traceError(n, err)
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
		return traceError(n, err)
	}

	eachNode := n.FindDirectChildWithType(FlowNodeTypeControlLoopEach)
	endNode := n.FindDirectChildWithType(FlowNodeTypeControlLoopEnd)

	nodeState := ctx.GetNodeState(n.ID)

	for i := 0; i < int(loopCount.Int()); i++ {
		if nodeState.LoopExited {
			break
		}

		if err := eachNode.Execute(ctx); err != nil {
			return traceError(n, err)
		}
	}

	if err := endNode.Execute(ctx); err != nil {
		return traceError(n, err)
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

func executeControlSleep(n *CompiledFlowNode, ctx *FlowContext) error {
	sleepSeconds, err := ctx.EvalTemplate(n.Data.SleepDurationSeconds)
	if err != nil {
		return traceError(n, err)
	}

	// Checked before converting, huge values overflow time.Duration.
	// Negative values would overflow into a huge duration too.
	seconds := max(sleepSeconds.Float(), 0)
	if seconds > maxSleepDuration.Seconds() {
		return traceError(n, fmt.Errorf("sleep can't be longer than %d days", int(maxSleepDuration.Hours()/24)))
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
		return traceError(n, fmt.Errorf("a flow can wait at most %d times for longer than %s", maxDurableSleeps, durableSleepThreshold))
	}

	// The execution ends here, so the auto defer would never fire. Discord
	// shows "This interaction failed" for interactions without a response.
	if err := n.deferUnanswered(ctx); err != nil {
		return traceError(n, err)
	}

	if err := ctx.suspendTimer(n.ID, time.Now().UTC().Add(duration)); err != nil {
		return traceError(n, err)
	}
	return nil
}

// ResumeAfterSleep continues a flow that suspended in the sleep block n. The
// blocks that led to the sleep don't run again, so errors are handled here
// like they would have been by them.
func (n *CompiledFlowNode) ResumeAfterSleep(ctx *FlowContext) error {
	if err := ctx.startOperation(0); err != nil {
		return err
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
