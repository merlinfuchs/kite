---
sidebar_position: 4.6
---

# Integrations

Integrations are the services your blocks can talk to. Discord and Roblox are always connected. Others, like [Cookie API](https://cookie-api.com), need a key from your account with that service, which you enter once under _Integrations_ in your app.

The key is stored encrypted and can't be read back. Kite sends it only to the service it belongs to, with the requests of that service's blocks, so the key is never part of a flow and can't be sent anywhere else by a flow you imported.

Blocks of an integration you haven't connected still show in the block explorer, marked with _Connect_. If a flow has one, the block is marked red in the editor and fails when it runs, until you connect the integration.
