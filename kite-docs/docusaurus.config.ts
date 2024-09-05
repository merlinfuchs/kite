import { themes as prismThemes } from "prism-react-renderer";
import type { Config } from "@docusaurus/types";
import type * as Preset from "@docusaurus/preset-classic";

const config: Config = {
  title: "Kite",
  tagline: "Discord Bots made easy",
  favicon: "img/favicon.ico",

  // Set the production url of your site here
  url: "https://docs.kite.onl",
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: "/",

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: "merlinfuchs", // Usually your GitHub org/user name.
  projectName: "kite", // Usually your repo name.

  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "classic",
      {
        docs: {
          sidebarPath: "./sidebars.ts",
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          routeBasePath: "/",
          editUrl: "https://github.com/merlinfuchs/kite/tree/main/kite-docs/",
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl: "https://github.com/merlinfuchs/kite/tree/main/kite-docs/",
        },
        theme: {
          customCss: "./src/css/custom.css",
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    image: "img/logo.png",
    navbar: {
      title: "Kite",
      logo: {
        alt: "Kite Logo",
        src: "img/logo.svg",
      },
      items: [
        {
          type: "docSidebar",
          sidebarId: "tutorialSidebar",
          position: "left",
          label: "Documentation",
        },
        { to: "/blog", label: "Blog", position: "left" },
        {
          href: "https://github.com/merlinfuchs/kite",
          label: "GitHub",
          position: "right",
        },
        {
          href: "https://discord.gg/rNd9jWHnXh",
          label: "Discord",
          position: "right",
        },
      ],
    },
    footer: {
      style: "dark",
      links: [
        {
          title: "Docs",
          items: [
            {
              label: "Welcome",
              to: "/",
            },
            {
              label: "Getting Started",
              to: "/guides/getting-started",
            },
          ],
        },
        {
          title: "Community",
          items: [
            {
              label: "Discord",
              href: "https://discord.gg/rNd9jWHnXh",
            },
            {
              label: "GitHub",
              href: "https://github.com/merlinfuchs/kite",
            },
          ],
        },
        {
          title: "Legal",
          items: [
            {
              label: "Terms of Service",
              href: "https://kite.onl/terms",
            },
            {
              label: "Privacy Policy",
              href: "https://kite.onl/privacy",
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} made by Merlin Fuchs`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
