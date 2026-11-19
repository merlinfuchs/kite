# Kite Lite

Status: steps 1 to 6 built 2026-11-18 and tested on a test bot, not deployed. The name is a placeholder.

Kite Lite gives a server pre-built bot features without creating a bot. The server adds the Kite bot, picks modules like verification or suggestions, and fills in their settings. Entering a bot token later turns it into a full Kite app with the flow editor.

The first users come from Xenon. After installing a Xenon template, Xenon offers to set up bot features for the server, and those features run on Kite Lite. Xenon doesn't build them itself.

Lite is deliberately narrow. It runs only official modules, so the Kite bot only ever does things we wrote. That removes most of the risk of a shared bot running user-built logic, which is why the general plan to put every new app on a shared bot is on hold.

## Lite apps

A Lite app is a Kite app row on the Kite bot. Each server has at most one Lite app, and the user who set it up owns it. Everything that works on app IDs today, like commands, event listeners, stored variables, logs, credits and template installs, works for it unchanged.

Modules are official templates marked `shared_bot: true` in their template file. A Lite app can install, remove and configure them, and adding any other template to it is refused. It can't open the flow editor, create commands or listeners, or install user templates. `TestSharedBotBundles` checks that every marked template only uses blocks from an allowlist and events the Kite bot gets with its intents.

Anyone with Manage Server in the server can change its Lite app, not only the owner. A Lite app belongs to the server more than to a person, and that's the permission Discord asks for to add a bot in the first place. Kite checks it with the Kite bot, which fetches the member and works out their permissions. That needs no privileged intent. The owner is who the app is listed under and who can delete it.

Lite apps follow new versions of their templates on their own. The user can't edit the flows, so an update can't conflict with their changes. Parameters and stored variables are kept.

The dashboard shows a Lite app as a reduced app page. It has the modules with their settings and messages marked `customize`, the logs, and a card to upgrade to a full app.

## Credits and plans

Lite apps use credits like any other app, with the same costs per block. Without a subscription a Lite app is on the free plan with 10,000 credits a month. The credit system doesn't change for Lite.

Lite apps can subscribe to Plus. On a Lite app, Plus raises the credits like everywhere else and also turns on server profiles. The Kite bot's global name and avatar are Kite's, so a server profile is the only way for a Lite server to give the bot its own look. Own bots already have their own name and avatar, which is why server profiles stay Premium for them.

The feature check becomes server profiles on the plan, or a Lite app on Plus or higher. The pricing page gets a line under Plus for the Kite bot's name and avatar in your server.

A Lite app that upgrades to its own bot keeps its subscription, and Plus is then Plus as today. Credits work the same on both, so upgrading doesn't change what the modules cost.

When a Lite app runs out of credits, it pauses like any other app, so the Kite bot looks broken in that server. The warning at 80% has to reach Lite owners who never open the Kite dashboard, and the Xenon summary should show it. Pausing only the modules that spend credits can come later.

## Modules

The Kite bot has no privileged intents today. Lite launches with modules that don't need them. Of the current templates, these fit:

- Verification, Suggestions, Giveaway, Moderation and Server Info only use commands and buttons.
- Leveling and Sticky Message listen for new messages but don't read their text. Without Message Content, Discord still sends the events with the text left empty.

Modmail and Reminders are left out, as they send DMs.

We also build one module for each privileged intent we apply for, so the application can show what it's for:

- Server Members: Welcomer, which greets members who join and leave.
- Message Content: AI Automod or Counting, which both read the text of messages. AI Automod costs credits like everywhere else.
- Presence: no template uses it yet. A module that gives roles by activity would need it, but we only apply if we want such a module.

These modules can't run on the Kite bot until Discord grants the intent. The Kite bot is verified and in 160 servers, so it can't turn privileged intents on without Discord's approval. We show the modules to Discord on a second, unverified test application, which has every intent while it's in fewer than 100 servers.

## The Kite bot

The Kite bot is the Discord application 844710218864656414. It's verified, in 160 servers, and has no privileged intents. It doesn't do anything in Discord today. It's added as an app in Kite, and it's the OAuth2 application for logging into Kite.

The Kite app row for it has to go before the shared process starts. Otherwise two gateways connect with the same token, and the commands of that app would be deployed globally and show up in every Lite server.

