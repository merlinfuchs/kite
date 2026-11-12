// earlyCreatorCutoff is the day the badges launched. Creators who listed a
// template in the Marketplace before it, without knowing about the badges, are
// early creators.
export const earlyCreatorCutoff = "2026-11-13";

// creatorMilestones are how many other users must have added any of a
// creator's templates for each badge, highest first. 10 is also what the
// creator role in the Kite server needs.
export const creatorMilestones = [1000, 100, 10];

export type CreatorBadge =
  | { kind: "early" }
  | { kind: "milestone"; installers: number };

// creatorBadges are the badges of a creator with the templates they have in
// the Marketplace. firstListedAt is when the oldest of them was made.
export function creatorBadges({
  installerCount,
  firstListedAt,
}: {
  installerCount: number;
  firstListedAt?: string;
}): CreatorBadge[] {
  const badges: CreatorBadge[] = [];
  if (firstListedAt && firstListedAt < earlyCreatorCutoff) {
    badges.push({ kind: "early" });
  }
  const milestone = creatorMilestones.find((m) => installerCount >= m);
  if (milestone) {
    badges.push({ kind: "milestone", installers: milestone });
  }
  return badges;
}
