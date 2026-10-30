---
sidebar_position: 4.2
---

# Publishing a Template

A template in the marketplace is added by servers you've never seen, with their own channels, roles and API keys. This guide is about building one that works for them, and getting it approved. How to create, update and share templates is in [Templates and Sharing](../reference/templates.md#your-own-templates).

## Make it Work in Any Server

Whatever is specific to your server has to be a [parameter](../reference/parameters.md). Write `{{params.log_channel}}` in a block instead of picking your log channel, and give the parameter the type channel. Whoever adds the template picks their own channel when they add it. The same goes for roles, categories and any text a server would want to change, like the instructions of an AI block.

Give every parameter a name and a description that someone who doesn't know your server understands, like `Staff role` and "Members with this role can close tickets". Mark the ones a template can't work without as required.

Under `Default values` you can share the value of a text parameter, which apps start with. Only share text that works for everyone. Never share anything private, like the name of your server or a key.

Check your messages and replies for things that only fit your server, like its name, its rules channel or the names of your staff. Message templates can't use parameters yet, so write their text so it fits any server.

## API Keys and Other Secrets

Never type an API key, token or password into a block. Everyone who adds the template would see it and use it, and pay for it with your account. Put it in a [secret](../reference/secrets.md) instead, and write `{{secrets.NAME}}` where the block needs it.

Secrets are never part of a template. Its blocks use the secret of the same name in the app that adds it, so its page lists the secrets it reads. Explain in `Before you start` where to get each key and what to name the secret, like "Create an API key at openweathermap.org and save it as the secret `OPENWEATHER_KEY`".

Only send a secret, or the bot token, to the service that needs it, and say on the page why. A template that sends keys or members' data somewhere its page doesn't mention isn't approved.

## Write a Clear Page

The page is all someone sees before adding your template, so write it for someone who has never seen your server.

- **Name:** what it does, and what sets it apart from similar templates, like Kite's own. `Moderation with Warnings` works better than `Moderation` or `Moderation by Alex`. Your name is already on the page.
- **Description:** one sentence on what it does.
- **How it works:** what each command and event listener does, and what members and staff see.
- **Before you start:** what to set up first, like roles, channel permissions, secrets or an account at another service. Leave it empty if there's nothing.
- **Categories and icon:** up to 3 categories it's useful for, and an icon that fits what it does. The first one you pick is its main one, the marketplace shows it in all of them.

## Before You Submit

Add your template to a test server, or ask someone to add it, and use every command once. That catches most problems before a review does:

- A channel or role that's still your server's instead of a parameter.
- A command that doesn't answer, or fails without saying why. Add a reply for the error output of blocks that can fail, like API requests.
- A block left empty, or a flow that ends early.
- A permission the bot needs but the page doesn't mention.

## What Kite Approves

Kite approves templates that work, are useful to other servers and follow the [template rules](https://kite.onl/terms#moderation). They don't need to be polished. Kite rejects a template when:

- It does something its page doesn't mention, like banning members or sending data to other websites.
- It's broken, for example because of empty blocks or commands that never answer.
- It only works in your server, because its blocks use the IDs of your channels or roles instead of parameters.
- It copies a Kite template or another listed one without adding anything.
- Its name, description or `How it works` are hard to follow.
- Its name and description don't tell it apart from similar templates.

An automatic check looks at every submission first. It rejects what's clearly not meant for the marketplace, like tests, templates that do nothing and scams, and a person decides on everything else. A rejection says what to change. Fix it and submit it again, or publish an update, which submits it again too. If you think a rejection is wrong, ask on the [Kite Discord server](https://discord.gg/rNd9jWHnXh).

## After it's Listed

New versions and changes to the page are reviewed too. Until they're approved, apps keep the last approved version and the marketplace shows the last approved page. Say what changed when you publish an update, apps that added your template see it.

Your listed templates are on your creator page, which your name on their pages links to.
