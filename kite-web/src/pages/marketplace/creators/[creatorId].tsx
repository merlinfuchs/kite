import { TemplateIconSvgs } from "@/components/app/TemplateIcon";
import { templateIconSvgs } from "@/lib/flow/templateIconSvgs";
import {
  FeaturedTemplateEntry,
  TemplateListEntry,
} from "@/components/app/TemplateListEntry";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import MarketplaceBar from "@/components/home/MarketplaceBar";
import { Button } from "@/components/ui/button";
import { fetchBundles } from "@/lib/apiData";
import { isExport } from "@/lib/env/client";
import {
  categoryPath,
  creatorBundles,
  creatorPath,
  minShownStars,
  Template,
  templateCategories,
  templateFromBundle,
  templatePath,
} from "@/lib/flow/templates";
import { cn, formatNumber } from "@/lib/utils";
import { creatorImageVersion, ogImagePath } from "@/lib/og/path";
import { Bundle } from "@/lib/types/wire.gen";
import userAvatar from "@/tools/common/utils/discordCdn";
import { GetStaticPaths, GetStaticProps } from "next";
import { CheckIcon, LinkIcon } from "lucide-react";
import Link from "next/link";
import { useMemo, useState } from "react";

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
  const avatar = userAvatar({ id: creatorId, avatar: owner.discord_avatar });
  const installs = templates.reduce((sum, t) => sum + (t.installs ?? 0), 0);
  const stars = templates.reduce((sum, t) => sum + (t.stars ?? 0), 0);
  const byInstalls = [...templates].sort(
    (a, b) => (b.installs ?? 0) - (a.installs ?? 0)
  );
  const top = byInstalls.slice(0, 3);
  const more = byInstalls.slice(3);
  const categories = templateCategories.filter((c) =>
    templates.some((t) => t.category === c.value)
  );
  const since = creatorSince(templates);

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
      <section className="border-b bg-muted/30">
        <div className="container max-w-5xl pt-12 pb-10 md:pt-16">
          <div className="flex flex-col md:flex-row md:items-center gap-6">
            <div className="relative h-24 w-24 md:h-28 md:w-28 flex-none">
              {/* A glow in the colors of their avatar. */}
              <img
                src={avatar}
                alt=""
                className="absolute inset-0 h-full w-full rounded-full scale-125 blur-2xl opacity-50"
              />
              <img
                src={avatar}
                alt=""
                className="relative h-full w-full rounded-full ring-4 ring-background"
              />
            </div>
            <div className="min-w-0 flex-1">
              <h1 className="text-3xl md:text-4xl font-bold truncate">
                {owner.display_name}
              </h1>
              <div className="text-muted-foreground mt-1">
                @{owner.discord_username}
                {since && ` · Creator since ${since}`}
              </div>
            </div>
            <CopyLinkButton />
          </div>

          <div className="flex flex-wrap items-center gap-x-8 gap-y-4 mt-8">
            <Stat
              value={templates.length}
              label={templates.length === 1 ? "template" : "templates"}
            />
            <Stat
              value={installs}
              label={installs === 1 ? "bot uses them" : "bots use them"}
            />
            {stars >= minShownStars && <Stat value={stars} label="stars" />}
            <div className="flex flex-wrap gap-2">
              {categories.map((c) => (
                <Link
                  key={c.value}
                  href={categoryPath(c.value)}
                  className={cn(
                    "rounded-full px-3 py-1 text-sm font-medium transition-opacity hover:opacity-80",
                    c.tile,
                    c.icon
                  )}
                >
                  {c.label}
                </Link>
              ))}
            </div>
          </div>
        </div>
      </section>

      <div className="container max-w-5xl pt-10 pb-16 space-y-12">
        <section>
          <h2 className="text-xl font-semibold">
            {more.length ? "Most added" : "Templates"}
          </h2>
          <p className="text-sm text-muted-foreground mt-1 mb-4">
            Made by a user of Kite, who is responsible for them. Kite reviews
            them before they&apos;re listed.
          </p>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {top.map((template) => (
              <FeaturedTemplateEntry
                key={template.id}
                template={template}
                href={templatePath(template.id)}
              />
            ))}
          </div>
        </section>
        {!!more.length && (
          <section>
            <h2 className="text-xl font-semibold mb-4">More templates</h2>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {more.map((template) => (
                <TemplateListEntry
                  key={template.id}
                  template={template}
                  href={templatePath(template.id)}
                />
              ))}
            </div>
          </section>
        )}
      </div>

      <HomeCTASection light />
      <HomeFooter />
    </HomeLayout>
  );
}

function Stat({ value, label }: { value: number; label: string }) {
  return (
    <div>
      <div className="text-2xl font-bold">{formatNumber(value)}</div>
      <div className="text-sm text-muted-foreground">{label}</div>
    </div>
  );
}

// creatorSince is the month of their first listed template.
function creatorSince(templates: Template[]) {
  const first = templates
    .flatMap((t) => t.createdAt ?? [])
    .sort()
    .at(0);
  if (!first) return undefined;
  return new Date(first).toLocaleDateString("en-US", {
    month: "long",
    year: "numeric",
    timeZone: "UTC",
  });
}

function CopyLinkButton() {
  const [copied, setCopied] = useState(false);

  return (
    <Button
      variant="outline"
      className="self-start md:self-center bg-background/60"
      onClick={() => {
        navigator.clipboard.writeText(window.location.href);
        setCopied(true);
        setTimeout(() => setCopied(false), 2000);
      }}
    >
      {copied ? (
        <CheckIcon className="mr-2 h-4 w-4" />
      ) : (
        <LinkIcon className="mr-2 h-4 w-4" />
      )}
      {copied ? "Copied" : "Copy link"}
    </Button>
  );
}