Logging in and adding the bot are the same Discord application. The install page can ask for `identify` and `bot` in one authorization, so a user who isn't logged in only sees one Discord screen.

The 160 servers it's in have no Lite app. Their events are dropped until someone with Manage Server sets Lite up on the install page.

## Data model

`apps` assumes one token per app. `discord_token` is `NOT NULL`, and `discord_id` is `UNIQUE`, so creating a second app with the same Discord application fails with `app_exists`.

A migration adds `bot_type TEXT NOT NULL DEFAULT 'own'` with the values `own` and `shared`, and `discord_guild_id`, set only for shared apps. A partial unique index on `discord_guild_id WHERE bot_type = 'shared'` keeps it to one Lite app per server. `discord_token` and `discord_id` become nullable and are null for shared apps. Postgres treats nulls as distinct, so the unique index on `discord_id` stays.

The Kite bot's token and application ID are config, like other secrets. Code that reads `app.DiscordID` or the token, like command deploys and `GetUserByAppDiscordID` for bans, has to handle shared apps.

## Shared gateway

The Kite bot runs in its own process, separate from the 4 cluster processes. Its own crashes, stalls and reconnect storms then don't reach own-token apps, and the reverse. It gets its own database connection pool, since connection exhaustion caused the 2026-10-29 stalls.

Today every app has one gateway session and one arikawa state, and the event handler gets the app ID from the session. The shared process has one session for the Kite bot instead. It keeps a map from server ID to Lite app ID, loaded from the database and updated on install and removal, and passes each event on with the mapped app ID. Events from servers without a Lite app and events without a server, like DMs, are dropped before the engine sees them.

The shared process runs its own engine with only the Lite apps. The cluster processes skip apps with `bot_type = 'shared'` in the gateway manager and the engine. Upgrading an app flips `bot_type`, so the shared process drops it and the cluster it hashes to picks it up. The app ID doesn't change.

Discord requires sharding from 2,500 servers. Kite has no sharding code, as no app comes close. The shared session can start unsharded but should be built on arikawa's shard manager from the start, since Xenon installs could pass 2,500 servers quickly.

The intents are the union of what Lite modules need, so at first only non-privileged ones. When Discord grants Server Members, the cache should only keep members for servers whose modules read them, since one cache serves every server.

## API

nginx sends each `/v1/apps/{id}/` request to the cluster the app ID hashes to, and the API reaches an app's Discord session through `AppClient`. Lite apps stay on that routing. For a shared app, `AppClient` returns a REST client with the Kite bot's token instead of the gateway session's client. Reads the dashboard needs, like the channels and roles of the server for parameter pickers, go to Discord's REST API instead of a cache in another process. They are rare enough for that.

So the shared process is only a gateway and an engine. It serves no HTTP, and the API keeps working when it's down.

Command deploys run on the primary cluster today and overwrite the global commands of the app's Discord application. Shared apps get guild commands instead, written with `PUT /applications/{kite bot}/guilds/{server}/commands`. Each server only sees the commands of its own modules.

## Scoping

Every Discord call of a Lite app, from the engine or the API, must target the app's server. The Kite bot is in every Lite server, so without the check one app can act in another.

Curated modules don't make the check optional. Modules take channels and roles as parameters, and a user can paste the ID of a channel in another server into a parameter. The Kite bot would then post into that server. Parameters are checked when they are saved, and every call is checked when it runs.

The check is in `internal/core/sharedbot`, below the engine and the API, so blocks, code, the Discord API Request block and the MCP Discord API go through it without knowing about it. Each Lite app gets its own view of the Kite bot's session, which shares the bot's connection, cache and rate limits.

The view has two parts. Its REST client checks the route of every request against an allowlist before it's sent. Routes under the app's server pass, and routes under a channel pass if the channel is in the server, looked up in the cache or asked from Discord once. A few global routes pass too, like reading users and the routes of interactions and webhook tokens. Everything else is refused with a 403 that never reaches Discord, including DMs and the Kite bot's own profile. Messages that reference a message in another server, like a forward, are refused by their body. Its cache only answers for the app's server, so reads of other servers miss and go to the REST client, which refuses them.

