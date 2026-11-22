import { badgeInfo } from "@/components/badges/Badges";
import { Template } from "@/lib/flow/templates";
import { BundleOwner } from "@/lib/types/wire.gen";
import userAvatar from "@/tools/common/utils/discordCdn";
import {
  drawableText,
  fetchAvatar,
  muted,
  Pill,
  plural,
  renderOgImage,
  svgDataUrl,
} from "./frame";

// The pills that fit next to the count, the others are summed up.
const shownTemplates = 3;

// The link preview of a creator's page, with their avatar, name and badges and
// the names of their templates.
export async function renderCreatorImage(
  owner: BundleOwner,
  templates: Template[]
) {
  const avatar = await fetchAvatar(
    userAvatar({ id: owner.discord_id, avatar: owner.discord_avatar }),
    256
  );
  const more = templates.length - shownTemplates;
  const name = drawableText(owner.display_name) || owner.discord_username;
  const badges = (owner.badges ?? []).flatMap((badge, i) => {
    const info = badgeInfo(badge);
    if (!info) return [];
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none">${info.svg(
      `badge${i}`
    )}</svg>`;
    return [{ title: info.title, icon: svgDataUrl(svg) }];
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
              fontSize: name.length > 20 ? 60 : 76,
              fontWeight: 700,
              lineHeight: 1.1,
              lineClamp: 1,
              display: "block",
            }}
          >
            {name}
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
                gap: 10,
                marginTop: 8,
                // What's left of the width next to the avatar, so they wrap.
                maxWidth: 840,
              }}
            >
              {badges.map((badge) => (
                <div
                  key={badge.title}
                  style={{
                    display: "flex",
                    alignItems: "center",
                    gap: 8,
                    padding: "6px 16px 6px 10px",
                    borderRadius: 999,
                    border: "2px solid #292524",
                    fontSize: 22,
                  }}
                >
                  <img src={badge.icon} width={28} height={28} alt="" />
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
            {drawableText(t.name)}
          </span>
        </Pill>
      )),
      more > 0 && <span key="more">{`+${more} more`}</span>,
    ],
  });
}
