---
sidebar_position: 4.6
---

# Integrations

Integrations are the services your blocks can talk to. You manage them under _Integrations_ in your app. Discord is always enabled. Roblox is enabled until you turn it off. Others, like [Cookie API](https://cookie-api.com), need a key from your account with that service, which you enter when you enable them. Disabling one keeps its key, removing it deletes the key.

Before you enable one, Kite shows what its blocks send to the service, with links to its privacy policy and terms. Cookie API's transcript block also sends your bot's token, as Cookie API reads the channel's messages itself.

The key is stored encrypted and can't be read back. Kite sends it only to the service it belongs to, with the requests of that service's blocks, so the key is never part of a flow and can't be sent anywhere else by a flow you imported.

Blocks of an integration you haven't enabled still show in the block explorer, marked with _Enable_. If a flow has one, the block is marked red in the editor and fails when it runs, until you enable the integration.
