import { TemplateIconSvgs } from "@/components/app/TemplateIcon";
import { templateIconSvgs } from "@/lib/flow/templateIconSvgs";
import {
  TemplateInstallCount,
  TemplateListEntry,
} from "@/components/app/TemplateListEntry";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import MarketplaceBar from "@/components/home/MarketplaceBar";
import { fetchBundles } from "@/lib/apiData";
import { isExport } from "@/lib/env/client";
import {
  creatorBundles,
  creatorPath,
  templateFromBundle,
  templatePath,
} from "@/lib/flow/templates";
import { creatorImageVersion, ogImagePath } from "@/lib/og/path";
import { Bundle } from "@/lib/types/wire.gen";
import userAvatar from "@/tools/common/utils/discordCdn";
import { GetStaticPaths, GetStaticProps } from "next";
import { useMemo } from "react";

type Props = {
  creatorId: string;
  // The templates the user listed, newest first.
  listed: Bundle[];
  templateIconSvgs: TemplateIconSvgs;
};

// On kite.onl the page of a creator is rendered on its first visit. The export
// for self-hosting has no templates of users, so it has none.
export const getStaticPaths: GetStaticPaths = () => ({
  paths: [],
  fallback: isExport ? false : "blocking",
});

export const getStaticProps: GetStaticProps<Props> = async ({ params }) => {
  const creatorId = String(params?.creatorId);
  const bundles = await fetchBundles();
  const listed = creatorBundles(bundles ?? [], creatorId);
  // Users are only here while they have a template in the marketplace.
  if (!listed.length) {
    return { notFound: true, revalidate: 60 };
  }
  return {
    props: {
      creatorId,
      listed,
      templateIconSvgs: templateIconSvgs(listed.map((b) => b.listing?.icon)),
    },
    revalidate: 60,
  };
};

// CreatorPage lists the templates a user has in the marketplace.
export default function CreatorPage({ creatorId, listed }: Props) {
  const templates = useMemo(
    () => listed.flatMap((b) => templateFromBundle(b) ?? []),
    [listed]
  );
  const owner = listed[0].owner!;
  const installs = templates.reduce((sum, t) => sum + (t.installs ?? 0), 0);

  return (
    <HomeLayout
      title={`Discord Bot Templates by ${owner.display_name}`}
      description={`Free templates for Discord bots that ${
        owner.display_name
      } made with Kite: ${templates.map((t) => t.name).join(", ")}.`}
      image={ogImagePath("creators", creatorId, creatorImageVersion(templates))}
      embedButtons={[
        { label: "View templates", path: creatorPath(creatorId) },
        { label: "All templates", path: "/marketplace" },
      ]}
    >
      <MarketplaceBar />
      <div className="container max-w-5xl pt-12 pb-24 md:pb-32 min-h-[70vh]">
        <div className="flex gap-5 items-center mb-10">
          <img
            src={userAvatar({
              id: creatorId,
              avatar: owner.discord_avatar,
            })}
            alt=""
            className="h-16 w-16 md:h-20 md:w-20 rounded-full flex-none"
          />
          <div className="min-w-0">
            <h1 className="text-2xl md:text-3xl font-semibold">
              {owner.display_name}
            </h1>
            <div className="flex flex-wrap items-center gap-x-4 gap-y-1 mt-1.5 text-sm text-muted-foreground">
              <span>@{owner.discord_username}</span>
              <span>
                {templates.length}{" "}
                {templates.length === 1 ? "template" : "templates"}
              </span>
              <TemplateInstallCount installs={installs} />
            </div>
          </div>
        </div>

        <p className="text-sm text-muted-foreground mb-4">
          Made by a user of Kite, who is responsible for them. Kite reviews them
          before they&apos;re listed.
        </p>
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {templates.map((template) => (
            <TemplateListEntry
              key={template.id}
              template={template}
              href={templatePath(template.id)}
            />
          ))}
        </div>
      </div>

      <HomeCTASection light />
      <HomeFooter />
    </HomeLayout>
  );
}
