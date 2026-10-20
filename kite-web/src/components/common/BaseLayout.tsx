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
  discordEmbed = true,
  children,
}: {
  title?: string;
  description?: string;
  noindex?: boolean;
  jsonLd?: object[];
  /** Set to false to fall back to the plain Open Graph preview in Discord. */
  discordEmbed?: boolean;
  children: React.ReactNode;
}) {
  const router = useRouter();

  const pageTitle = title ?? "Free Discord Bot Maker, No Coding Required";
  const fullTitle = `Kite | ${pageTitle}`;
  const fullDescription =
    description ||
    "Make your own Discord bot for free with Kite. Build slash commands, buttons and events in a visual editor, no coding required. Hosted 24/7.";
  const canonicalUrl = env.NEXT_PUBLIC_SITE_URL + router.pathname;
  const imageUrl = `${env.NEXT_PUBLIC_SITE_URL}/og.png`;

  return (
    <>
      <Head>
        <title key="title">{fullTitle}</title>
        <meta name="description" key="description" content={fullDescription} />
        {noindex ? (
          <meta name="robots" key="robots" content="noindex" />
        ) : (
          <link rel="canonical" key="canonical" href={canonicalUrl} />
        )}
        <meta property="og:site_name" key="og:site_name" content="Kite" />
        <meta property="og:type" key="og:type" content="website" />
        <meta property="og:title" key="og:title" content={fullTitle} />
        <meta
          property="og:description"
          key="og:description"
          content={fullDescription}
        />
        <meta property="og:url" key="og:url" content={canonicalUrl} />
        <meta property="og:image" key="og:image" content={imageUrl} />
        <meta
          name="twitter:card"
          key="twitter:card"
          content="summary_large_image"
        />
        <meta name="theme-color" key="theme-color" content="#f97316" />
        {discordEmbed && (
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
