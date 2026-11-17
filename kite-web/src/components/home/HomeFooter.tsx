import orangeLogo from "@/assets/logo/orange@scalable.svg";
import DiscordIcon from "@/components/icons/DiscordIcon";
import env from "@/lib/env/client";
import { MailIcon } from "lucide-react";
import Link from "next/link";
import HomeSectionDivider from "./HomeSectionDivider";

const botLinks = [
  { label: "Bot hosting", href: "/discord-bot-hosting" },
  { label: "Moderation bot", href: "/discord-moderation-bot" },
  { label: "Welcome bot", href: "/discord-welcome-bot" },
  { label: "Leveling bot", href: "/discord-leveling-bot" },
  { label: "AI bot", href: "/ai-discord-bot" },
  { label: "AI bot maker", href: "/ai-discord-bot-maker" },
  { label: "ER:LC bot", href: "/erlc-discord-bot" },
  { label: "Roblox ranking bot", href: "/roblox-ranking-bot" },
  { label: "MCP server", href: "/discord-mcp-server" },
  { label: "Cookie API bot", href: "/cookie-api" },
];

const compareLinks = [
  { label: "vs. BotGhost", href: "/alternatives/botghost" },
  { label: "vs. inventor.bot", href: "/alternatives/inventor" },
  { label: "vs. BotWiz", href: "/alternatives/botwiz" },
  { label: "vs. BDFD", href: "/alternatives/bdfd" },
  { label: "Best free bot makers", href: "/alternatives" },
];

const resourceLinks = [
  { label: "Documentation", href: env.NEXT_PUBLIC_DOCS_LINK, external: true },
  { label: "Features", href: "/features" },
  { label: "What's new", href: "/changelog" },
  { label: "Guides", href: "/guides" },
  { label: "Discord tools", href: "/tools" },
  { label: "Bot templates", href: "/marketplace" },
];

const legalLinks = [
  { label: "Terms of Service", href: "/terms" },
  { label: "Privacy Policy", href: "/privacy" },
  { label: "Refund Policy", href: "/refund" },
  { label: "Legal Notice", href: "/imprint" },
];

export default function HomeFooter() {
  return (
    <footer id="footer">
      <HomeSectionDivider
        from="background"
        to="muted"
        variant="horizon"
        accent="primary"
      />

      <div className="relative overflow-hidden bg-muted/30">
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-0 bg-[radial-gradient(90%_75%_at_50%_105%,hsl(var(--primary)/0.05),transparent_70%)] dark:bg-[radial-gradient(90%_85%_at_50%_105%,hsl(var(--primary)/0.24),transparent_72%)]"
        />
        <div
          aria-hidden="true"
          className="pointer-events-none absolute inset-x-0 bottom-0 h-[55%] bg-gradient-to-t from-muted/70 dark:from-primary/[0.12] to-transparent"
        />

        <section className="container relative z-10 py-14 sm:py-16 grid grid-cols-1 lg:grid-cols-12 gap-10 lg:gap-12">
          <div className="lg:col-span-4 flex flex-col justify-between gap-8">
            <div className="space-y-4">
              <Link
                href="/"
                className="inline-flex items-center gap-3 font-bold text-xl"
              >
                <img src={orangeLogo.src} alt="Kite Logo" className="h-9 w-9" />
                <span>Kite</span>
              </Link>

              <p className="text-sm text-muted-foreground max-w-xs leading-relaxed">
                Free Discord bot maker with slash commands, events, and AI,
                hosted 24/7.
              </p>

              <div className="flex items-center gap-2 pt-1">
                <a
                  rel="noreferrer noopener"
                  href={env.NEXT_PUBLIC_GITHUB_LINK}
                  target="_blank"
                  aria-label="GitHub"
                  className="flex h-9 w-9 items-center justify-center rounded-lg border bg-background/70 text-muted-foreground hover:text-foreground hover:border-primary/40 transition-colors"
                >
                  <svg
                    viewBox="0 0 24 24"
                    fill="currentColor"
                    className="h-4 w-4"
                    aria-hidden="true"
                  >
                    <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0 1 12 6.844a9.59 9.59 0 0 1 2.504.337c1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.02 10.02 0 0 0 22 12.017C22 6.484 17.522 2 12 2Z" />
                  </svg>
                </a>

                <a
                  rel="noreferrer noopener"
                  href={env.NEXT_PUBLIC_DISCORD_LINK}
                  target="_blank"
                  aria-label="Discord"
                  className="flex h-9 w-9 items-center justify-center rounded-lg border bg-background/70 text-muted-foreground hover:text-foreground hover:border-primary/40 transition-colors"
                >
                  <DiscordIcon className="h-4 w-4" />
                </a>

                <a
                  rel="noreferrer noopener"
                  href={`mailto:${env.NEXT_PUBLIC_CONTACT_EMAIL}`}
                  target="_blank"
                  aria-label="Email"
                  className="flex h-9 w-9 items-center justify-center rounded-lg border bg-background/70 text-muted-foreground hover:text-foreground hover:border-primary/40 transition-colors"
                >
                  <MailIcon className="h-4 w-4" />
                </a>
              </div>
            </div>

            <p className="text-xs text-muted-foreground">
              Copyright &copy; {new Date().getFullYear()} made by{" "}
              <a
                target="_blank"
                rel="noopener"
                href="https://merlinfuchs.com"
                className="text-primary hover:underline"
              >
                Merlin Fuchs
              </a>{" "}
              and{" "}
              <a
                target="_blank"
                rel="noreferrer noopener"
                href="https://github.com/merlinfuchs/kite/graphs/contributors?all=1"
                className="text-primary hover:underline"
              >
                contributors
              </a>
            </p>
          </div>

          <div className="lg:col-span-8 grid grid-cols-2 sm:grid-cols-4 gap-x-8 gap-y-10">
            <div className="flex flex-col gap-2">
              <h3 className="text-sm font-semibold mb-1">Discord bots</h3>
              {botLinks.map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                >
                  {link.label}
                </Link>
              ))}
            </div>

            <div className="flex flex-col gap-2">
              <h3 className="text-sm font-semibold mb-1">Compare</h3>
              {compareLinks.map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                >
                  {link.label}
                </Link>
              ))}
            </div>

            <div className="flex flex-col gap-2">
              <h3 className="text-sm font-semibold mb-1">Resources</h3>
              {resourceLinks.map((link) =>
                link.external ? (
                  <a
                    key={link.href}
                    href={link.href}
                    target="_blank"
                    rel="noreferrer noopener"
                    className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                  >
                    {link.label}
                  </a>
                ) : (
                  <Link
                    key={link.href}
                    href={link.href}
                    className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                  >
                    {link.label}
                  </Link>
                )
              )}
            </div>

            <div className="flex flex-col gap-2">
              <h3 className="text-sm font-semibold mb-1">Legal</h3>
              {legalLinks.map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  className="text-sm text-muted-foreground hover:text-foreground transition-colors"
                >
                  {link.label}
                </Link>
              ))}
            </div>
          </div>
        </section>
      </div>
    </footer>
  );
}
