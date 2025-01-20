import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";

export default function RefundPolicyPage() {
  return (
    <HomeLayout title="Refund Policy">
      <div className="prose dark:prose-invert text-foreground mx-auto my-16 max-w-7xl px-5">
        <h1>Refund Policy</h1>
        <p>Last updated: January 20, 2025</p>
        <p>Thank you for subscribing to Kite.onl.</p>
        <p>
          If, for any reason, You are not completely satisfied with a purchase
          We invite You to review our policy on refunds.
        </p>
        <p>
          The following terms are applicable for any products that You purchased
          with Us.
        </p>
        <h2>Interpretation and Definitions</h2>
        <h3>Interpretation</h3>
        <p>
          The words of which the initial letter is capitalized have meanings
          defined under the following conditions. The following definitions
          shall have the same meaning regardless of whether they appear in
          singular or in plural.
        </p>
        <h3>Definitions</h3>
        <p>For the purposes of this Refund Policy:</p>
        <ul>
          <li>
            <p>
              <strong>Company</strong> (referred to as either &quot;the
              Company&quot;, &quot;We&quot;, &quot;Us&quot; or &quot;Our&quot;
              in this Agreement) refers to Merlin Fuchs, Alte Str. 5, 04229
              Leipzig, Germany.
            </p>
          </li>
          <li>
            <p>
              <strong>Goods</strong> refer to the items offered for sale on the
              Service.
            </p>
          </li>
          <li>
            <p>
              <strong>Orders</strong> mean a request by You to purchase Goods
              from Us.
            </p>
          </li>
          <li>
            <p>
              <strong>Service</strong> refers to the Website.
            </p>
          </li>
          <li>
            <p>
              <strong>Website</strong> refers to Kite.onl, accessible from{" "}
              <a
                href="https://kite.onl"
                rel="external nofollow noopener"
                target="_blank"
              >
                https://kite.onl
              </a>
            </p>
          </li>
          <li>
            <p>
              <strong>You</strong> means the individual accessing or using the
              Service, or the company, or other legal entity on behalf of which
              such individual is accessing or using the Service, as applicable.
            </p>
          </li>
        </ul>
        <h2>Your Refund Rights</h2>
        <p>
          The deadline for requesting a refund is 14 days from the date on which
          you have made the payment.
        </p>
        <p>
          In order to exercise Your right to a refund, You must inform Us of
          your decision by means of a clear statement. You can inform us of your
          decision by:
        </p>
        <p>
          We ask You to provide a reason for your refund request so we can
          improve our service and prevent future issues.
        </p>
        <ul>
          <li>By email: contact@kite.onl</li>
        </ul>
        <p>
          We will reimburse You no later than 14 days from the day on which you
          have requested the refund. We will use the same means of payment as
          You used for the Order, which maybe be subject to fees depending on
          the payment provider.
        </p>
        <h3>Contact Us</h3>
        <p>
          If you have any questions about our Refund Policy, please contact us:
        </p>
        <ul>
          <li>By email: contact@kite.onl</li>
        </ul>
      </div>

      <HomeFooter />
    </HomeLayout>
  );
}
