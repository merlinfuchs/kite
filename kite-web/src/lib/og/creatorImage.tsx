import { Template } from "@/lib/flow/templates";
import { BundleOwner } from "@/lib/types/wire.gen";
import userAvatar from "@/tools/common/utils/discordCdn";
import { fetchAvatar, muted, Pill, plural, renderOgImage } from "./frame";

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
