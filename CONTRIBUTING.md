# Contributing to Kite

Thanks for helping out! To keep reviews fast, please follow these rules. PRs that don't will get closed or sent back without a detailed review.

## Before you start

- For new features, UI changes or anything bigger than a small fix, open an issue first and wait for a go-ahead. Nobody wants you to spend a weekend on something that won't be merged.
- Keep each PR to one feature or fix. Seven new blocks are seven PRs, or at least one PR per closely related group.
- Check that the feature doesn't already exist. Many things can be done with existing blocks.

## Writing the code

[AGENTS.md](./AGENTS.md) describes the repo layout, the checks, and step-by-step checklists for adding blocks and event listeners. Read it, and if you use an AI coding tool, make sure it reads it too.

Using AI tools is fine, but you're responsible for the result. You should be able to explain every line of your PR. If you can't, it's not ready.

## Before opening a PR

- All checks in [AGENTS.md](./AGENTS.md#checks) pass locally. CI runs the same checks and a PR with failing CI won't be reviewed.
- Generated files (`*.gen.ts`, `pgmodel`, `catalog.json`) were regenerated, not edited.
- New blocks have a docs page.
- The diff contains nothing unrelated to the change.
- The PR description says what the change does and includes a screenshot or video of it working.

If a review asks for changes and there's no update for two weeks, the PR will be closed. You can always reopen it once it's ready.
