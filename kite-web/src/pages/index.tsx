import HomeAISection from "@/components/home/HomeAISection";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQList } from "@/components/home/HomeFAQSections";
import HomeFeaturesSection from "@/components/home/HomeFeaturesSection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeHeroSection from "@/components/home/HomeHeroSection";
import HomeHowItWorksSection from "@/components/home/HomeHowItWorksSection";
import HomeLayout from "@/components/home/HomeLayout";
import HomePricingSection from "@/components/home/HomePricingSection";
import HomeStatsSection from "@/components/home/HomeStatsSection";
import HomeUseCasesSection from "@/components/home/HomeUseCasesSection";
import { faqJsonLd, softwareApplicationJsonLd } from "@/lib/seo";

export default function Home() {
  return (
    <HomeLayout jsonLd={[softwareApplicationJsonLd, faqJsonLd(FAQList)]}>
      <HomeHeroSection />
      <HomeStatsSection />
      <HomeUseCasesSection />
      <HomeHowItWorksSection />
      <HomeAISection />
      <HomeFeaturesSection />
      <HomePricingSection />
      <HomeFAQSection />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
