import { TemplateDetails } from "@/components/app/TemplateDetails";
import HomeCTASection from "@/components/home/HomeCTASection";
import HomeFooter from "@/components/home/HomeFooter";
import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { getTemplate, getTemplates, Template } from "@/lib/flow/templates";
import { GetStaticPaths, GetStaticProps } from "next";
import Link from "next/link";

type Props = { templateId: string };

// Every template is exported as its own page, so crawlers see its content
// without running JavaScript.
export const getStaticPaths: GetStaticPaths = () => ({
  paths: getTemplates().map((t) => ({ params: { templateId: t.id } })),
  fallback: false,
});

export const getStaticProps: GetStaticProps<Props> = ({ params }) => ({
  props: { templateId: String(params?.templateId) },
});

export default function TemplatePage({ templateId }: Props) {
  const template = getTemplate(templateId)!;

  return (
    <HomeLayout
      title={`${template.name} Template for Discord Bots`}
      description={`${template.description} A free template for Kite, the no-code Discord bot builder.`}
    >
      <div className="container max-w-5xl py-12">
        <TemplateDetails
          template={template}
          backHref="/templates"
          aside={<TemplateGetStartedCard template={template} />}
        />
      </div>

      <HomeCTASection />
      <HomeFooter />
    </HomeLayout>
  );
}

function TemplateGetStartedCard({ template }: { template: Template }) {
  return (
    <Card className="p-5">
      <div className="font-semibold mb-1">Use this template</div>
      <p className="text-sm text-muted-foreground mb-4">
        Create your bot in Kite for free, then add {template.name} from
        Templates in your app and change anything you like in the editor.
      </p>
      <Button className="w-full" asChild>
        <Link href="/apps">Get started</Link>
      </Button>
    </Card>
  );
}
