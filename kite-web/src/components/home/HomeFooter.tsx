import logo from "@/assets/logo/orange@1024.png";
import Link from "next/link";
import env from "@/lib/env/client";

const botLinks = [
  { label: "Bot hosting", href: "/discord-bot-hosting" },
  { label: "Moderation bot", href: "/discord-moderation-bot" },
  { label: "Welcome bot", href: "/discord-welcome-bot" },
  { label: "Leveling bot", href: "/discord-leveling-bot" },
  { label: "AI bot", href: "/ai-discord-bot" },
  { label: "AI bot maker", href: "/ai-discord-bot-maker" },
  { label: "ER:LC bot", href: "/erlc-discord-bot" },
  { label: "MCP server", href: "/discord-mcp-server" },
  { label: "Cookie API bot", href: "/cookie-api" },
];

const compareLinks = [
  { label: "vs. BotGhost", href: "/alternatives/botghost" },
  { label: "vs. inventor.bot", href: "/alternatives/inventor" },
  { label: "vs. BotWiz", href: "/alternatives/botwiz" },
  { label: "vs. BDFD", href: "/alternatives/bdfd" },
  { label: "All comparisons", href: "/alternatives" },
];

export default function HomeFooter() {
  return (
    <footer id="footer">
      <div className="container">
        <hr />
      </div>

      <section className="container py-20 grid grid-cols-2 md:grid-cols-5 gap-x-12 gap-y-8">
        <div className="col-span-full">
          <Link
            rel="noreferrer noopener"
            href="/"
            className="font-bold text-xl flex items-center"
          >
            <img src={logo.src} alt="Kite Logo" className="h-10 w-10 mr-3" />
            <div>Kite</div>
          </Link>
        </div>

        <div className="flex flex-col gap-2">
          <h3 className="font-bold text-lg">Get in touch</h3>
          <div>
            <a
              rel="noreferrer noopener"
              href={env.NEXT_PUBLIC_GITHUB_LINK}
              target="_blank"
              className="opacity-60 hover:opacity-100"
            >
              GitHub
            </a>
          </div>

          <div>
            <a
              rel="noreferrer noopener"
              href={env.NEXT_PUBLIC_DISCORD_LINK}
              target="_blank"
              className="opacity-60 hover:opacity-100"
            >
              Discord
            </a>
          </div>

          <div>
            <a
              rel="noreferrer noopener"
              href={`mailto:${env.NEXT_PUBLIC_CONTACT_EMAIL}`}
              target="_blank"
              className="opacity-60 hover:opacity-100"
            >
              Email
            </a>
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <h3 className="font-bold text-lg">Discord bots</h3>
          {botLinks.map((link) => (
            <div key={link.href}>
              <Link href={link.href} className="opacity-60 hover:opacity-100">
                {link.label}
              </Link>
            </div>
          ))}
        </div>

        <div className="flex flex-col gap-2">
          <h3 className="font-bold text-lg">Compare</h3>
          {compareLinks.map((link) => (
            <div key={link.href}>
              <Link href={link.href} className="opacity-60 hover:opacity-100">
                {link.label}
              </Link>
            </div>
          ))}
        </div>

        <div className="flex flex-col gap-2">
          <h3 className="font-bold text-lg">Resources</h3>
          <div>
            <a
              href={env.NEXT_PUBLIC_DOCS_LINK}
              className="opacity-60 hover:opacity-100"
              target="_blank"
            >
              Documentation
            </a>
          </div>
          <div>
            <Link href="/features" className="opacity-60 hover:opacity-100">
              Features
            </Link>
          </div>
          <div>
            <Link href="/changelog" className="opacity-60 hover:opacity-100">
              What&apos;s new
            </Link>
          </div>
          <div>
            <Link href="/guides" className="opacity-60 hover:opacity-100">
              Guides
            </Link>
          </div>
          <div>
            <Link href="/tools" className="opacity-60 hover:opacity-100">
              Discord tools
            </Link>
          </div>
          <div>
            <Link href="/marketplace" className="opacity-60 hover:opacity-100">
              Bot templates
            </Link>
          </div>
        </div>

        <div className="flex flex-col gap-2">
          <h3 className="font-bold text-lg">Legal</h3>
          <div>
            <Link href="/terms" className="opacity-60 hover:opacity-100">
              Terms of Service
            </Link>
          </div>

          <div>
            <Link href="/privacy" className="opacity-60 hover:opacity-100">
              Privacy Policy
            </Link>
          </div>

          <div>
            <Link href="/refund" className="opacity-60 hover:opacity-100">
              Refund Policy
            </Link>
          </div>

          <div>
            <Link href="/imprint" className="opacity-60 hover:opacity-100">
              Legal Notice
            </Link>
          </div>
        </div>
      </section>

      <section className="container pb-14 text-center">
        <h3>
          Copyright &copy; {new Date().getFullYear()} made by{" "}
          <a
            target="_blank"
            rel="noreferrer noopener"
            href="https://merlinfuchs.com"
            className="text-primary transition-all border-primary hover:border-b-2"
          >
            Merlin Fuchs
          </a>{" "}
          and{" "}
          <a
            target="_blank"
            rel="noreferrer noopener"
            href="https://github.com/merlinfuchs/kite/graphs/contributors?all=1"
            className="text-primary transition-all border-primary hover:border-b-2"
          >
            Contributors
          </a>
        </h3>
      </section>
    </footer>
  );
}
