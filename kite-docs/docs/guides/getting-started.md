---
sidebar_position: 1
---

# Getting Started

Let's get started! The first step is to create a new Discord application!

:::tip

Discord uses the terms `bot` and `app` interchangeably in their documentation and basically everywhere else. While `app` is technically the right term, we also some times use the term `bot` here.

:::

## Creating Your App

Open the [Discord Developer Portal](https://discord.com/developers/applications) and sign-in with your Discord account. This is where you create and manage your Discord applications.

1. Click on `New Application` in the top-right corner and give your app a name
2. Upload an icon for your app
3. Make sure the `Interaction Endpoint Url` is blank
4. Switch to the `Bot` section on the left side
5. Enable the `Presence Intent`, `Server Members Intent`, and `Message Content Intent`
6. Click on `Reset Token` and copy the token for your bot

Instead of creating a new application, you can also add any existing application to Kite. Just make sure to configure them correctly as described above.

## Adding the App to Kite

After creating your app in Discord, it's now time to add it to Kite.

Open the [Kite dashboard](https://kite.onl/apps) and sign-in with your Discord account. This is where you will add your app, manage commands, and more.

1. Click on `Create app` at the bottom
2. Fill in the bot token and click on `Create app`
3. Click on `Open app` next to the app's name

You have now added your app to Kite and are ready to create your first command!

## Invite the App

You can now invite your app to your Discord server. It usually makes sense to first test your app on a small Discord server before adding it to a real one with a lot of people.

Just click on `Invite app` at top-right of the dashboard's overview page and select the server that you want to add your app to.

## Creating Your First Command

1. Click on the `/` icon in the sidebar on the left side of the dashboard
2. Click on `Create command` and give your command a name and description
3. In the no-code editor drag in a `Create response message` block from the block explorer on the left side
4. Connect it to the command block by dragging your mouse from the blue dot of one block to the other
5. Click on the `Create response message` block and type in some text for the response
6. Save the command by clicking on `Save Changes` at the top

![Example Flow](./img/example-flow.png)

It's time to try it out inside Discord!

It can take up to a minute for your new command to appear inside Discord. Make sure you have invited the app to your server and restart your Discord client if the command doesn't appear!

## Creating Your First Message Template

1. Click on the envelope icon in the sidebar on the left side of the dashboard
2. Click on `Create message` and give your message a name and description
3. In the message editor type in some text, edit the embed, and design the message however you like
4. Save the message by clicking on `Save Changes` at the top
5. Click on `Send Message` to send the message to a channel that your app has access to

![Example Message](./img/example-message.png)

You can also use the newly created message template as a response to your previously created command!
