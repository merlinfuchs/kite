---
sidebar_position: 3
---

# Using the Flow Editor

You build commands and event listeners in the flow editor by adding blocks and connecting them with wires. The editor works with a mouse and keyboard on a computer and with touch controls on a phone.

## Navigating the Canvas

Drag the background with your left mouse button to pan around, and use the mouse wheel to zoom in and out.

- **Center flow**: Right-click the background and choose `Center flow` to bring all your blocks into view.
- **Reset zoom**: Right-click the background and choose `Reset zoom` to return to default 100% zoom.

## Adding Blocks

- **Drag and drop**: Drag any block from the catalog on the left directly onto the canvas.
- **Bulk add**: Right click any block in the catalog to add 1, 2, 3, or a custom quantity to your canvas at once.
- **Collapsible categories**: Click category headers in the sidebar (like Responses, Messages, etc) to collapse or expand them.

## Connecting Blocks and Wires

- **Wiring blocks**: Drag from the output handle of one block to the input handle of another block.
- **Deleting wires**: Click the `x` button over the wire, or right click the wire and click `Delete connection`.
- **Inserting a block into a wire**: Right click any wire and choose `Insert block here...`. Pick a block from the catalog, and Kite connects it in between the two blocks already connected. Fixed branch wires (like the ones that are between a Condition and its Else branch) cannot be spliced.

## Selecting Blocks

- **Box select**: Right click and drag across the canvas to draw a selection box. All blocks that are mostly inside the box are selected.
- **Add to selection**: Hold `Shift` while right click dragging to add more blocks to your current selection without deselecting what you already picked.
- **Add single blocks to selection**: Hold `Shift` and right click on any block to toggle its selection on or off without opening the context menu.
- **Select all**: Right click the canvas and choose `Select all in view` (blocks visible on screen) or `Select all in flow` (every block in the flow).

## Editing Blocks and Multi-Selection

Clicking a block opens its settings in the sidebar on the left.

- **Single block**: The sidebar displays that block's settings as well as some information on the block.
- **Multiple blocks**: When you select more than one block, the sidebar displays a **Selected block** dropdown at the top. Use the dropdown to switch between any of the selected blocks to edit their settings without losing your canvas multi-selection.
- **Moving multiple blocks**: Dragging any selected block moves all selected blocks together.

## Context Menus

Right clicking gives you actions depending on what you click;

### On a Block

- **Rename (`F2`)**: Give the block a custom label to keep flows organized.
- **Copy (`Ctrl + C`)**: Copies selected block(s) to your clipboard.
- **Cut (`Ctrl + X`)**: Copies and deletes selected block(s).
- **Duplicate (`Ctrl + D`)**: Creates a duplicate copy nearby.
- **Duplicate with connections**: Duplicates selected block(s) while preserving the wires connecting them to each other.
- **Disconnect**: Unplugs all wires attached to the block without deleting it.
- **Delete (`Delete` / `Backspace`)**: Removes the block and its wires.

### On a Wire

- **Insert block here...**: Opens the block picker and splices the chosen block into the connection.
- **Delete connection**: Deletes the wire.

### On the Background

- **Paste (`Ctrl + V`)**: Pastes copied block(s) where your mouse cursor is.
- **Undo (`Ctrl + Z`)** & **Redo (`Ctrl + Y`)**: Steps backward or forward through your changes.
- **Auto arrange layout**: Automatically cleans up and organizes your blocks into a tidy layout.
- **Center flow** & **Reset zoom**: Centers the flow on screen or resets zoom to 100%.
- **Select all**: Selects all blocks currently in view or across the entire flow.
- **Clear selection**: Deselects all blocks.

## Keyboard Shortcuts

| Shortcut | Action |
| :--- | :--- |
| `Ctrl` + `C` | Copy selected block(s) |
| `Ctrl` + `X` | Cut selected block(s) |
| `Ctrl` + `V` | Paste at cursor position |
| `Ctrl` + `D` | Duplicate selected block(s) |
| `Ctrl` + `Z` | Undo |
| `Ctrl` + `Y` / `Ctrl` + `Shift` + `Z` | Redo |
| `F2` | Rename custom block label |
| `Delete` / `Backspace` | Delete selected blocks or wires |
| `Right-click` + drag | Box select |
| `Shift` + `Right-click` + drag | Additive box select |
| `Shift` + `Right-click` (tap) | Toggle block in selection |

*On macOS, use `Cmd` instead of `Ctrl`.*

## On Mobile

### Moving and Zooming

- Drag with one finger on the canvas to pan.
- Pinch with two fingers to zoom.
- Tap `Center flow` in the bottom bar to reset your view.

### Selecting Blocks on Mobile

- **Long-press to select**: Press and hold a block until your phone vibrates to enter selection mode.
- **Multi-select**: Once in selection mode, tap any other blocks to add or remove them from the selection.
- **Selected drawer**: Tap `Selected` in the bottom bar to view and change settings for your selected blocks, move, copy, cut, duplicate, disconnect, or delete them.
- **Move Mode**: Toggle move mode to drag blocks around easily without accidentally panning the canvas.

### Inserting into Wires on Mobile

- Press and hold any wire for one second. Your phone will vibrate and open the block catalog. Tap any block to splice it into the wire.

### Mobile Toolbar

The bottom bar gives you quick access to:
- **Add Block**: Opens the block catalog. Long press any block in the sheet to add multiple copies at once.
- **Quick actions**: Buttons for Delete, Paste, Undo, Redo, Center Flow, Auto Arrange Layout, and your App Logs.

