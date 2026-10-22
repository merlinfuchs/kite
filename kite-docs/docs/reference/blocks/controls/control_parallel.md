---
sidebar_position: 43.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Run in Parallel

<EmbedFlowNode type="control_parallel" />

The `Run in Parallel` block runs the blocks connected to its `Run at the Same Time` output at the same time, instead of one after the other. Once all of them finished, the blocks connected to its `After All Finished` output run.

This is useful when the flow does several things that don't depend on each other, like fetching two members or sending messages to different channels. The flow only waits as long as the slowest of them, instead of all of them added up.

The blocks after `After All Finished` can use the results of all branches. Blocks in one branch can't use the results of another branch, because they run at the same time.

A few things to keep in mind:

- A block can only be in one branch. To run something once all branches finished, connect it to `After All Finished` instead of to the end of several branches.
- Only one branch should respond to a command or button. Two branches responding at the same time fail, so send follow-up messages or messages to the channel from the others.
- If a block in one branch fails, the other branches still finish. The blocks after `After All Finished` don't run, and the error goes to the nearest [Handle Errors](./control_error_handler.md) block around the `Run in Parallel` block.
- [Wait](./control_sleep.md) blocks inside a branch always happen within the current run, so they're limited by its 30 second execution time. Waits after `After All Finished` work as usual.
- All branches share the credits and operations limits of the run.

<NodeInfoExplorer type="control_parallel" />
