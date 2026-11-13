import LegalPageLayout from "@/components/home/LegalPageLayout";
import { legalUpdatedAt } from "@/lib/legal";
import Link from "next/link";

export default function RefundPolicyPage() {
  return (
    <LegalPageLayout noindex title="Refund Policy" updatedAt={legalUpdatedAt}>
      <p>
        Kite&apos;s paid plans are sold by Lemon Squeezy as merchant of record,
        as described in our <Link href="/terms#plans">Terms of Service</Link>.
        This policy explains when you get your money back.
      </p>

      <h2 id="first-payment">1. Your first payment</h2>
      <p>
        You can get a full refund of the first payment of a subscription within
        14 days of it, for any reason. This applies once per app. Email us at{" "}
        <a href="mailto:contact@kite.onl">contact@kite.onl</a> with the email
        address you paid with or the app the subscription is for. You don&apos;t
        have to tell us why, but it helps us improve Kite.
      </p>
      <p>
        When we refund a payment, we also cancel the subscription, so it
        isn&apos;t charged again.
      </p>

      <h2 id="later">2. Renewals and later payments</h2>
      <p>
        We don&apos;t refund renewals, changes to another plan, or a first
        payment after 14 days. You can cancel a subscription at any time, and
        your app keeps its plan until the end of the period you paid for.
      </p>
      <p>
        If something went wrong on our side, like a payment that was charged
        twice, contact us and we&apos;ll refund it no matter how long ago it
        was.
      </p>

      <h2 id="payout">3. How you get the money</h2>
      <p>
        Lemon Squeezy pays the refund back to the payment method you used,
        within 14 days of your request. How long it then takes to show up
        depends on your bank or payment provider.
      </p>
      <p>
        Lemon Squeezy&apos;s{" "}
        <a
          href="https://www.lemonsqueezy.com/buyer-terms"
          target="_blank"
          rel="noreferrer"
        >
          buyer terms
        </a>{" "}
        also apply to your purchase. This policy gives you a refund even where
        they don&apos;t.
      </p>
    </LegalPageLayout>
  );
}
