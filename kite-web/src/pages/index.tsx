import HomeAISection from "@/components/home/HomeAISection";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFAQSection, { FAQList } from "@/components/home/HomeFAQSections";
import HomeFeaturesSection from "@/components/home/HomeFeaturesSection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeHeroSection from "@/components/home/HomeHeroSection";
import HomeHowItWorksSection from "@/components/home/HomeHowItWorksSection";
import HomeLayout from "@/components/home/HomeLayout";
import HomePricingSection from "@/components/home/HomePricingSection";
import HomeProfileSection from "@/components/home/HomeProfileSection";
import HomeSectionDivider from "@/components/home/HomeSectionDivider";
import HomeStatsSection from "@/components/home/HomeStatsSection";
import HomeUseCasesSection from "@/components/home/HomeUseCasesSection";
import {
  fetchPlans,
  fetchStats,
  Plans,
  Stats,
  withApiData,
} from "@/lib/apiData";
import { faqJsonLd, softwareApplicationJsonLd } from "@/lib/seo";
import { GetStaticProps } from "next";
import { useEffect } from "react";

type Props = { stats: Stats; plans: Plans };

export const getStaticProps: GetStaticProps<Props> = async () => {
  const [stats, plans] = await Promise.all([fetchStats(), fetchPlans()]);
  return withApiData({ stats, plans });
};

export default function Home({ stats, plans }: Props) {
  useEffect(() => {
    if (!window.matchMedia("(prefers-reduced-motion: no-preference)").matches) {
      return;
    }
    const root = document.documentElement;
    const prev = root.style.scrollBehavior;
    root.style.scrollBehavior = "smooth";
    return () => {
      root.style.scrollBehavior = prev;
    };
  }, []);

  return (
    <HomeLayout jsonLd={[softwareApplicationJsonLd, faqJsonLd(FAQList)]}>
      <HomeHeroSection />
      <HomeSectionDivider
        from="background"
        to="muted"
        variant="horizon"
        accent="primary"
      />

      <HomeStatsSection initial={stats} />

      <HomeUseCasesSection />
      <HomeSectionDivider
        from="background"
        to="muted"
        variant="arch"
        accent="discord"
      />

      <HomeHowItWorksSection />

      <HomeAISection />
      <HomeSectionDivider
        from="background"
        to="muted"
        variant="wave"
        accent="primary"
      />

      <HomeFeaturesSection />
      <HomeSectionDivider
        from="muted"
        to="background"
        variant="wave"
        accent="none"
      />

      <HomeProfileSection />
      <HomeSectionDivider
        from="background"
        to="muted"
        variant="horizon"
        accent="primary"
      />

      <div className="bg-muted/30">
        <HomePricingSection initial={plans} />
      </div>

      <HomeFAQSection />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
