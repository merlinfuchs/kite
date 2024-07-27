import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  CodeIcon,
  DatabaseZapIcon,
  EditIcon,
  EyeIcon,
  GamepadIcon,
  RouterIcon,
  ServerIcon,
  UsersIcon,
} from "lucide-react";

interface FeatureProps {
  icon: JSX.Element;
  title: string;
  description: string;
}

const features: FeatureProps[] = [
  {
    icon: <DatabaseZapIcon />,
    title: "24/7 Hosting",
    description:
      "Kite provides 24/7 hosting for your Discord bot, so you don't have to worry about uptime.",
  },

  {
    icon: <EditIcon />,
    title: "Customization",
    description:
      "Customize the look and feel of your Discord bot with Kite's easy-to-use interface.",
  },
  {
    icon: <CodeIcon />,
    title: "No Code",
    description:
      "You can create your own Discord bot without writing a single line of code.",
  },
  {
    icon: <UsersIcon />,
    title: "Collaboration",
    description:
      "Work together to create the perfect Discord bot for your server.",
  },
];

export default function HomeFeaturesSection() {
  return (
    <section id="features" className="container text-center py-24 sm:py-32">
      <h2 className="text-3xl md:text-4xl font-bold ">
        <span className="bg-gradient-to-b from-primary/60 to-primary text-transparent bg-clip-text">
          Everything{" "}
        </span>
        In One Place
      </h2>
      <p className="md:w-3/4 mx-auto mt-4 mb-8 text-xl text-muted-foreground">
        Kite provides all the tools you need to create a Discord bot for your
        server.
      </p>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8">
        {features.map(({ icon, title, description }: FeatureProps) => (
          <Card key={title} className="bg-muted/50">
            <CardHeader>
              <CardTitle className="grid gap-4 place-items-center">
                {icon}
                {title}
              </CardTitle>
            </CardHeader>
            <CardContent>{description}</CardContent>
          </Card>
        ))}
      </div>
    </section>
  );
}