Commands sent over the bot's gateway connection don't go through the view. The engine's `DiscordProvider` refuses status changes for Lite apps, and only joins voice channels of the app's server.

This replaces checking each method of `DiscordProvider` and `policy.target` in `operations.json`, which would each need every new block or operation to remember the check. Sticker and emoji IDs of other servers in a message aren't checked yet.

Template settings with a channel or role are checked against the server when they're saved.

Discord's global rate limit is 50 requests a second per bot token, shared by every Lite server. Each server should get its own request budget so that one busy server can't slow the others, which isn't built yet. Once Discord refuses 200 requests of one app within 10 minutes with a 401, 403 or 429, its requests are refused without sending them for 10 minutes. Refused requests count toward Discord's limit of 10,000 per 10 minutes, which is per IP and already shared with every Kite app on the host.

Own-bot apps count how often their requests name a channel that isn't in the cache, in `shared_bot_shadow_total`. That's how often a Lite app would ask Discord for a channel's server.

## Adding the bot

Users see the two kinds of apps as the Kite bot and a custom bot. Lite is only the name in this doc and in code comments.

One link logs the user in and adds the Kite bot on the same page of Discord: `/v1/auth/install`. The Kite bot's application is also the one users log in with on kite.onl, so Discord asks for both in one authorization, with the scopes `identify`, `email`, `bot` and `applications.commands`. `guild_id` preselects the server, and the rest of the query, like `templates=verification,suggestions&verification.role_id=…` from Xenon, is passed on to the setup page. The dashboard's own button for adding the Kite bot uses the same link. Links people see, like Xenon's, can use a short `kite.onl/add` on the site that forwards to it.

On the callback, Kite logs the user in like a normal login and reads the server from the token, which only Discord can make. The `guild_id` of the redirect could be changed by anyone. Discord only lets users with Manage Server add a bot, so whoever comes back can change the server's app. The first one creates and owns it, later ones become collaborators. Losing Manage Server later doesn't remove them yet. The new app counts toward the 10 apps of its owner, and banned users can't create one.

Then the user lands on the setup page of the app, `/apps/{id}/setup`, where they pick templates and fill in their settings. The settings come after adding the bot, because the channel and role pickers need the bot in the server. The page uses the same install as the Marketplace, which posts panels like the verification message.

The Kite bot's application needs `https://api.kite.onl/v1/auth/install/callback` as a redirect in the Discord Developer Portal. The create app dialog only offers the Kite bot with `NEXT_PUBLIC_KITE_BOT_ENABLED=true` in the web app, which stays off until step 6 is done.

When the bot is removed from a server, its app is stopped with a reason that says so and kept. Adding the bot again starts it. Servers that removed the bot while it was offline are caught when it connects. Servers that added it back while it was offline stay stopped until the next time the bot is added. The 160 servers the bot is in already have no app until someone there adds it again through the link.

Xenon knows the channels and roles it just created from its template, so its link can fill in most settings.

## Xenon API

Later, Xenon sets up the modules in its own dashboard and shows what's configured in Kite. Users then don't need to visit Kite at all, and Xenon and Kite feel like one product.

Xenon's backend calls a partner API with a key from Kite's config. The key only works on Lite apps, and on the server and user Xenon names. Xenon checks that the user has Manage Server before calling, as its dashboard does today. Kite trusts that because the key is ours. User API tokens and Kite's OAuth would each need the user to connect Kite to Xenon first, which is the step this avoids.

The API has a few endpoints.

- List the Lite modules with their parameters, so Xenon can render the settings with its own channel and role pickers.
- Create or get the Lite app of a server, for a Discord user as the owner.
- Install, configure and remove modules.
- Get the summary of a server: whether the bot is in it, the modules with their settings and status, recent errors, and a link to the app in the Kite dashboard.

The app can be set up before the bot is in the server. Xenon then shows Discord's plain bot invite with the server preselected. Kite sees the bot join through the gateway, starts the app and posts the panels. Xenon reads the summary to show that it's done.

`users.email` is `NOT NULL`, and Kite only creates users when they log in. Creating a Lite app for a user who never logged into Kite needs a user row without an email, filled in at their first login.

Xenon knows the channels and roles it just created from its template, so it can fill in most settings for the user.

## Upgrading

