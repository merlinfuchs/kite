import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  MousePointerClickIcon,
  SatelliteDishIcon,
  SquareSlash,
} from "lucide-react";
import dynamic from "next/dynamic";
import Reveal from "./Reveal";

const FlowExample = dynamic(() => import("../flow/FlowExample"), {
  ssr: false,
});

interface ServiceProps {
  title: string;
  description: string;
  icon: JSX.Element;
}

const serviceList: ServiceProps[] = [
  {
    title: "Custom Commands",
    description:
      "Create custom commands that can be used by your users to interact with your bot.",
    icon: <SquareSlash className="h-5 w-5" />,
  },
  {
    title: "Interactive Components",
    description:
      "Create buttons and dropdowns that your users can interact with and customize the behavior.",
    icon: <MousePointerClickIcon className="h-5 w-5" />,
  },
  {
    title: "Event Listeners",
    description:
      "Listen for events in your Discord server and respond to them with custom logic.",
    icon: <SatelliteDishIcon className="h-5 w-5" />,
  },
];

export default function HomeFlowSection() {
  return (
    <section id="flow" className="container py-24 sm:py-32">
      <div className="grid lg:grid-cols-[1fr,1fr] gap-8 place-items-center">
        <div>
          <Reveal from="left">
            <h2 className="text-3xl md:text-4xl font-bold">
              Powerful{" "}
              <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
                Visual Scripting
              </span>
            </h2>
          </Reveal>

          <Reveal from="left" delay={100}>
            <p className="text-muted-foreground text-xl mt-4 mb-8 ">
              Write custom logic for your slash commands, buttons, and more with
              the visual scripting interface of Kite.
            </p>
          </Reveal>

          <div className="flex flex-col gap-8">
            {serviceList.map(
              ({ icon, title, description }: ServiceProps, i) => (
                <Reveal key={title} from="left" delay={180 + i * 110}>
                  <Card className="home-card">
                    <CardHeader className="space-y-1 flex md:flex-row justify-start items-start gap-4">
                      <div className="mt-1 bg-primary/20 p-2 rounded-2xl">
                        {icon}
                      </div>
                      <div>
                        <CardTitle>{title}</CardTitle>
                        <CardDescription className="text-md mt-2">
                          {description}
                        </CardDescription>
                      </div>
                    </CardHeader>
                  </Card>
                </Reveal>
              )
            )}
          </div>
        </div>

        <Reveal
          from="right"
          duration={1100}
          className="w-full hidden sm:block md:w-[600px] lg:w-[600px] h-[800px]"
        >
          <FlowExample />
        </Reveal>
      </div>
    </section>
  );
}
