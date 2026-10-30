import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQProps } from "@/components/home/HomeFAQSections";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import CookieApiPlayground from "@/components/landing/CookieApiPlayground";
import ShowcaseVisual from "@/components/landing/ShowcaseVisual";
import { noCodeFaq } from "@/components/landing/copy";
import { cookieApiTranscriptFlow } from "@/components/landing/flows";
import { Button } from "@/components/ui/button";
import env from "@/lib/env/client";
import { faqJsonLd } from "@/lib/seo";
import { CoinsIcon, LockIcon, PlugIcon, SparklesIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode } from "react";

const blocks = [
  "Transcripts",
  "Welcome cards",
  "Captcha",
  "Minecraft",
  "QR codes",
];

const cookieApiUrl = "https://www.cookie-api.com";

const steps: { title: string; description: ReactNode }[] = [
  {
    title: "Get Your API Key",
    description: (
      <>
        Sign up on{" "}
        <ExternalLink href={cookieApiUrl}>cookie-api.com</ExternalLink>, which
        has a free plan, and copy your key from the{" "}
        <ExternalLink href="https://dashboard.cookie-api.com">
          dashboard
        </ExternalLink>
        .
      </>
    ),
  },
  {
    title: "Enable Cookie API in Kite",
    description:
      "Open your app's Integrations and paste the key. Kite stores it encrypted and only sends it to Cookie API.",
  },
  {
    title: "Add the Blocks",
    description:
      "The Cookie API blocks are in the block explorer. Add them to a command or event like any other block.",
  },
];

