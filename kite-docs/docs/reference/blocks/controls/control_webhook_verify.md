---
sidebar_position: 38.5
---

import EmbedFlowNode from "../../../../src/components/EmbedFlowNode";
import NodeInfoExplorer from "../../../../src/components/NodeInfoExplorer";

# Verify webhook signature

<EmbedFlowNode type="control_webhook_verify" />

The `Verify webhook signature` block checks that a request to a [webhook event listener](../../event.md#webhook-event-listeners) really comes from the service that signs it, and continues with its **Valid** or **Invalid** branch. Put it first in the flow and answer invalid requests with a [Respond to webhook](../actions/action_webhook_response_create.md) block and status `401`. Some services, like ER:LC, only accept a webhook URL that does this.

Pick the service that sends the requests:

| Service           | Key                                                                                                                        |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------- |
| ER:LC             | None, ER:LC's public key is built in                                                                                       |
| GitHub            | The secret of the webhook                                                                                                  |
| Stripe            | The signing secret starting with `whsec_`                                                                                  |
| Slack             | The app's signing secret                                                                                                   |
| Discord           | None, your app's public key is used                                                                                        |
| Standard Webhooks | The signing secret starting with `whsec_`, or the public key starting with `whpk_`. Used by Svix, Clerk, Resend and others |

With Discord, use it for [webhook events](https://discord.com/developers/docs/events/webhook-events), like your app being installed. Discord sends a ping when you save the URL in the Developer Portal, which the flow has to answer with status `204` after verifying it. Don't use the URL as the Interactions Endpoint URL, Kite already receives your app's interactions.

Store a secret key as a [secret](../../secrets.md) and enter it as `{{secrets.NAME}}`. Requests whose timestamp is more than 5 minutes off are invalid for the services that sign one, so a request that got out can't be sent again later.

For another service, pick **Custom** and set the algorithm (HMAC with SHA-256, SHA-1 or SHA-512, or Ed25519), what the service signs, the header with the signature and its encoding. For example, a service that signs a timestamp header and the body with HMAC SHA-256 needs the signed message `{{webhook.headers['x-timestamp']}}.{{webhook.body}}` and the signature `{{webhook.headers['x-signature']}}`. Remove a prefix like `sha256=` with `{{ trimPrefix(webhook.headers['x-signature'], 'sha256=') }}`.

Anything wrong with the request, like a missing header, makes it invalid. The block only fails when its settings are wrong, like a key that isn't one. It costs no credits.

<NodeInfoExplorer type="control_webhook_verify" />
