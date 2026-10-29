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
import {
  fetchPlans,
  fetchStats,
  Plans,
  Stats,
  withApiData,
} from "@/lib/apiData";
import { GetStaticProps } from "next";

type Props = { stats: Stats; plans: Plans };

export const getStaticProps: GetStaticProps<Props> = async () => {
  const [stats, plans] = await Promise.all([fetchStats(), fetchPlans()]);
  return withApiData({ stats, plans });
};

export default function Home({ stats, plans }: Props) {
  return (
    <HomeLayout jsonLd={[softwareApplicationJsonLd, faqJsonLd(FAQList)]}>
      <HomeHeroSection />
      <HomeStatsSection initial={stats} />
      <HomeUseCasesSection />
      <HomeHowItWorksSection />
      <HomeAISection />
      <HomeFeaturesSection />
      <HomePricingSection initial={plans} />
      <HomeFAQSection />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
