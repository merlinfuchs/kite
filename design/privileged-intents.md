# Privileged intents for the Kite bot

Status: draft answers written 2026-11-19, gaps open, not submitted.

The Kite bot (application 844710218864656414) asks Discord for Server Members and Message Content, so the templates that need them can run on it. It doesn't ask for Presence. This follows Xenon's application, see `docs/privileged-intent-review.md` in the Xen repo, which has the form and what Discord's review guide asks for.

## Why we have to apply

Since June 2026, apps that reach fewer than 10,000 unique users turn privileged intents on in the Developer Portal. Bigger apps apply, and approved apps reapply every year. On 2026-11-19 the Kite bot was in 116 servers with 117,930 members in total. The biggest has 58,807. Even with overlap it's far over 10,000, so it applies.

Discord's guide asks for three things. The use case has to need the access and be impossible with interactions. Stored data needs a reason, a retention and security. Screenshots or a video have to show the feature working. Unclear or incomplete answers are a reason to reject.

## What each intent is for

| Intent | Template | Why interactions can't do it | What is stored |
| --- | --- | --- | --- |
| Message Content | AI Automod | It checks messages other people write, as they're sent or edited | Nothing. Deleted messages are copied into the server's own log channel on Discord |
| Message Content | Starboard | It reposts the text of a message that got enough ⭐ reactions, and reading a message over REST returns no text without the intent | The ID of the starboard post for each starred message, until the template is removed |
| Message Content | Counting | Members type the next number as a normal message, which no interaction fires for | The current number and the ID of who counted last, per channel, until the template is removed |
| Server Members | Welcomer | It needs member join and leave events | Nothing |

Every template is one we wrote. Servers can't build their own flows on the Kite bot, so these are the only things that read message text or member events. A server turns a template on in the dashboard, and only people with Manage Server can.

## Safeguards and where each one has to live

| Claim | Where it's enforced |
| --- | --- |
| Message text only reaches the templates of servers that turned on one that reads it | Not built yet, see Gaps |
| Member join and leave events only reach servers with Welcomer | Not built yet, see Gaps |
| Message text isn't stored or logged | Flow logs are errors without block inputs (`model.LogEntry`, `flow.ErrorDetails`). The AI log line in `engine/providers.go` only has token counts |
| Removing a template deletes what it stored | `HandleBundleInstallDelete` deletes the install with its variables |
| Server admins delete everything by deleting the app | App settings, `HandleAppDelete` |
| Anyone can ask for deletion | contact@kite.onl, privacy policy section 7 and "Your rights" |
| Not used to train models | Privacy policy section 8. OpenAI requests set `store: false`, OpenRouter and its providers are set up not to train |
| Encrypted at rest, in Germany | Encrypted disks at Hetzner and netcup, privacy policy section 4 |

## Gaps to close before submitting

**Message text reaches every app.** Intents are global, so once Discord grants Message Content the shared process gets the text of every message in all servers. The privacy policy already says the text is only read in servers with a template that needs it. The shared process has to drop the text before dispatching to apps without such a template, so the answer rests on code.

**Member events reach every app.** Same for joins and leaves, which only Welcomer needs.

**The Kite bot can't ask for the intents yet.** Its intents are a constant in `sharedbot/gateway.go`. They have to come from config, so prod can turn each one on the day Discord grants it.

**The four templates aren't on the Kite bot.** They need `shared_bot: true`, their blocks and events in `TestSharedBotBundles`, and a way to stay hidden on the setup page until the intent they need is on.

**Users can't opt out of message reading.** The form asks. The answer is no, and it says why. AI Automod is moderation, and letting the people it moderates opt out would defeat it. Starboard and Counting only act on messages in channels the server picked.

**The video has to show the bot.** Xenon's runbook says the video shows the bot under review. The Kite bot can't read message text until it's granted, so the videos are recorded on the test bot, which runs the same code, and the answer says so. Whether Discord accepts that is the main open question.

## Evidence to record

One unlisted video per intent, on the test bot in the Kite Test server, with a second account so messages and joins come from someone else. Captions in a bar above the recording, like Xenon's.

Message Content, in this order:

