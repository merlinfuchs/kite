import LandingPage, {
  LandingPageContent,
} from "@/components/landing/LandingPage";
import { noCodeFaq } from "@/components/landing/copy";
import { cookieApiTranscriptFlow } from "@/components/landing/flows";
import env from "@/lib/env/client";

const content: LandingPageContent = {
  title: "Use Cookie API in a Discord Bot Without Code",
  description:
    "Use Cookie API in your own Discord bot without writing code. Save ticket transcripts, make welcome cards, verify members with a captcha and link Minecraft accounts with Kite.",
  heading: "Use Cookie API in Your Own Discord Bot",
  intro:
    "Kite builds Discord bots from blocks. Enable Cookie API with your key, and its transcripts, cards, captchas and Minecraft verification become blocks you add to your commands. Your bot keeps your own name and avatar.",
  secondaryAction: {
    label: "Read the Guide",
    href: `${env.NEXT_PUBLIC_DOCS_LINK}/guides/cookie-api`,
  },
  showcase: {
    flow: cookieApiTranscriptFlow,
    command: "close",
    reply:
      "Transcript saved: https://www.cookie-api.com/transcripts/1/ticket-42.html",
  },
  sections: [
    {
      heading: "Cookie API as Blocks",
      body: "Enable Cookie API under Integrations with the key from your Cookie API dashboard. Each block costs 1 credit when it runs.",
      points: [
        {
          title: "Create Transcript",
          description:
            "Save up to the last 500 messages of a channel as a web page, with a password if you want.",
        },
        {
          title: "Generate Card",
          description:
            "Turn a design from Cookie API's Card Builder into an image, with the member's name and avatar filled in.",
        },
        {
          title: "Create and Check Captcha",
          description:
            "Send a captcha page from Cloudflare or Google, then check if the member solved it.",
        },
        {
          title: "Verify Minecraft Player",
          description:
            "Get the Java or Bedrock player behind a code from Cookie API's verification server.",
        },
        {
          title: "Generate QR Code",
          description:
            "Make a QR code for a link, with squares or dots and your logo in the middle.",
        },
      ],
    },
    {
      heading: "Build the Bot Your Server Needs",
      points: [
        {
          title: "Ticket Transcripts",
          description:
            "Save a ticket when staff close it, post the link to a log channel and delete the channel.",
        },
        {
          title: "Welcome Cards",
          description:
            "Greet every new member with a card that shows their name and avatar.",
        },
        {
          title: "Captcha Verification",
          description:
            "Give new members a role only after they solve a captcha, to keep bots out.",
        },
        {
          title: "Minecraft Linking",
          description:
            "Let members link their Minecraft account and set it as their nickname.",
        },
        {
          title: "No Token Pasting",
          description:
            "Transcripts need your bot's token, which Kite adds to the request itself instead of you typing it into a block.",
        },
        {
          title: "Let AI Build It",
          description:
            "Describe the command you want, like a /close command that saves a transcript, and Kite's AI builds it.",
        },
      ],
    },
  ],
  faqs: [
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
  ],
};

export default function CookieAPIPage() {
  return <LandingPage content={content} />;
}
