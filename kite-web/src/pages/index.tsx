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
import HomeTemplatesSection from "@/components/home/HomeTemplatesSection";
import HomeUseCasesSection from "@/components/home/HomeUseCasesSection";
import { TemplateIconSvgs } from "@/components/app/TemplateIcon";
import {
  fetchBundles,
  fetchPlans,
  fetchStats,
  Plans,
  Stats,
  withApiData,
} from "@/lib/apiData";
import {
  featuredTemplates,
  templateFromBundle,
  utcDay,
} from "@/lib/flow/templates";
import { templateIconSvgs } from "@/lib/flow/templateIconSvgs";
import { Bundle } from "@/lib/types/wire.gen";
import { faqJsonLd, softwareApplicationJsonLd } from "@/lib/seo";
import { GetStaticProps } from "next";

type Props = {
  stats: Stats;
  plans: Plans;
  // The templates shown in the templates section, null without the API.
  featuredBundles: Bundle[] | null;
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
  const ids = new Set(featuredTemplates(templates, utcDay()).map((t) => t.id));
  const featuredBundles = bundles ? bundles.filter((b) => ids.has(b.id)) : null;

  return withApiData({
    stats,
    plans,
    featuredBundles,
    templateCount: templates.length,
    templateIconSvgs: templateIconSvgs(
      (featuredBundles ?? []).map((b) => b.listing?.icon)
    ),
  });
};

export default function Home({
  stats,
  plans,
  featuredBundles,
  templateCount,
}: Props) {
  return (
    <HomeLayout jsonLd={[softwareApplicationJsonLd, faqJsonLd(FAQList)]}>
      <div className="relative overflow-hidden">
        <HomeHeroSection />
        <HomeSectionDivider
          from="background"
          to="muted"
          variant="horizon"
          accent="primary"
        />
      </div>

      <HomeStatsSection initial={stats} />

      <HomeUseCasesSection />
      <HomeTemplatesSection bundles={featuredBundles} total={templateCount} />
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