const extras = [
  {
    icon: LockIcon,
    title: "No Token Pasting",
    description:
      "Transcripts need your bot's token. Kite adds it to the request itself, so it's never in a block.",
  },
  {
    icon: CoinsIcon,
    title: "1 Credit per Block",
    description:
      "Each Cookie API block costs 1 of your 10,000 free credits a month when it runs.",
  },
  {
    icon: SparklesIcon,
    title: "Let AI Build It",
    description:
      "Describe the command you want, like a /close command that saves a transcript, and Kite's AI builds it.",
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

function ExternalLink({
  href,
  children,
}: {
  href: string;
  children: ReactNode;
}) {
  return (
    <a
      href={href}
      target="_blank"
      rel="noopener"
      className="text-foreground underline decoration-primary underline-offset-4 hover:text-primary"
    >
      {children}
    </a>
  );
}

export default function CookieAPIPage() {
  return (
    <HomeLayout
      title="Use Cookie API in a Discord Bot Without Code"
      description="Use Cookie API in your own Discord bot without writing code. Save ticket transcripts, make welcome cards, verify members with a captcha and link Minecraft accounts with Kite."
      jsonLd={[faqJsonLd(faqs)]}
    >
      <section className="relative container grid grid-cols-1 lg:grid-cols-2 gap-12 items-center py-16 md:py-24">
        <div className="z-10 text-center lg:text-start">
          <h1 className="text-4xl md:text-6xl font-bold">
            Use{" "}
            <span className="bg-gradient-to-r from-[#f9ad15] to-primary text-transparent bg-clip-text">
              Cookie API
            </span>{" "}
            in Your Own Discord Bot
          </h1>
          <p className="text-xl text-muted-foreground mt-6">
            <ExternalLink href={cookieApiUrl}>Cookie API</ExternalLink> makes
            ticket transcripts, welcome cards, captchas and Minecraft
            verification for Discord communities. Enable it in Kite with your
            key, and they become blocks you add to your bot&apos;s commands, no
            code needed.
          </p>
          <div className="mt-8 flex flex-col md:flex-row justify-center lg:justify-start gap-4">
            <Button className="md:w-48" asChild>
              <Link href="/apps">Get Started for Free</Link>
            </Button>
            <Button variant="outline" className="md:w-48" asChild>
              <a href={`${env.NEXT_PUBLIC_DOCS_LINK}/guides/cookie-api`}>
                Read the Guide
              </a>
            </Button>
          </div>
          <div className="mt-8 flex flex-wrap items-center justify-center lg:justify-start gap-2 text-sm text-muted-foreground">
            <span>Blocks for</span>
            {blocks.map((block) => (
              <span
                key={block}
                className="rounded-md border bg-muted/50 px-2 py-0.5 font-medium text-foreground"
              >
                {block}
              </span>
            ))}
          </div>
        </div>

        <div className="z-10 min-w-0">
          <ShowcaseVisual
            showcase={{
              flow: cookieApiTranscriptFlow,
              command: "close",
              reply:
                "Transcript saved: https://www.cookie-api.com/transcripts/1/ticket-42.html",
            }}
          />
        </div>

        {/* Shadow effect */}
        <div className="shadow"></div>
      </section>

      <section className="container grid grid-cols-1 lg:grid-cols-[2fr_3fr] gap-12 items-center py-16 md:py-24">
        <div>
          <h2 className="text-3xl md:text-4xl font-bold">
            Set It Up in Two Minutes
          </h2>
          <ol className="relative mt-8 space-y-8">
            {/* Connects the numbers of the steps. */}
            <div className="absolute left-5 top-5 bottom-5 border-l-2 border-dashed" />
            {steps.map((step, i) => (
              <li key={step.title} className="relative flex gap-4">
                <div className="flex-none size-10 rounded-full bg-primary text-primary-foreground flex items-center justify-center font-bold">
                  {i + 1}
                </div>
                <div>
                  <h3 className="text-lg font-semibold">{step.title}</h3>
                  <p className="text-muted-foreground mt-1">
                    {step.description}
                  </p>
                </div>
              </li>
            ))}
          </ol>
        </div>
        <EnableMockup />
      </section>

      <section className="container py-16 md:py-24">
        <div className="text-center max-w-3xl mx-auto">
          <h2 className="text-3xl md:text-4xl font-bold">See the Blocks Run</h2>
          <p className="text-lg text-muted-foreground mt-4">
            Pick an example to see the flow in Kite and what your bot replies in
            Discord.
          </p>
        </div>
        <div className="mt-12 max-w-5xl mx-auto">
          <CookieApiPlayground />
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mt-12 max-w-5xl mx-auto">
          {extras.map((extra) => (
            <div key={extra.title} className="flex gap-4">
              <div className="flex-none size-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
                <extra.icon className="size-5" />
              </div>
              <div>
                <h3 className="font-semibold">{extra.title}</h3>
                <p className="text-muted-foreground mt-1">
                  {extra.description}
                </p>
              </div>
            </div>
          ))}
        </div>
      </section>

      <HomeFAQSection faqs={faqs} />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}

// A simplified dialog that enables Cookie API in an app's integrations.
function EnableMockup() {
  return (
    <div className="min-w-0 rounded-xl border bg-card p-6 shadow-sm">
      <div className="flex items-center gap-3">
        <div className="size-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
          <PlugIcon className="size-5" />
        </div>
        <div>
          <div className="font-semibold">Enable Cookie API</div>
          <div className="text-sm text-muted-foreground">
            Transcripts, cards, QR codes, captchas and Minecraft verification
          </div>
        </div>
      </div>
      <div className="mt-6 flex justify-between text-sm">
        <span className="font-medium">API key</span>
        <a
          href="https://docs.cookie-api.com/en/docs/getting-started/faq/api-key/"
          target="_blank"
          rel="noopener"
          className="text-primary hover:underline"
        >
          Where to find it
        </a>
      </div>
      <div className="mt-1.5 rounded-md border bg-background px-3 py-2 text-sm text-muted-foreground tracking-widest">
        sk_••••••••••••••••
      </div>
      <div className="mt-4 rounded-md bg-muted p-3 text-sm text-muted-foreground">
        Its blocks send Cookie API the settings you fill in. Create transcript
        also sends your bot&apos;s token, which Cookie API uses to read the
        channel&apos;s messages.
      </div>
      <div className="mt-5 flex justify-end">
        <div className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground">
          Enable
        </div>
      </div>
    </div>
  );
}
