import Head from "next/head";

export default function BaseLayout({
  title,
  description,
  children,
}: {
  title?: string;
  description?: string;
  children: React.ReactNode;
}) {
  return (
    <>
      <Head>
        <title key="title">
          {`Kite | ${title ?? "Custom Discord Bots for Free"}`}
        </title>
        <meta
          name="description"
          key="description"
          content={
            description ||
            "Kite - Create Discord Bots for free with no coding required."
          }
        />
        <meta property="og:site_name" key="og:site_name" content="kite.onl" />
        <meta property="og:image" key="og:site_name" content="/logo.png" />
      </Head>

      {children}
    </>
  );
}
