import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQProps } from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import CookieApiCardDemo from "@/components/landing/CookieApiCardDemo";
import CookieApiPerks from "@/components/landing/CookieApiPerks";
import CookieApiPlayground from "@/components/landing/CookieApiPlayground";
import { CookieShape } from "@/components/landing/CookieApiVisuals";
import FloatingCookies from "@/components/landing/FloatingCookies";
import Reveal from "@/components/landing/Reveal";
import { noCodeFaq } from "@/components/landing/copy";
import { Button } from "@/components/ui/button";
import env from "@/lib/env/client";
import { faqJsonLd } from "@/lib/seo";
import {
  IdCardIcon,
  LucideIcon,
  PickaxeIcon,
  QrCodeIcon,
  ScrollTextIcon,
  ShieldCheckIcon,
} from "lucide-react";
import Link from "next/link";

// Cookie API is a partner, so its page is more playful than the other
// landing pages, like Cookie API's own site.

const blocks: { title: string; description: string; icon: LucideIcon }[] = [
  {
    title: "Create Transcript",
    description:
      "Save up to the last 500 messages of a channel as a web page, with a password if you want.",
    icon: ScrollTextIcon,
  },
  {
    title: "Generate Card",
    description:
      "Turn a design from Cookie API's Card Builder into an image, with the member's name and avatar filled in.",
    icon: IdCardIcon,
  },
  {
    title: "Create and Check Captcha",
    description:
      "Send a captcha page from Cloudflare or Google, then check if the member solved it.",
    icon: ShieldCheckIcon,
  },
  {
    title: "Verify Minecraft Player",
    description:
      "Get the Java or Bedrock player behind a code from Cookie API's verification server.",
    icon: PickaxeIcon,
  },
  {
    title: "Generate QR Code",
    description:
      "Make a QR code for a link, with squares or dots and your logo in the middle.",
    icon: QrCodeIcon,
  },
];

const faqs: FAQProps[] = [
  {
    question: "Do I need a Cookie API account?",
    answer:
      "Yes. Sign up on cookie-api.com, which has a free plan, and copy your API key from its dashboard. You enter it under Integrations in your app, where Kite stores it encrypted and only sends it to Cookie API.",
  },
  {
    question: "What does it cost on Kite?",
    answer:
      "Building and hosting your bot on Kite is free. Each Cookie API block costs 1 of your 10,000 free credits per month when it runs.",
  },
  {
    question: "Why does the transcript block send my bot's token?",
    answer:
      "Cookie API reads the channel's messages itself, so it needs your bot's token. Kite adds the token to that one request, so it's never written in your flow or visible to anyone who sees it.",
  },
  {
    question: "Who do I ask for help?",
    answer:
      "Ask in Kite's Discord server for help with your bot, and in Cookie API's Discord server for your Cookie API account and key.",
  },
  noCodeFaq,
];

export default function CookieAPIPage() {
  return (
    <HomeLayout
      title="Use Cookie API in a Discord Bot Without Code"
      description="Use Cookie API in your own Discord bot without writing code. Save ticket transcripts, make welcome cards, verify members with a captcha and link Minecraft accounts with Kite."
      jsonLd={[faqJsonLd(faqs)]}
    >
      <section className="relative">
        <FloatingCookies />
        <div className="container relative pt-16 pb-12 md:pt-24 text-center pointer-events-none">
          <div className="inline-flex items-center gap-2 rounded-full border bg-muted/50 px-3 py-1 text-sm text-muted-foreground">
            <CookieShape id="badge" className="h-4 w-4" />
            Kite and Cookie API
          </div>
          <h1 className="text-4xl md:text-6xl font-bold mt-6 max-w-4xl mx-auto leading-tight">
            Use{" "}
            <span className="inline-block -rotate-2 rounded-xl bg-[#f9ad15] px-3 text-black pointer-events-auto hover:animate-wiggle">
              Cookie API
            </span>{" "}
            in Your Own Discord Bot
          </h1>
          <p className="text-lg md:text-xl text-muted-foreground mt-6 max-w-2xl mx-auto">
            Kite builds Discord bots from blocks. Enable Cookie API with your
            key, and its transcripts, cards, captchas and Minecraft verification
            become blocks you add to your commands.
          </p>
          <div className="mt-8 flex flex-col md:flex-row justify-center gap-4 pointer-events-auto">
            <Button className="md:w-48" asChild>
              <Link href="/apps">Get Started for Free</Link>
            </Button>
            <Button variant="outline" className="md:w-48" asChild>
              <a href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/cookie-api`}>
                Read the Guide
              </a>
            </Button>
          </div>
        </div>
        <div className="container relative max-w-5xl pb-16 md:pb-24">
          <Reveal>
            <CookieApiPlayground />
          </Reveal>
        </div>
      </section>

      <section className="container py-12 md:py-16">
        <Reveal>
          <h2 className="text-3xl md:text-4xl font-bold text-center">
            Cookie API as Blocks
          </h2>
          <p className="text-lg text-muted-foreground mt-4 max-w-3xl mx-auto text-center">
            Enable Cookie API under Integrations with the key from your Cookie
            API dashboard. Each block costs 1 credit when it runs.
          </p>
        </Reveal>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-10">
          {blocks.map((block, i) => (
            <Reveal key={block.title} delay={i * 80}>
              <div className="group h-full rounded-2xl border bg-muted/50 p-6 transition duration-300 hover:-translate-y-1 hover:border-[#f9ad15]/60 hover:shadow-xl hover:shadow-[#f9ad15]/10">
                <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-[#f9ad15]/15 text-[#b45309] dark:text-[#f9ad15] transition-transform duration-300 group-hover:-rotate-12 group-hover:scale-110">
                  <block.icon className="h-6 w-6" />
                </div>
                <h3 className="text-xl font-semibold mt-4">{block.title}</h3>
                <p className="text-muted-foreground mt-2">
                  {block.description}
                </p>
              </div>
            </Reveal>
          ))}
        </div>
      </section>

      <section className="container py-12 md:py-20">
        <Reveal>
          <CookieApiCardDemo />
        </Reveal>
      </section>

      <section className="container py-12 md:py-16">
        <Reveal>
          <CookieApiPerks />
        </Reveal>
      </section>

      <HomeFAQSection faqs={faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
