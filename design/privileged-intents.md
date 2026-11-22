# Privileged intents for the Kite bot

Status: draft answers written 2026-11-19. Gaps closed in code, videos not recorded, not submitted.

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

AI Automod leads the Message Content answer and video. A counting bot's application for the intent was declined about a year ago as not convincing or unique enough, so Counting is named in the answer, because it reads the text, but only in one sentence at the end. Starboard comes in between.

Every template is one we wrote. Servers can't build their own flows on the Kite bot, so these are the only things that read message text or member events. A server turns a template on in the dashboard, and only people with Manage Server can.

## Safeguards and where each one lives

| Claim | Where it's enforced |
| --- | --- |
| Only AI Automod, Starboard and Counting read message text | `intents` on each template, which `TestBundleIntents` checks against its flows |
| Only Welcomer uses member events | The same test, for `guild_members` |
| Templates that need an intent can't be added before it's granted | `SyncOfficial` stores them without `shared_bot` until `shared_bot.intents` lists the intent |
| Message text isn't stored or logged | Flow logs are errors without block inputs (`model.LogEntry`, `flow.ErrorDetails`). The AI log line in `engine/providers.go` only has token counts |
| Removing a template deletes what it stored | `HandleBundleInstallDelete` deletes the install with its variables |
| Server admins delete everything by deleting the app | App settings, `HandleAppDelete` |
| Anyone can ask for deletion | contact@kite.onl, privacy policy section 7 and "Your rights" |
| Not used to train models, and not kept by the AI provider | Privacy policy section 8. AI Automod's checks go through OpenRouter's decisions API to `typesafe/jev-1.14`. OpenRouter's account has zero data retention on and every training toggle off, and TypeSafe's endpoint is on its zero data retention list |
| The AI provider is bound in writing, as section 12(a) of the Developer Terms asks | OpenRouter's terms (section 10.2) incorporate its data processing agreement, under which it processes only on our instructions (section 2.3). Its model providers are bound by their own no-retention, no-training policies, which OpenRouter checks for the list but doesn't guarantee |
| Encrypted at rest, in Germany | Encrypted disks at Hetzner and netcup, privacy policy section 4 |

## Gaps

**Message text reaches every app's flows.** Intents are global, so once Discord grants Message Content, every app on the Kite bot gets message events with their text, like any bot with the intent. Only the templates that declare it read the text, which the test checks. Filtering the text out for the other apps was built and dropped. Text stays in Discord and with people who can read it there, as the Kite bot has no HTTP blocks and stored variables stay in the server. So the answers say which templates read the text, not that it only reaches them.

**Asking for the intents.** `shared_bot.intents` in the config lists the granted ones, and the bot asks for those. It's empty until Discord approves.

**Users can't opt out of message reading.** The form asks. The answer is no, and it says why. AI Automod is moderation, and letting the people it moderates opt out would defeat it. Starboard and Counting only act on messages in channels the server picked.

**The video has to show the bot.** Xenon's runbook says the video shows the bot under review. The Kite bot can't read message text until it's granted, so the videos are recorded on the test bot, which runs the same code, and the answer says so. Whether Discord accepts that is the main open question.

## Evidence to record

One unlisted video per intent, on the test bot in the Kite Test server, with a second account so messages and joins come from someone else. Captions in a bar above the recording, like Xenon's.

Message Content, in this order:

1. The setup page on kite.onl, turning on AI Automod with its log channel and a rule of the server's own, like no links to other servers.
2. AI Automod, most of the video. The second account posts a scam that keyword filters miss, like a "free nitro" link worded differently. It's deleted and shows up in the log channel with who sent it and where. Then a message that breaks the server's own rule. Then a borderline message, which stays up and goes to the log channel for a moderator to decide. Then a normal message, which is left alone.
3. Starboard. Three ⭐ reactions on a message, and it appears in the starboard channel.
4. Counting, a few seconds. Counting to 3 and a wrong number.

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
> Message Content is mainly used by AI Automod, which moderates messages with AI, and also by Starboard and Counting. Server Members is used by Welcomer. Both are explained below, with videos:
> Message Content: [video]
> Server Members: [video]

Deletion, for both intents:

> Server admins remove a template or delete their app in the Kite dashboard, which deletes everything it stored. Anyone can email contact@kite.onl to have their data deleted, and we answer within one month.

## Draft answers

### Message Content

> We need Message Content mainly for AI Automod, a moderation template on the Kite bot. Servers turn it on in the Kite dashboard to keep spam, scams, ads and insults out of their channels.
>
> AI Automod checks every new and edited message with an AI model, against a default set of rules and the rules the server adds, like no links to other servers. It deletes a message only when the model is at least 90% sure it breaks a rule, and posts a copy in the server's log channel with who sent it and where. Messages it's less sure about stay up and go to the log channel with a link, so a moderator decides. Members who can manage messages are never checked, and servers can limit the checks to members who joined recently, where most spam bots and raiders are.
>
> This can't be done with interactions, because it acts on messages members write, not on commands. Discord's AutoMod matches keywords and patterns, and scams are worded differently every time. AI Automod catches the ones that get past it. The text is sent to our AI provider, OpenRouter, only for the check. OpenRouter processes it on our instructions under its data processing agreement, with zero data retention, and only routes it to a model provider that keeps nothing and doesn't train on it. We don't store or log it either.
>
> Two smaller templates read message text too. Starboard reposts a message that got enough ⭐ reactions to a starboard channel. It reads the message when the reaction comes in, and we store only the ID of the starboard post. Counting checks the numbers members post in one counting channel, and we store the current number and who counted last.
>
> Only these templates read message text, and only in servers that turned one of them on. It's kept in memory while the template runs and isn't stored or logged. Everything a template stores is deleted when the server removes it. People can't opt out, because AI Automod is moderation and Starboard and Counting only act in channels the server picked. Anyone can email contact@kite.onl to have stored data deleted.
>
> The videos are recorded on our test bot, which runs the same code. The Kite bot can't read message text until this intent is granted.

### Server Members

> Welcomer greets new members in a channel the server picks and says goodbye when they leave. It needs member join and leave events, which Discord only sends with this intent. Nothing about the member is stored. Only Welcomer uses member events, and only in servers that turned it on.

## Before submitting

- Keep the decision models on endpoints in OpenRouter's zero data retention list. Bumping `aiDecisionModel` or `screenModel` to a version that isn't on it makes OpenRouter reject the requests.
- Prompt logging stays off in the OpenRouter account. With it on, OpenRouter may use inputs for its own purposes (section 6.2 of its terms), which section 12(a) of Discord's Developer Terms rules out.

## Open questions

- Does Discord accept videos from a test bot for a first application?
- Should a join roles template go in with Welcomer? Discord's guide names welcome messages as a valid use, and join roles would be a second one.
