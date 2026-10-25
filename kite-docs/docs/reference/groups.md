---
sidebar_position: 4.46
---

# Groups

Groups are labeled frames around blocks that keep big commands and event listeners organized. Collapse a group to shrink the blocks in it to a single bar, so a long flow fits on the screen. Groups only change how the flow looks in the editor, not how it runs, and they're saved and shared with the flow like [notes](./notes.md).

## Creating a group

Select the blocks you want to group, by holding `Shift` and dragging a box around them, and press `Ctrl+G` (`Cmd+G` on Mac), click the group button in the controls at the bottom right of the editor, or right click the selection and pick _Group_. Without a selection, an empty group is added in the middle of the screen, or where you right clicked the canvas and picked _Add group_.

A block or note is in a group when its center is inside the frame. To add a block to a group, drag it into the frame. To take it out, drag it out of the frame. Groups can be inside other groups.

## Editing a group

- **Rename**: double-click the name, or pick _Rename_ in the group's `...` menu.
- **Color**: pick a color in the `...` menu.
- **Move**: drag the group by its header. The blocks and notes in it move with it.
- **Resize**: select the group and drag its corners or edges, or pick _Fit to blocks_ in the `...` menu to fit it around what's in it.
- **Select blocks**: pick _Select blocks_ in the `...` menu to select everything in the group, for example to move or copy it.

## Collapsing a group

Click the arrow on the right of the header to collapse the group to a single bar that shows its name and how many blocks are in it. Click it again, or double-click the bar, to expand the group. Connections to the blocks in a collapsed group are drawn to the group as dashed lines.

Blocks you drag over a collapsed group don't join it, so they never disappear by accident.

## Copying, removing and arranging

- **Copy, cut and duplicate**: a group is copied together with the blocks and notes in it, and pasted expanded.
- **Ungroup**: removes the frame and keeps the blocks. Selecting the group and pressing `Delete` does the same.
- **Delete group and blocks**: removes the frame together with the blocks in it.

All of these can be undone with `Ctrl+Z`.

Auto arrange keeps the blocks of each group together and fits the groups around them. Collapsed groups are placed like a single block, and their blocks move with them.
