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

// Tells search engines that the page is a free tool, and where it sits.
export function toolJsonLd(title: string, path: string, description: string) {
  const url = env.NEXT_PUBLIC_SITE_URL + path;
  return [
    {
      "@context": "https://schema.org",
      "@type": "WebApplication",
      name: title,
      url,
      description,
      applicationCategory: "UtilitiesApplication",
      operatingSystem: "Web",
      isAccessibleForFree: true,
      offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
      publisher: {
        "@type": "Organization",
        name: "Kite",
        url: env.NEXT_PUBLIC_SITE_URL,
      },
    },
    {
      "@context": "https://schema.org",
      "@type": "BreadcrumbList",
      itemListElement: [
        {
          "@type": "ListItem",
          position: 1,
          name: "Tools",
          item: env.NEXT_PUBLIC_SITE_URL + "/tools",
        },
        { "@type": "ListItem", position: 2, name: title, item: url },
      ],
    },
  ];
}
