import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { formatLegalDate, legalUpdatedAt } from "@/lib/legal";
import Link from "next/link";

export default function TermsOfServicePage() {
  return (
    <HomeLayout title="Terms of Service" noindex>
      <div className="prose dark:prose-invert text-foreground mx-auto my-16 max-w-4xl px-5">
        <h1>Terms of Service</h1>
        <p>Last updated: {formatLegalDate(legalUpdatedAt)}</p>
        <p>
          These terms are an agreement between you and Merlin Fuchs, Alte Str.
          5, 04229 Leipzig, Germany (&quot;Kite&quot;, &quot;we&quot;,
          &quot;us&quot;). They cover kite.onl, the dashboard, the API and the
          bots Kite runs for you (together, the &quot;Service&quot;). The{" "}
          <Link href="/privacy">Privacy Policy</Link> explains which data we
          process and why.
        </p>

        <h2 id="account">1. Your account</h2>
        <p>
          You must be at least 18 years old to use Kite. If you use Kite for an
          organization, you confirm that you may agree to these terms for it.
        </p>
        <p>
          You log in with your Discord account. You&apos;re responsible for what
          happens through your account and for the apps you create.
          Collaborators you invite to an app can change it, so only invite
          people you trust.
        </p>

        <h2 id="bots">2. Your bots</h2>
        <p>
          Every app in Kite is a Discord application that you create in the
          Discord Developer Portal and own. You give Kite its bot token so Kite
          can run the bot for you. Towards Discord, you stay the developer of
          the application.
        </p>
        <p>
          You must follow the{" "}
          <a href="https://discord.com/terms" target="_blank" rel="noreferrer">
            Discord Terms of Service
          </a>
          ,{" "}
          <a
            href="https://discord.com/guidelines"
            target="_blank"
            rel="noreferrer"
          >
            Community Guidelines
          </a>
          ,{" "}
          <a
            href="https://support-dev.discord.com/hc/en-us/articles/8562894815383-Discord-Developer-Terms-of-Service"
            target="_blank"
            rel="noreferrer"
          >
            Developer Terms of Service
          </a>{" "}
          and{" "}
          <a
            href="https://support-dev.discord.com/hc/en-us/articles/8563934450327-Discord-Developer-Policy"
            target="_blank"
            rel="noreferrer"
          >
            Developer Policy
          </a>
          . Among other things, Discord requires your bot to have a privacy
          policy and to give its users a way to report problems. You&apos;re
          responsible for what your bot does, for the content of your flows and
          messages, and for following the laws that apply to your bot.
        </p>
        <p>
          Kite isn&apos;t affiliated with or endorsed by Discord. Discord can
          take action against your application on its own, and we have no
          influence on that.
        </p>

        <h2 id="service-provider">3. Kite as your service provider</h2>
        <p>
          Under the Discord Developer Terms of Service, Kite is a service
          provider that runs your application on your behalf. By adding a bot
          token to Kite, you instruct us to use it for you, and we agree to:
        </p>
        <ul>
          <li>
            only access and use the Discord APIs, your bot token and the data we
            get through them (&quot;API Data&quot;) for you, at your direction,
            to run and develop your application in compliance with the Discord
            Developer Terms of Service, and for no other purpose, including our
            own;
          </li>
          <li>
            keep your bot token encrypted and never show it to anyone, including
            you and other collaborators of your app;
          </li>
          <li>
            follow the Discord Developer Terms of Service and Developer Policy
            as if we were in your place;
          </li>
          <li>
            when you delete your app or stop using Kite, immediately stop using
            the Discord APIs for it and promptly delete the API Data we have for
            it, as described in the Privacy Policy.
          </li>
        </ul>
        <p>
          Your instructions are the flows, messages and settings you create in
          Kite for your app, and the actions you and your collaborators take for
          it in the dashboard or through our API. This section is our written
          agreement as your service provider under section 12(a) of the Discord
          Developer Terms of Service.
        </p>

        <h2 id="data-processing">4. Data processing on your behalf</h2>
        <p>
          When your bot runs, Kite processes personal data of people on Discord,
          like their user IDs, names, roles and messages, and anything your
          flows store or send elsewhere. For this data, you are the controller
          and Kite is your processor under Art. 28 GDPR. This section is our
          data processing agreement.
        </p>
        <ul>
          <li>
            Subject, purpose and duration: running your bot as you build it in
            Kite, for as long as the app exists in Kite.
          </li>
          <li>
            Data and data subjects: data about Discord users, server members,
            servers, channels, roles and messages that Discord sends to your bot
            or that your flows request, store in variables, write to logs or
            send to other services.
          </li>
          <li>
            We only process this data on your instructions, which are your
            flows, messages and settings in Kite and the actions taken for your
            app in the dashboard or through our API. If we think an instruction
            breaks data protection law, we&apos;ll tell you.
          </li>
          <li>Everyone who can access the data is bound to confidentiality.</li>
          <li>
            We protect the data with appropriate technical and organizational
            measures under Art. 32 GDPR. Among them: servers in Germany with
            encrypted disks, encrypted connections and backups, bot tokens and
            secrets encrypted once more, and access limited to the people who
            run Kite.
          </li>
          <li>
            You allow us to use the subprocessors listed in the{" "}
            <Link href="/privacy#recipients">Privacy Policy</Link>. We bind them
            to the same data protection obligations. We add new subprocessors to
            that list before we use them, and you can object by contacting us.
            If we can&apos;t resolve the objection, you can stop using Kite for
            the affected app.
          </li>
          <li>
            We help you answer requests of data subjects and meet your
            obligations under Art. 32 to 36 GDPR, as far as we can.
          </li>
          <li>
            We tell you without undue delay if we become aware of a breach that
            affects this data, and we notify Discord where the Discord Developer
            Terms of Service require it.
          </li>
          <li>
            When you delete your app, we delete its data as described in the
            Privacy Policy, unless the law requires us to keep it.
          </li>
          <li>
            We give you the information you need to check that we follow this
            section, and we allow and contribute to audits, normally by
            answering your questions in writing.
          </li>
        </ul>

        <h2 id="ai">5. AI features</h2>
        <p>
          Some features use AI models, like the AI blocks in flows and the
          assistant that edits flows. The data they need is sent to the AI
          providers listed in the Privacy Policy. AI output can be wrong,
          incomplete or inappropriate. Check it before you rely on it, and
          don&apos;t use AI features to create content that is illegal or breaks
          Discord&apos;s rules.
        </p>

        <h2 id="acceptable-use">6. Acceptable use</h2>
        <p>You must not use Kite to:</p>
        <ul>
          <li>break the law or the rules of Discord;</li>
          <li>
            spam, raid servers, send unsolicited direct messages, or inflate
            member counts or activity;
          </li>
          <li>
            harass, threaten or deceive people, including phishing, scams and
            impersonation;
          </li>
          <li>
            collect, scrape or track data about Discord users beyond what your
            bot needs for its features;
          </li>
          <li>
            run user accounts (&quot;self-bots&quot;) or ask people for their
            Discord passwords or user tokens;
          </li>
          <li>
            attack, overload or probe Kite, Discord or other services, including
            through HTTP requests from flows;
          </li>
          <li>
            get around plan limits or credits, for example with several accounts
            or apps.
          </li>
        </ul>
        <p>
          If you break these terms, we can disable a flow or bot, remove
          content, or suspend your account. We&apos;ll usually warn you first,
          unless the violation is serious or we have to act right away to
          protect people, Discord or Kite.
        </p>

        <h2 id="plans">7. Plans, credits and payment</h2>
        <p>
          Kite has a free plan and paid plans, which are described on the{" "}
          <Link href="/pricing">pricing page</Link>. Running blocks uses
          credits, and each plan includes a number of credits per month. When an
          app runs out of credits, its bot pauses until it has credits again.
        </p>
        <p>
          Payments are handled by Lemon Squeezy, which sells the subscription to
          you as merchant of record. Its terms apply to the purchase. You can
          cancel a subscription at any time, and it then ends with the current
          billing period. Refunds are covered by our{" "}
          <Link href="/refund">Refund Policy</Link>.
        </p>
        <p>
          We may change plans, including their credits, limits and features, and
          prices. Changes to the free plan apply when we make them. Changes to a
          paid subscription you already have only apply from its next billing
          period, and we&apos;ll tell you at least two weeks in advance, so you
          can cancel before.
        </p>

        <h2 id="content">8. Your content and templates</h2>
        <p>
          You keep all rights to the flows, messages, images and other content
          you add to Kite. You give us the rights we need to store, run and show
          it to provide the Service to you and your collaborators.
        </p>
        <p>
          Anyone with the link to a template you make can see it and add it to
          their apps. When you list a template in the marketplace, it becomes
          public: everyone can see it, search engines can find it, and it shows
          your Discord name and avatar and links to a page with all templates
          you listed.
        </p>
        <p>
          By making a template or a share code, you give other Kite users a
          free, non-exclusive, worldwide and permanent license to add it to
          their apps, use and change it there, and keep what they added if you
          change or delete the template later. You also give us the right to
          show and describe it in Kite and on kite.onl. Only share content you
          have the rights to, and never put secrets or other people&apos;s
          personal data in a template.
        </p>
        <p>
          Templates from other users are their creators&apos; responsibility,
          not ours. We review templates before they are listed in the
          marketplace, but we can&apos;t check every detail, and templates
          shared only by link aren&apos;t reviewed. Check what a template can
          do, which its page lists, before you add it.
        </p>
        <p>You may use and change the templates we make in your apps.</p>

        <h2 id="moderation">9. Template rules and moderation</h2>
        <p>
          Templates must follow the rules in{" "}
          <a href="#acceptable-use">section 6</a>. They must also not:
        </p>
        <ul>
          <li>
            harm the servers that add them, for example by deleting channels,
            banning members or giving anyone permissions the server didn&apos;t
            expect;
          </li>
          <li>
            send data about the server or its members to their creator or to
            other services without saying so in their description;
          </li>
          <li>mislead about what they do, or copy others&apos; templates;</li>
          <li>
            contain illegal, hateful, sexual or violent content, advertising or
            spam.
          </li>
        </ul>
        <p>
          When you submit a template to the marketplace, we review it by hand
          before it is listed, and we review new versions of listed templates
          the same way. Until a new version is approved, apps keep getting the
          last approved one. If we reject a template, its page tells you why.
        </p>
        <p>
          You can report a template on its page when you are logged in. Anyone
          can also report a template, or other content on Kite they believe is
          illegal, by email to{" "}
          <a href="mailto:contact@kite.onl">contact@kite.onl</a>, with the link
          and the reason. We look at every report by hand. If you report by
          email, we&apos;ll tell you what we decided.
        </p>
        <p>
          If a template breaks these rules, we can reject it, remove it from the
          marketplace or take it down. A template that was taken down can no
          longer be seen or added, and its creator sees why on its page. Apps
          that added it keep what they added. For serious or repeated
          violations, we can also suspend the creator&apos;s account. If you
          disagree with a decision, you can email us, write to us on Discord or
          go to court.
        </p>
        <p>
          Authorities, the European Commission and users can reach us about
          content on Kite at{" "}
          <a href="mailto:contact@kite.onl">contact@kite.onl</a>, in English or
          German.
        </p>

        <h2 id="availability">10. Availability and changes</h2>
        <p>
          We work to keep Kite running, but we can&apos;t promise it will always
          be available. Maintenance, problems with Discord or other services,
          and things outside our control can interrupt it. Integrations with
          other services only work while those services allow it.
        </p>
        <p>
          We keep developing Kite and may change or remove features. We&apos;ll
          tell you in advance about changes that significantly affect how you
          use Kite.
        </p>

        <h2 id="liability">11. Liability</h2>
        <p>
          We are liable without limitation for intent and gross negligence, for
          injury to life, body or health, under the German Product Liability
          Act, and where we have given a guarantee.
        </p>
        <p>
          For slight negligence, we are only liable if we breach an obligation
          that is essential for the contract and that you may regularly rely on,
          and only for the damage that was typical and foreseeable when the
          contract was made. Otherwise, we aren&apos;t liable for slight
          negligence. If data is lost through slight negligence, we are only
          liable for the effort it would have taken to restore it from copies
          you could reasonably have kept, like exports of your flows. These
          limits also apply to the people who work for us.
        </p>
        <p>
          If you culpably break these terms and someone makes a claim against us
          because of it, you will compensate us for the damage this causes,
          including reasonable legal costs.
        </p>

        <h2 id="termination">12. Ending the agreement</h2>
        <p>
          You can stop using Kite at any time. You can delete an app in its
          settings. To delete your account, email us at{" "}
          <a href="mailto:contact@kite.onl">contact@kite.onl</a>.
        </p>
        <p>
          We can end the agreement with two weeks&apos; notice. If you have a
          paid subscription, it ends at the end of its current billing period at
          the earliest. Both of us can end it without notice for good cause,
          like a serious or repeated violation of these terms. When the
          agreement ends, we stop running your bots and delete your data as
          described in the Privacy Policy.
        </p>

        <h2 id="changes">13. Changes to these terms</h2>
        <p>
          We may change these terms for a valid reason, like new features,
          changes in the law, or changes to Discord&apos;s terms. We&apos;ll
          tell you about changes at least two weeks before they apply to you, in
          the dashboard or by email. If you don&apos;t agree, you can stop using
          Kite and delete your apps before then. If you keep using Kite after
          that date, the new terms apply, and we&apos;ll point this out when we
          tell you about the change. Changes that only benefit you or that the
          law requires can apply right away. Changes to plans and prices follow{" "}
          <a href="#plans">section 7</a> instead.
        </p>

        <h2 id="law">14. Applicable law</h2>
        <p>
          German law applies, excluding the UN Convention on Contracts for the
          International Sale of Goods. If you are a consumer, this doesn&apos;t
          take away the protection of the mandatory laws of the country you live
          in. If you are a merchant, the courts of Leipzig have jurisdiction.
        </p>
        <p>
          We aren&apos;t obliged to and don&apos;t take part in dispute
          resolution proceedings before a consumer arbitration board.
        </p>
        <p>If a part of these terms is invalid, the rest stays valid.</p>

        <h2 id="contact">15. Contact</h2>
        <p>
          Questions about these terms? Email us at{" "}
          <a href="mailto:contact@kite.onl">contact@kite.onl</a>. More details
          are in the <Link href="/imprint">legal notice</Link>.
        </p>
      </div>

      <HomeFooter />
    </HomeLayout>
  );
}
