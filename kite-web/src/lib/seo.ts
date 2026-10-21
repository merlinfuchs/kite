import env from "@/lib/env/client";

export function faqJsonLd(faqs: { question: string; answer: string }[]) {
  return {
    "@context": "https://schema.org",
    "@type": "FAQPage",
    mainEntity: faqs.map(({ question, answer }) => ({
      "@type": "Question",
      name: question,
      acceptedAnswer: { "@type": "Answer", text: answer },
    })),
  };
}

export const softwareApplicationJsonLd = {
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  name: "Kite",
  url: env.NEXT_PUBLIC_SITE_URL,
  applicationCategory: "DeveloperApplication",
  operatingSystem: "Web",
  description:
    "Free, open source Discord bot maker. Build slash commands and event listeners in a visual editor without coding, hosted 24/7.",
  offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
};
