# Kite

Make your own Discord Bot with Kite for free without a single line of code. With support for slash commands, buttons, and more.

![Flow Example](./example-flow.png)

## TODO

- [x] App home page
- [x] App settings page
- [x] Enforce limits (e.g. max number of apps)
- [x] Implement validation on flow data
- [x] Implement command registration with all supported options
- [x] Design and implement flow values properly
- [x] Design and implement node types
- [x] Design and implement node data
- [ ] Design and implement variable system
  - [x] Implement template engine
  - Tenorary vs Persisted?
  - Variables
    - statically defined
    - persisted
    - Scopeed by guild, user, member, channel, global, or custom
  - Fields
    - dynamically defined
- [x] Add some more common flow nodes
- [ ] Implement all existing flow nodes
- [x] Merge kite-common and kite-service
- [x] Move flow and template into independent modules
- [x] Add invite to app home page
- [x] Detect correct intents before connecting
- [x] Add button to Enable and Disable bot in app settings
- [x] Merge engine and gateway so the engine can access gateway state?

### Node Types

#### Entry & Options

- [x] Command entry
- [x] Command arguments
- [x] Command permission check
- [x] Command context check

- [ ] Error entry?

### Control Flow

- [x] Compare condition
- [ ] Permissions condition?
- [ ] Argument conditions?

- [ ] Loop N times
- [ ] Sleep

#### Actions

- [x] Create interaction response

  - [x] Plain text
  - [ ] Embeds
  - [ ] Components

- [x] Create message
  - [x] Current channel
  - [ ] Other channel
- [ ] Delete message
- [ ] Edit message
- [ ] Purge messages

- [ ] Create channel
- [ ] Delete channel
- [ ] Edit channel

- [ ] Create thread
- [ ] Delete thread
- [ ] Edit thread

- [ ] Create role
- [ ] Delete role
- [ ] Edit role

- [ ] Ban user
- [ ] Kick user
- [ ] Timeout member
- [ ] Edit member

- [ ] Set variable
- [ ] Delete variable

- [ ] Set KV key
- [ ] Delete KV key

- [ ] API request
- [ ] Random number
