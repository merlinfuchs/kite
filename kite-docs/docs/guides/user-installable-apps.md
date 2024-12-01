---
sidebar_position: 2
---

# User Installable Apps

Most apps on Discord are installed by server admins to their servers and are only available on that server. However, some apps are designed to be installed by users to their own accounts which makes them available on any server the user is on. While these types of apps are less powerful, they can be useful for utility apps that are not server specific.

With Kite you can create user installable Discord apps in a few clicks without having to write a single line of code. In this guide, we will go through the steps to create a user installable app with Kite and add it to your account.

Some servers disallow the usage of user installable apps for moderation reasons. If you are experiencing issues with your app not working, please check the server's settings.

## Creating Your app

To create your Discord app, please follow the steps in the [getting started guide](/guides/getting-started).

## Configuring the Installation Contexts

Once your app is created, you need to configure the installation contexts. To do this, open your app in the [Discord Developer Portal](https://discord.com/developers/applications) and switch to the `Installation' section on the left side.

Under `Installation Contexts`, check the box for `User Install` and save the changes. You can optionally disable `Guild Install` if you don't want your app to be installable on servers.

![Developer Portal Installation](./img/devportal-installation.png)

## Adding the App to Your Account

Now you can already add the app to your Discord account by clicking on `Invite app` in the top right corner of the Kite dashboard. You can then select between installing the app to your account or a specific server. For this guide, we will install the app to our account.

Once the app is installed, you can use any commands of the app in any server you are in and also in private messages with the bot or other users.

## Configuring Commands

By default, all commands created with Kite are available both when the app is installed to a server and when it is installed to a user's account. If you want some commands to only be available when the app is installed to a server or only when it is installed to a user's account, you can configure that using the `Command Contexts` option node in the Kite Flow editor.

With the same node, you can also configure if your command should be available only in DMs or only on servers.

![Command Contexts](./img/example-cmd-contexts.png)
