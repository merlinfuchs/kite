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
import HomeStatsSection from "@/components/home/HomeStatsSection";
import HomeTemplatesSection from "@/components/home/HomeTemplatesSection";
import HomeUseCasesSection from "@/components/home/HomeUseCasesSection";
import { faqJsonLd, softwareApplicationJsonLd } from "@/lib/seo";
import { TemplateIconSvgs } from "@/components/app/TemplateIcon";
import {
  fetchBundles,
  fetchPlans,
  fetchStats,
  Plans,
  Stats,
  withApiData,
} from "@/lib/apiData";
import { templateFromBundle } from "@/lib/flow/templates";
import { templateIconSvgs } from "@/lib/flow/templateIconSvgs";
import { Bundle } from "@/lib/types/wire.gen";
import { GetStaticProps } from "next";

type Props = {
  stats: Stats;
  plans: Plans;
  // The bundles of the most added templates, null without the API.
  topBundles: Bundle[] | null;
  templateCount: number;
  templateIconSvgs: TemplateIconSvgs;
};

export const getStaticProps: GetStaticProps<Props> = async () => {
  const [stats, plans, bundles] = await Promise.all([
    fetchStats(),
    fetchPlans(),
    fetchBundles(),
  ]);

  // Only the bundles of the shown templates go in the page's data.
  const templates = (bundles ?? []).flatMap((b) => templateFromBundle(b) ?? []);
  const topIds = templates
    .sort((a, b) => (b.installs ?? 0) - (a.installs ?? 0))
    .slice(0, 6)
    .map((t) => t.id);
  const topBundles = bundles
    ? topIds.flatMap((id) => bundles.find((b) => b.id === id) ?? [])
    : null;

  return withApiData({
    stats,
    plans,
    topBundles,
    templateCount: templates.length,
    templateIconSvgs: templateIconSvgs(
      (topBundles ?? []).map((b) => b.listing?.icon)
    ),
  });
};

export default function Home({
  stats,
  plans,
  topBundles,
  templateCount,
}: Props) {
  return (
    <HomeLayout jsonLd={[softwareApplicationJsonLd, faqJsonLd(FAQList)]}>
      <HomeHeroSection />
      <HomeStatsSection initial={stats} />
      <HomeUseCasesSection />
      <HomeTemplatesSection bundles={topBundles} total={templateCount} />
      <HomeHowItWorksSection />
      <HomeAISection />
      <HomeFeaturesSection />
      <HomeProfileSection />
      <HomePricingSection initial={plans} />
      <HomeFAQSection />
      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}
