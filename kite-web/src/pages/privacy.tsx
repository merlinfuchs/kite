import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { formatLegalDate, legalUpdatedAt } from "@/lib/legal";
import Link from "next/link";

export default function PrivacyPolicyPage() {
  return (
    <HomeLayout title="Privacy Policy" noindex>
      <div className="prose dark:prose-invert text-foreground mx-auto my-16 max-w-4xl px-5">
        <h1>Privacy Policy</h1>
        <p>Last updated: {formatLegalDate(legalUpdatedAt)}</p>
        <p>
          Kite is a no-code Discord bot builder. This policy explains which
          personal data we process when you use kite.onl and when Kite runs a
          bot, why we process it, and how long we keep it.
        </p>

        <h2 id="controller">1. Who is responsible</h2>
        <p>
          Merlin Fuchs
          <br />
          Alte Str. 5, 04229 Leipzig, Germany
          <br />
          <a href="mailto:contact@kite.onl">contact@kite.onl</a>
        </p>
        <p>
          For the data of your Kite account and your use of kite.onl, we are the
          controller. For the data of people on Discord that a bot built with
          Kite processes, the owner of that bot is the controller, and we
          process the data on their behalf (see{" "}
          <a href="#bot-data">section 6</a>).
        </p>

        <h2 id="website">2. Visiting kite.onl</h2>
        <p>
          When you open kite.onl, our server receives your IP address, the page
          you request, the time, your browser&apos;s user agent and the
          referring page. We need this to deliver the website and to protect it
          against attacks (Art. 6(1)(f) GDPR). Server logs are deleted after 14
          days. Logs of our services contain technical data like app and Discord
          IDs, and are deleted after 30 days.
        </p>
        <p>
          kite.onl, its API and its docs are delivered through Cloudflare, which
          protects them against attacks. Every request passes through
          Cloudflare&apos;s network, including your IP address and what you send
          to Kite (Art. 6(1)(f) GDPR).
        </p>
        <p>
          Avatars and emojis from Discord are loaded directly from
          Discord&apos;s servers, which receive your IP address. Our bot
          building guide has a video from YouTube (youtube-nocookie.com). It
          only loads when you click play, and YouTube then receives your IP
          address.
        </p>
        <p>
          We measure how kite.onl and the dashboard are used with OpenPanel,
          which we run ourselves on a server in Germany. It records page views
          and actions like creating a command or using the flow AI, together
          with your browser, device type and the country derived from your IP
          address. It doesn&apos;t use cookies, and it doesn&apos;t keep your IP
          address. When you are logged in, we link these events to your account
          and send OpenPanel your user ID, Discord username and display name, so
          we can see where people get stuck and improve Kite (Art. 6(1)(f)
          GDPR). You can object at any time by emailing us.
        </p>

        <h2 id="storage">3. Cookies and local storage</h2>
        <p>
          When you log in, we set a session cookie that keeps you logged in. It
          expires after 7 days without use. Your browser&apos;s local storage
          keeps settings like your theme, the position of the flow editor and
          the notices you dismissed. Both are necessary for the features you use
          (&sect; 25(2) no. 2 TDDDG) and never leave your device except the
          session cookie, which is sent to our API.
        </p>

        <h2 id="account">4. Your account and apps</h2>
        <p>
          You log in with Discord. Discord gives us your Discord user ID,
          username, display name, avatar and email address. We use them to
          create your account, show who you are to collaborators of your apps,
          and contact you about your account and important changes. We keep them
          until you delete your account.
        </p>
        <p>
          For your apps, we store what you create in Kite: the bot token, name,
          description and avatar of the bot, your commands, event listeners,
          messages, variables, images, secrets, integration settings and the
          collaborators you invite. Everything is stored on encrypted disks in
          Germany, and bot tokens and secrets are encrypted once more. Images
          are stored encrypted with an object storage provider in Germany. This
          data is kept until you delete the app.
        </p>
        <p>While an app runs, we also keep records with shorter lifetimes:</p>
        <ul>
          <li>logs of your flows: 30 days;</li>
          <li>credit usage records: 40 days;</li>
          <li>
            waiting flows, like timers and buttons that continue a flow: up to
            90 days;
          </li>
          <li>
            references to messages your bot sent, so it can edit them later: 90
            days, or 360 days for messages sent from the dashboard;
          </li>
          <li>share codes: 90 days.</li>
        </ul>
        <p>
          When you ask the flow AI to edit a flow, we store your prompt, the
          model and the number of tokens it used, to count it towards your
          plan&apos;s limit. Prompts are deleted with the app.
        </p>
        <p>
          We process this data to provide Kite to you (Art. 6(1)(b) GDPR).
          Encrypted backups of the database are stored with an object storage
          provider in Germany and overwritten within five weeks.
        </p>

        <h2 id="templates">5. Templates and the marketplace</h2>
        <p>
          Anyone with the link to a template you make can see it, together with
          your Discord display name, username, avatar and user ID. Content you
          share with a share code is visible to everyone who has the code.
        </p>
        <p>
          When you list a template in the marketplace, it is public and search
          engines can index it. Its page shows your Discord display name,
          username and avatar, and links to a creator page at
          kite.onl/marketplace/creators/ followed by your Discord user ID, which
          lists your templates in the marketplace. You choose to list a
          template, so the legal basis is Art. 6(1)(b) GDPR. When you remove
          your last template from the marketplace, your creator page goes away.
          Search engines may still show it for a while.
        </p>
        <p>
          When you submit a template, we keep the submitted version and our
          decision, including the reason if we reject it. When you report a
          template, we keep your account, the reason and the details you wrote.
          Only the people who review reports see them, and we don&apos;t tell
          the creator who reported their template. We keep reviews and reports
          until the template is deleted, and reports also until you delete your
          account. We process them to moderate the marketplace, as the EU
          Digital Services Act requires (Art. 6(1)(c) and (f) GDPR).
        </p>

        <h2 id="bot-data">6. Data processed by bots built with Kite</h2>
        <p>
          Bots built with Kite receive data from Discord, like the user IDs,
          names, roles and messages of the people in the servers they are in,
          and run the flows their owner built. They may store this data in
          variables and logs or send it to other services, depending on how the
          bot is built. Kite processes this data only on behalf of the
          bot&apos;s owner, who decides what it is used for (Art. 28 GDPR). Our{" "}
          <Link href="/terms#data-processing">Terms of Service</Link> contain
          the data processing agreement with bot owners.
        </p>
        <p>
          If you have questions about how a bot uses your data, or want it
          deleted, contact the bot&apos;s owner. If you can&apos;t reach them,
          email us and we&apos;ll forward your request. When a bot&apos;s owner
          deletes the app in Kite, we delete its data.
        </p>
        <p>
          Flows can send HTTP requests to services their owner chooses. These
          requests go through a proxy server we run ourselves in Germany.
        </p>

        <h2 id="ai">7. AI features</h2>
        <p>
          Some features send data to AI models: the flow AI that edits flows, AI
          blocks in flows, and the support bot in our Discord server.
        </p>
        <ul>
          <li>
            The flow AI sends your prompt, the earlier messages of the
            conversation, the flow you are editing and the names of your
            app&apos;s variables to OpenAI.
          </li>
          <li>
            AI blocks send what their flow puts into the prompt to OpenAI. The
            AI Choice, Check and Score blocks send it to OpenRouter, which
            forwards it to the provider of the model. Which data that is depends
            on how the bot is built, and it can include Discord messages.
          </li>
          <li>
            The support bot sends your question, and the earlier questions and
            answers of the conversation, to OpenAI. We don&apos;t store them.
          </li>
        </ul>
        <p>
          We tell OpenAI not to store these requests, and we&apos;ve set up
          OpenAI and OpenRouter so that they and the model providers don&apos;t
          use them to train models. Providers may keep requests for a limited
          time to detect abuse. For the flow AI and support bot, the legal basis
          is Art. 6(1)(b) GDPR. For AI blocks, it is the instruction of the
          bot&apos;s owner (see <a href="#bot-data">section 6</a>).
        </p>

        <h2 id="payments">8. Payments</h2>
        <p>
          Subscriptions are sold by Lemon Squeezy as merchant of record. Lemon
          Squeezy processes your payment and billing details as its own
          controller under its{" "}
          <a
            href="https://www.lemonsqueezy.com/privacy"
            target="_blank"
            rel="noreferrer"
          >
            privacy policy
          </a>
          . We never see your payment details. We send Lemon Squeezy your user
          ID and the app the subscription is for, and receive the status and
          plan of your subscription, to give your app the features it paid for
          (Art. 6(1)(b) GDPR). If you are in our Discord server, your Discord
          account gets a role for your plan there.
        </p>

        <h2 id="support">9. Support</h2>
        <p>
          When you email us, we use your email address and message to answer you
          (Art. 6(1)(b) or (f) GDPR) and delete them when they are no longer
          needed. Our support server on Discord is subject to{" "}
          <a
            href="https://discord.com/privacy"
            target="_blank"
            rel="noreferrer"
          >
            Discord&apos;s privacy policy
          </a>
          .
        </p>

        <h2 id="recipients">10. Who we share data with</h2>
        <p>
          We don&apos;t sell personal data or use it for advertising. We only
          share it with these providers, who process it for us:
        </p>
        <ul>
          <li>
            Hosting providers in Germany, currently Hetzner Online GmbH and
            netcup GmbH: servers, database and storage.
          </li>
          <li>
            Cloudflare, Inc., USA: delivers kite.onl, its API and its docs and
            protects them against attacks.
          </li>
          <li>OpenAI, USA: flow AI, AI blocks and the support bot.</li>
          <li>
            OpenRouter, USA, and the providers of the models it routes to: the
            AI Choice, Check and Score blocks.
          </li>
        </ul>
        <p>
          Lemon Squeezy (USA) processes payment data as its own controller, see{" "}
          <a href="#payments">section 8</a>. Bots built with Kite also send data
          to Discord and to the services their owner chooses. We disclose data
          to authorities only when the law requires it.
        </p>
        <p>
          Transfers to the USA are based on the EU-US Data Privacy Framework
          where the provider is certified under it, and otherwise on the
          European Commission&apos;s Standard Contractual Clauses (Art. 46(2)(c)
          GDPR).
        </p>

        <h2 id="rights">11. Your rights</h2>
        <p>Under the GDPR, you have the right to:</p>
        <ul>
          <li>get a copy of the data we have about you (Art. 15);</li>
          <li>have wrong data corrected (Art. 16);</li>
          <li>have your data deleted (Art. 17);</li>
          <li>have its processing restricted (Art. 18);</li>
          <li>receive your data in a machine-readable format (Art. 20);</li>
          <li>
            object to processing based on Art. 6(1)(f) GDPR, for reasons that
            arise from your situation (Art. 21);
          </li>
          <li>
            complain to a data protection authority (Art. 77), for example ours:
            Sächsische Datenschutz- und Transparenzbeauftragte, Devrientstraße
            5, 01067 Dresden, Germany.
          </li>
        </ul>
        <p>
          You can delete apps yourself in their settings, which deletes their
          data, including the bot token. For everything else, including deleting
          your account, email us at{" "}
          <a href="mailto:contact@kite.onl">contact@kite.onl</a>. We answer
          within one month.
        </p>

        <h2 id="other">12. Other information</h2>
        <p>
          Kite is only for people who are at least 18 years old. We don&apos;t
          make decisions about you based solely on automated processing.
          Providing your data is required to use Kite, but you aren&apos;t
          legally obliged to use it.
        </p>
        <p>
          We&apos;ll update this policy when we change how we process data. The
          current version is always on this page.
        </p>
      </div>

      <HomeFooter />
    </HomeLayout>
  );
}
