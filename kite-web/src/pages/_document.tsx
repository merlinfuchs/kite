import { Html, Head, Main, NextScript } from "next/document";

export default function Document() {
  return (
    <Html lang="en">
      <Head>
        <link rel="manifest" href="/manifest.json" />
        <link rel="apple-touch-icon" href="/icons/apple-touch-icon.png" />
        <meta name="apple-mobile-web-app-title" content="Kite" />
        <meta name="apple-mobile-web-app-capable" content="yes" />
        <meta
          name="apple-mobile-web-app-status-bar-style"
          content="black-translucent"
        />
        <meta
          name="viewport"
          content="width=device-width, initial-scale=1, viewport-fit=cover"
        />
        <script src="https://app.lemonsqueezy.com/js/lemon.js" defer></script>
      </Head>
      <body>
        <div id="pwa-startup" aria-hidden="true">
          <img src="/icons/icon-192x192.png" alt="" />
          <span className="pwa-startup-spinner" />
        </div>
        <Main />
        <NextScript />
      </body>
    </Html>
  );
}
