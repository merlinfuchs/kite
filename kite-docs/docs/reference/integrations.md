---
sidebar_position: 4.6
---

# Integrations

Integrations are the services your blocks can talk to. You manage them under _Integrations_ in your app. Discord is always enabled. Roblox is enabled until you turn it off. Others, like [Cookie API](https://cookie-api.com), need a key from your account with that service, which you enter once to connect them.

The key is stored encrypted and can't be read back. Kite sends it only to the service it belongs to, with the requests of that service's blocks, so the key is never part of a flow and can't be sent anywhere else by a flow you imported.

Blocks of an integration you haven't enabled or connected still show in the block explorer, marked with _Enable_ or _Connect_. If a flow has one, the block is marked red in the editor and fails when it runs, until you enable the integration.
