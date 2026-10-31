import Head from "next/head";
import { useRouter } from "next/router";
import env from "@/lib/env/client";
import {
  buildKiteComponentEmbed,
  serializeComponentEmbed,
} from "@/lib/discordComponentEmbed";

export default function BaseLayout({
  title,
  description,
  noindex,
  jsonLd,
  image,
  embedButtons,
  discordEmbed = true,
  children,
}: {
  title?: string;
  description?: string;
  /** Path of the link preview image, like og.png by default. */
  image?: string;
  noindex?: boolean;
  jsonLd?: object[];
  /** Set to false to fall back to the plain Open Graph preview in Discord. */
  discordEmbed?: boolean;
  /** Buttons of the Discord preview, with paths instead of URLs. */
  embedButtons?: { label: string; path: string }[];
  children: React.ReactNode;
}) {
  const router = useRouter();

  const pageTitle = title ?? "Free Discord Bot Maker, No Coding Required";
  const fullTitle = `Kite | ${pageTitle}`;
  const fullDescription =
    description ||
    "Make your own Discord bot for free with Kite. Build slash commands, buttons and events in a visual editor, no coding required. Hosted 24/7.";
  // A page exported from a dynamic route like /marketplace/[templateId] gets its
  // own URL. One exported as a single shell for every ID, like
  // /bundles/[bundleId], doesn't know its ID until it runs in the browser,
  // which crawlers don't do, so it gets no URL and they use the one they
  // were given.
  const pathParams = router.pathname.match(/\[\w+\]/g) ?? [];
  const canonicalUrl = pathParams.every((p) => router.query[p.slice(1, -1)])
    ? env.NEXT_PUBLIC_SITE_URL +
      router.pathname.replace(/\[(\w+)\]/g, (_, key) =>
        String(router.query[key])
      )
    : undefined;
  const imageUrl = env.NEXT_PUBLIC_SITE_URL + (image ?? "/og.png");

  return (
    <>
      <Head>
        <title key="title">{fullTitle}</title>
        <meta name="description" key="description" content={fullDescription} />
        {noindex ? (
          <meta name="robots" key="robots" content="noindex" />
        ) : (
          canonicalUrl && (
            <link rel="canonical" key="canonical" href={canonicalUrl} />
          )
        )}
        <meta property="og:site_name" key="og:site_name" content="Kite" />
        <meta property="og:type" key="og:type" content="website" />
        <meta property="og:title" key="og:title" content={fullTitle} />
        <meta
          property="og:description"
          key="og:description"
          content={fullDescription}
        />
        {canonicalUrl && (
          <meta property="og:url" key="og:url" content={canonicalUrl} />
        )}
        <meta property="og:image" key="og:image" content={imageUrl} />
        <meta
          name="twitter:card"
          key="twitter:card"
          content="summary_large_image"
        />
        <meta name="theme-color" key="theme-color" content="#f97316" />
        {discordEmbed && canonicalUrl && (
          <script
            key="discord:component-embed"
            id="discord:component-embed"
            type="application/json"
            dangerouslySetInnerHTML={{
              __html: serializeComponentEmbed(
                buildKiteComponentEmbed({
                  title: pageTitle,
                  description: fullDescription,
                  url: canonicalUrl,
                  imageUrl,
                  buttons: embedButtons?.map((b) => ({
                    label: b.label,
                    url: env.NEXT_PUBLIC_SITE_URL + b.path,
                  })),
                })
              ),
            }}
          />
        )}
        {jsonLd?.map((data, i) => (
          <script
            key={`jsonld-${i}`}
            type="application/ld+json"
            dangerouslySetInnerHTML={{
              __html: JSON.stringify(data).replace(/</g, "\\u003c"),
            }}
          />
        ))}
      </Head>

      {children}
    </>
  );
}
