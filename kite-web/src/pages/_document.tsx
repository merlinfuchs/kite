import { Html, Head, Main, NextScript } from "next/document";

export default function Document() {
  return (
    <Html lang="en">
      <Head>
        <link rel="manifest" href="/manifest.json" />
        <meta
          name="viewport"
          content="width=device-width, initial-scale=1, viewport-fit=cover"
        />
        <meta name="theme-color" content="#141211" />
        <meta name="apple-mobile-web-app-capable" content="yes" />
        <meta
          name="apple-mobile-web-app-status-bar-style"
          content="black-translucent"
        />
        <meta name="apple-mobile-web-app-title" content="Kite" />
        <link rel="apple-touch-icon" href="/icons/icon-192x192.png" />
        <script
          dangerouslySetInnerHTML={{
            __html: `!function(){try{var c=localStorage.getItem("theme");var isDark=c==="dark"||(!c&&window.matchMedia("(prefers-color-scheme: dark)").matches);var color=isDark?"#141211":"#FCFCFB";var m=document.querySelector('meta[name="theme-color"]');if(m)m.setAttribute("content",color);var s=document.querySelector('meta[name="apple-mobile-web-app-status-bar-style"]');if(s)s.setAttribute("content",isDark?"black-translucent":"default");}catch(e){}}();`,
          }}
        />
        <script src="https://app.lemonsqueezy.com/js/lemon.js" defer></script>
      </Head>
      <body>
        <Main />
        <NextScript />
      </body>
    </Html>
  );
}
