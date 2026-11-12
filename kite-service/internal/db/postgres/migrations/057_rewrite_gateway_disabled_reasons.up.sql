-- Apps disabled before the gateway gave plain reasons for these errors have
-- the raw error. The service has to match the new text exactly, see
-- model.AppDisabledReasonInvalidToken and AppDisabledReasonDisallowedIntents.
UPDATE apps SET disabled_reason = 'Discord bot token is invalid, please update it'
WHERE NOT enabled AND disabled_reason LIKE '%4004 Authentication failed%';

UPDATE apps SET disabled_reason = 'Your bot needs privileged intents that are turned off. Turn them on under Bot in the Discord Developer Portal, then start the app again.'
WHERE NOT enabled AND disabled_reason LIKE '%4014 Disallowed intent%';
