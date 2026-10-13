import { FAQProps } from "@/components/home/HomeFAQSections";
import { ReactNode } from "react";

// Explains a tool below it, which is also what search engines read.
export default function ToolArticle({
  children,
  faqs,
}: {
  children: ReactNode;
  faqs: FAQProps[];
}) {
  return (
    <div className="mt-16 prose dark:prose-invert text-foreground max-w-none prose-code:before:content-none prose-code:after:content-none">
      {children}
      <h2>Frequently Asked Questions</h2>
      {faqs.map((faq) => (
        <div key={faq.question}>
          <h3>{faq.question}</h3>
          <p>{faq.answer}</p>
        </div>
      ))}
    </div>
  );
}
