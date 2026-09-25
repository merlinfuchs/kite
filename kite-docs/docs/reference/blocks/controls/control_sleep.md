---
sidebar_position: 44
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Wait

<EmbedFlowNode type="control_sleep" />

The `Wait` block pauses the flow execution for a specified amount of time. This is useful for creating delays, rate limiting, or scheduling actions to occur after a certain period.

Waits of up to 5 seconds happen within the current run. Longer waits, up to 30 days, end the current run and continue the flow when the time is up, so they survive restarts. A few things behave differently after a long wait:

- Discord only accepts responses to a command or button for 15 minutes. After a longer wait, response blocks fail, so send a message to the channel instead. Kite acknowledges the interaction before the wait starts, so Discord doesn't show it as failed.
- Inside a [loop](./control_loop.md), waits always happen within the current run, so they're limited by its 30 second execution time.
- Blocks in other branches of the flow don't wait. If a block has two branches and only one contains a long wait, the other branch runs right away.
- An Error Handler around the Wait block still catches errors of the blocks after it.
- If you delete the Wait block while a flow is waiting in it, the flow won't continue.
- A single run can wait longer than 5 seconds at most 10 times in a row, so a Wait block in a cycle can't keep it going forever. Each button click or modal submission starts counting again, so menus can be used indefinitely.
- An app can have at most 1000 flows waiting at the same time.

<NodeInfoExplorer type="control_sleep" />