The user creates a bot in the Discord developer portal and enters its token on the upgrade card, the same step as creating a Kite app today. Then the app is a full Kite app on their own bot.

1. Kite checks the token and asks the user to add their bot to the server, with the server preselected.
2. Once the bot is in, the app's row gets the token, `discord_id` and `bot_type = 'own'`, and `discord_guild_id` is cleared. The shared process drops the app, and its cluster starts a gateway for it.
3. Kite deploys the commands globally under the new bot and deletes the Kite bot's guild commands in that server.
4. Kite reposts every message the Kite bot sent for the app's message templates from the new bot, and deletes the old ones. A bot can't edit another bot's messages. Buttons on the old messages would also still go to the Kite bot, which no longer has an app for the server.
5. The Kite bot leaves the server.

Stored variables, levels, logs and parameters carry over, since it's the same app. The server's modules become normal installs of their templates, and the user can edit them. A Plus subscription stays on the app.

There's no way back from a full app to Lite.

## Discord approval

The Kite bot is verified already, so it can grow past 100 servers today. As a verified bot, it needs Discord's approval for each privileged intent.

The application for privileged intents lists the modules that need each one and says what they read. Message contents are only read in servers with a module that needs them, kept in memory and never stored. Exceptions are what a module stores by design, like the current count in Counting. The privacy policy has to say this before the application.

Lite doesn't wait for the answer. Modules that need an intent appear once it's granted.

## Order of work

Steps 1 to 6 are built. The shared process runs with `server start --shared_bot` and `shared_bot.token`. Each step is one PR or a short stack, and each one can be deployed without changing anything for own-bot apps.

1. Data model. The migration for `bot_type` and `discord_guild_id`, nullable `discord_token` and `discord_id`, and every place that reads them handling shared apps. The gateway manager, the engine and command deploys skip shared apps. Nothing creates one yet.
2. Shared process, end to end on a test bot. The Kite app row of the Kite bot is removed before the process runs with its token. The config for the process, one unsharded session, the map from server to app, the engine for shared apps, guild command deploys, and `AppClient` with the shared token in the API. A Lite app is created by hand and runs Verification in a test server. This is the step that finds the places that still assume one token per app, so it comes early.
3. Scoping. The view of the session for each Lite app, the check of settings on save, and pausing apps whose requests Discord keeps refusing. Own-bot apps count how often a channel isn't in the cache. No Lite app is created outside our test servers before this is merged.
4. Modules. The `shared_bot` flag in the template format, the test that marked templates only use allowed blocks, and the flag on the launch modules.
5. Adding the bot. The link that logs in and adds the Kite bot at once, the app for the server on the callback, stopping and starting apps when the bot leaves and comes back, and the setup page in the dashboard.
6. The Lite app page in the dashboard. The API refuses every change to commands, event listeners, message templates, Flow AI, secrets, integrations and plugins of Lite apps, so the API and MCP can't build anything on the Kite bot. The sidebar has four pages. The dashboard shows whether the bot can run the templates, with recent errors and the credits, and the added templates with their settings. The others are the logs, the plan and a few settings. The setup page has no sidebar, so nothing distracts from picking templates. There's no editor, invite link, bot profile, token or collaborator list, as Manage Server decides who can change the app. Deleting a Lite app removes its commands from the server. There's no upgrade card yet, as it would promise step 9.
7. Plus on Lite. Server profiles for Lite apps on Plus, checkout from a Lite app, the line on the pricing page, and the credit warnings reaching Lite owners.
8. Launch with Xenon. Xenon links to the install page after a template install. We watch the first servers, the logs and the cost before Xenon shows it to everyone.
9. Upgrading. Until this is done, a Lite server that wants the editor creates a new app.
10. The intent application. Welcomer and AI Automod or Counting run on a test application, the privacy policy says how message contents are handled, and we apply. This can start any time after step 4 and runs in parallel, as Discord takes a while.
11. The Xenon API, and users without an email.
12. Request budgets per server and sharding, before the Kite bot gets close to 2,500 servers.

Start with steps 1 and 2. They're the most uncertain part and change nothing in prod until a Lite app exists. Step 3 follows right after, since everything later depends on it.

## Open questions

- When the owner leaves the server or loses Manage Server, does the app move to someone else? A Plus subscription stays with whoever pays for it.
