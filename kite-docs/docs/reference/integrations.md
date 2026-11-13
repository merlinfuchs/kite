---
sidebar_position: 4.6
---

# Integrations

Integrations are the services your blocks can talk to. You manage them under _Integrations_ in your app. Discord and Roblox are always enabled. Others, like [Bloxlink](https://blox.link), [Roblox Open Cloud](#roblox-open-cloud), [Cookie API](https://cookie-api.com) and [ER:LC](https://apidocs.erlc.gg), need a key from your account with that service, which you enter when you enable them. Disabling one keeps its key, removing it deletes the key.

Before you enable one, Kite shows what its blocks send to the service, with links to its privacy policy and terms if it has them. Cookie API's transcript block also sends your bot's token, as Cookie API reads the channel's messages itself.

ER:LC only runs commands from apps the owner of the private server authorized. After you enter the server key, ER:LC has an _Authorize Kite_ button, which opens ER:LC's page to do that. The [ER:LC guide](../guides/erlc.md) walks through setting it up and building commands for your server.

The key is stored encrypted and can't be read back. Kite sends it only to the service it belongs to, with the requests of that service's blocks, so the key is never part of a flow and can't be sent anywhere else by a flow you imported.

Blocks of an integration you haven't enabled still show in the block explorer, marked with _Enable_. If a flow has one, the block is marked red in the editor and fails when it runs, until you enable the integration.

## Roblox Open Cloud

The Roblox group blocks use an API key from Roblox's Creator Hub. Create it on the account that manages the group, as the key can only do what that account can:

1. Open [API keys](https://create.roblox.com/dashboard/credentials?activeTab=ApiKeysTab) of your own account, not the group's, and create an API key. Keys owned by a group can't use the Groups API. If Roblox shows "Group API Keys do not support all Open Cloud APIs", you're on the group's page.
2. Under _Access Permissions_, select the _groups_ API system and add the _group:read_ and _group:write_ operations.
3. Don't restrict the key to IP addresses. If Roblox asks for accepted IP addresses, add `0.0.0.0/0`.
4. Copy the key and enter it when you enable Roblox Open Cloud in Kite.
