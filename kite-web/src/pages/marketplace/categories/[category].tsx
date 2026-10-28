import { TemplateIconSvgs } from "@/components/app/TemplateIcon";
import { TemplateList } from "@/components/app/TemplateList";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import MarketplaceBar from "@/components/home/MarketplaceBar";
import { fetchBundles, withApiData } from "@/lib/apiData";
import { templateIconSvgs } from "@/lib/flow/templateIconSvgs";
import {
  categoryPath,
  TemplateCategory,
  templateCategories,
  templateFromBundle,
  templatePath,
} from "@/lib/flow/templates";
import { Bundle } from "@/lib/types/wire.gen";
import { GetStaticPaths, GetStaticProps } from "next";
import { useMemo } from "react";

type Props = {
  category: TemplateCategory;
  // The templates of the category, the official ones first.
  bundles: Bundle[];
  templateIconSvgs: TemplateIconSvgs;
};

export const getStaticPaths: GetStaticPaths = () => ({
  paths: templateCategories.map((c) => categoryPath(c.value)),
  fallback: false,
});

export const getStaticProps: GetStaticProps<Props> = async ({ params }) => {
  const category = templateCategories.find(
    (c) => c.value === params?.category
  )!.value;
  const bundles = ((await fetchBundles()) ?? []).filter(
    (b) => templateFromBundle(b)?.category === category
  );
  return withApiData(
    {
      category,
      bundles,
      templateIconSvgs: templateIconSvgs(bundles.map((b) => b.listing?.icon)),
    },
    60
  );
};

// CategoryPage lists all templates of a category, which the marketplace only
// shows the most added of.
export default function CategoryPage({ category, bundles }: Props) {
  const templates = useMemo(
    () => bundles.flatMap((b) => templateFromBundle(b) ?? []),
    [bundles]
  );
  const { title, description } = templateCategories.find(
    (c) => c.value === category
  )!;

  return (
    <HomeLayout
      title={title}
      description={`${description} Free templates for Discord bots made with Kite.`}
    >
      <MarketplaceBar category={category} search={false} />
      <TemplateList
        templates={templates}
        href={(t) => templatePath(t.id)}
        buildHref="/apps"
        rows={false}
        hero={
          <div>
            <h1 className="text-3xl md:text-4xl font-bold">{title}</h1>
            <p className="text-lg text-muted-foreground mt-2">{description}</p>
          </div>
        }
      />
      <HomeCTASection light />
      <HomeFooter />
    </HomeLayout>
  );
}