1. The setup page on kite.onl, turning on AI Automod, Starboard and Counting and picking their channels.
2. AI Automod. The second account posts a scam link. It's deleted and shows up in the log channel.
3. Starboard. Three ⭐ reactions on a message, and it appears in the starboard channel.
4. Counting. Counting to 5, then a wrong number, and the bot says who ruined it.

Server Members:

1. Turning on Welcomer and picking the channel.
2. The second account joins and is greeted, then leaves and the goodbye is posted.

Screenshots: the authorization screen with the privacy policy and terms links, and the Kite bot section of the privacy policy.

## The form

| Section | Field | Answer |
| --- | --- | --- |
| Application details | What does your application do? | The summary below |
| | Public privacy policy? | Yes |
| | Where is it available? | On the authorization screen when adding the bot, and in the footer of kite.onl |
| | Privacy policy link | https://kite.onl/privacy#kite-bot |
| Intents | Which intents? | Server Members and Message Content. Not Presence. |
| Server Members | Why do you need it? | The Server Members draft |
| | Screenshots or videos | The Server Members video |
| | Storing API data off-platform? | No |
| | How users request deletion | The deletion text below |
| | Encrypted at rest? | Yes |
| Message Content | Can users opt out? | No |
| | Storing it off-platform? | Yes, IDs and the current count only |
| | Stored for 30 days or less? | No |
| | How users request deletion | The deletion text below |
| | Encrypted at rest? | Yes |
| | Used to train ML or AI models? | No |
| | Why do you need it? | The Message Content draft |
| | Screenshots or videos | The Message Content video |
| Acknowledgement | Certify the answers | Check it last |

Check that the Kite bot's privacy policy and terms URLs are set in the Developer Portal, so Discord shows them on the authorization screen.

"What does your application do?":

> Kite (https://kite.onl) lets people build Discord bots without code. The Kite bot is Kite's own bot for servers that don't want to make one. A server admin adds it and turns on ready-made templates from a dashboard, like verification, leveling, giveaways, suggestions and moderation commands. We write every template. Servers only choose which ones run and set things like channels and roles, and they can't add their own logic to the Kite bot.
>
> Message Content is used by three templates: AI Automod, Starboard and Counting. Server Members is used by Welcomer. Both are explained below, with videos:
> Message Content: [video]
> Server Members: [video]

Deletion, for both intents:

> Server admins remove a template or delete their app in the Kite dashboard, which deletes everything it stored. Anyone can email contact@kite.onl to have their data deleted, and we answer within one month.

## Draft answers

### Message Content

> Three templates on the Kite bot read the text of messages. Each one acts on messages people write in the server, so slash commands can't replace them. A server only gets them when an admin turns them on in the dashboard.
>
> AI Automod checks new and edited messages against the rules the server wrote, with an AI model. It deletes spam, scams, ads and insults and posts a copy in the server's log channel for its moderators, or flags messages it isn't sure about for review. Keyword filters miss most of these, because scams are worded differently every time. The text is sent to the model for the check and isn't stored. We've set up our AI providers so they don't use it to train models.
>
> Starboard reposts a message to a starboard channel once it gets enough ⭐ reactions. The bot reads the message over REST when the reaction comes in, which returns no text without this intent. We store only the ID of the starboard post for each starred message, so the star count can be updated.
>
> Counting runs a counting game in one channel. Members type the next number, and the bot reads each message in that channel to check it. We store the current number and the ID of who counted last.
>
> Message text only reaches these templates, and only in servers that turned one of them on. It's kept in memory while the template runs and isn't stored or logged. Everything a template stores is deleted when the server removes it. People can't opt out, because AI Automod is moderation and Starboard and Counting only act in channels the server picked. Anyone can email contact@kite.onl to have stored data deleted.
>
> The videos are recorded on our test bot, which runs the same code. The Kite bot can't read message text until this intent is granted.

### Server Members

> Welcomer greets new members in a channel the server picks and says goodbye when they leave. It needs member join and leave events, which Discord only sends with this intent. Nothing about the member is stored. Member events only reach Welcomer, and only in servers that turned it on.

## Open questions

- Does Discord accept videos from a test bot for a first application?
- Should a join roles template go in with Welcomer? Discord's guide names welcome messages as a valid use, and join roles would be a second one.
