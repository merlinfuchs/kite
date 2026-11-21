import { Template } from "@/lib/flow/templates";
import { BundleOwner, UserBadge } from "@/lib/types/wire.gen";
import userAvatar from "@/tools/common/utils/discordCdn";
import { formatNumber } from "@/lib/utils";
import {
  badgeIcons,
  CreatorTier,
  StarTier,
} from "@/components/badges/badgeIcons";
import {
  fetchAvatar,
  muted,
  Pill,
  plural,
  renderOgImage,
  svgDataUrl,
} from "./frame";

// The pills that fit next to the count, the others are summed up.
const shownTemplates = 3;

// The link preview of a creator's page, with their avatar and name and the
// names of their templates.
export async function renderCreatorImage(
  owner: BundleOwner,
  templates: Template[]
) {
  const avatar = await fetchAvatar(
    userAvatar({ id: owner.discord_id, avatar: owner.discord_avatar }),
    256
  );
  const more = templates.length - shownTemplates;
  const badges = (owner.badges ?? []).flatMap((badge, index) => {
    const info = creatorBadgeInfo(badge);
    if (!info) return [];
    const prefix = `creatorbadge${index}`;
    return [
      {
        ...info,
        icon: svgDataUrl(
          `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">${info.svg(
            prefix
          )}</svg>`
        ),
      },
    ];
  });

  return renderOgImage({
    label: "creator",
    body: [
      <div
        key="creator"
        style={{ display: "flex", alignItems: "center", gap: 40 }}
      >
        {avatar && (
          <img
            src={avatar}
            width={160}
            height={160}
            style={{ borderRadius: 999, flexShrink: 0 }}
            alt=""
          />
        )}
        <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
          <div
            style={{
              fontSize: owner.display_name.length > 20 ? 60 : 76,
              fontWeight: 700,
              lineHeight: 1.1,
              lineClamp: 1,
              display: "block",
            }}
          >
            {owner.display_name}
          </div>
          <div style={{ display: "flex", fontSize: 34, color: muted }}>
            {`@${owner.discord_username} · ${plural(
              templates.length,
              "template"
            )}`}
          </div>
          {!!badges.length && (
            <div
              style={{
                display: "flex",
                flexWrap: "wrap",
                gap: 8,
                marginTop: 4,
              }}
            >
              {badges.map((badge) => (
                <div
                  key={badge.title}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 8,
                    padding: "4px 12px",
                    borderRadius: 999,
                    border: "2px solid #292524",
                    fontSize: 18,
                  }}
                >
                  <img src={badge.icon} width={24} height={24} alt="" />
                  {badge.title}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>,
    ],
    footer: [
      ...templates.slice(0, shownTemplates).map((t) => (
        <Pill key={t.id}>
          <span style={{ lineClamp: 1, display: "block", maxWidth: 300 }}>
            {t.name}
          </span>
        </Pill>
      )),
      more > 0 && <span key="more">{`+${more} more`}</span>,
    ],
  });
}

function creatorBadgeInfo(badge: UserBadge) {
  switch (badge.kind) {
    case "staff":
      return { title: "Kite staff", svg: badgeIcons.staff };
    case "contributor":
      return { title: "Contributor", svg: badgeIcons.contributor };
    case "early_supporter":
      return { title: "Early supporter", svg: badgeIcons.early_supporter };
    case "subscriber":
      return {
        title: badge.plan ? `${badge.plan} subscriber` : "Subscriber",
        svg: badgeIcons.subscriber,
      };
    case "early_creator":
      return { title: "Early creator", svg: badgeIcons.early_creator };
    case "creator": {
      const tier = (badge.milestone ?? 0) as CreatorTier;
      if (![0, 10, 100, 1000].includes(tier)) return undefined;
      return {
        title: tier
          ? `Template creator · ${formatNumber(tier)}+`
          : "Template creator",
        svg: (prefix: string) => badgeIcons.creator(prefix, tier),
      };
    }
    case "star": {
      const tier = (badge.milestone ?? 0) as StarTier;
      if (![5, 20, 50].includes(tier)) return undefined;
      return {
        title: `Star creator · ${formatNumber(tier)}+`,
        svg: (prefix: string) => badgeIcons.star(prefix, tier),
      };
    }
  }
}
