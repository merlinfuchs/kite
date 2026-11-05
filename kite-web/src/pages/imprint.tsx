import LegalPageLayout from "@/components/home/LegalPageLayout";
import env from "@/lib/env/client";

export default function ImprintPage() {
  return (
    <LegalPageLayout
      noindex
      title="Legal Notice"
      heading="Legal Notice (Impressum)"
    >
      <p>Information according to &sect; 5 DDG:</p>
      <p>
        Merlin Fuchs
        <br />
        Alte Str. 5
        <br />
        04229 Leipzig
        <br />
        Germany
      </p>
      <h2 id="contact">Contact</h2>
      <p>
        Email: <a href="mailto:contact@kite.onl">contact@kite.onl</a>
        <br />
        Discord:{" "}
        <a href={env.NEXT_PUBLIC_DISCORD_LINK} target="_blank" rel="noreferrer">
          Kite support server
        </a>
      </p>
      <p>
        This is also our point of contact for authorities, the European
        Commission and users under Art. 11 and 12 of the Digital Services Act.
        You can write to us in English or German.
      </p>
      <h2 id="vat">VAT ID</h2>
      <p>VAT identification number according to &sect; 27a UStG: DE355903271</p>
      <h2 id="disputes">Consumer dispute resolution</h2>
      <p>
        We aren&apos;t obliged to and don&apos;t take part in dispute resolution
        proceedings before a consumer arbitration board.
      </p>
    </LegalPageLayout>
  );
}
