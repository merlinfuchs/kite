import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  FileSearchIcon,
  FolderSearchIcon,
  ImageIcon,
  MailPlusIcon,
  PaletteIcon,
  SearchIcon,
  SnowflakeIcon,
  UserSearchIcon,
  WebhookIcon,
} from "lucide-react";
import Link from "next/link";

const tools = [
  {
    title: "Message Creator",
    description:
      "Create good looking Discord messages and send them through webhooks.",
    icon: MailPlusIcon,
    href: "https://message.style/app",
    target: "_blank",
  },
  {
    title: "Colored Text",
    description:
      "Generate colored text that you can use in your Discord message.",
    icon: PaletteIcon,
    href: "https://message.style/app/tools/colored-text",
    target: "_blank",
  },
  {
    title: "Embed Links",
    description:
      "Generate embeddable links for Discord messages with custom titles, descriptions, and images.",
    icon: ImageIcon,
    href: "https://message.style/app/tools/embed-links",
    target: "_blank",
  },
  {
    title: "Snowflake Decoder",
    description: "Get general information about a Discord ID.",
    icon: SnowflakeIcon,
    href: "https://dis.wtf/lookup/snowflake",
    target: "_blank",
  },
  {
    title: "User Lookup",
    description:
      "Get information about a Discord user by entering their user ID.",
    icon: UserSearchIcon,
    href: "https://dis.wtf/lookup/user",
    target: "_blank",
  },
  {
    title: "Server Lookup",
    description:
      "Get information about a Discord server by entering its server ID.",
    icon: FolderSearchIcon,
    href: "https://dis.wtf/lookup/guild",
    target: "_blank",
  },
  {
    title: "Invite Resolver",
    description:
      "Get information about a Discord server by entering an invite code.",
    icon: FolderSearchIcon,
    href: "https://dis.wtf/lookup/invite",
    target: "_blank",
  },
  {
    title: "Application Lookup",
    description:
      "Get information about a Discord application by entering its app ID.",
    icon: FileSearchIcon,
    href: "https://dis.wtf/lookup/app",
    target: "_blank",
  },
  {
    title: "Webhook Info",
    description: "Get information about Discord webhooks from the webhook URL.",
    icon: WebhookIcon,
    href: "/tools/webhook-info",
  },
];

export default function ToolsPage() {
  return (
    <HomeLayout title="Tools">
      <div className="py-20 px-5 max-w-4xl mx-auto">
        <div className="flex flex-col space-y-2 mb-10">
          <h1 className="text-3xl font-semibold leading-none tracking-tight">
            Discord Tools
          </h1>
          <p className="text-sm text-muted-foreground">
            A collection of useful tools around the Discord ecosystem and Kite.
          </p>
        </div>

        <div className="grid gap-6 lg:grid-cols-2">
          {tools.map((tool, i) => (
            <Card key={i}>
              <CardHeader>
                <div className="flex justify-between items-start">
                  <CardTitle className="flex space-x-3">
                    <tool.icon className="w-6 h-6 text-muted-foreground" />
                    <div>{tool.title}</div>
                  </CardTitle>
                  <Button variant="outline" asChild className="lg:mb-2">
                    <Link href={tool.href} target={tool.target}>
                      Open tool
                    </Link>
                  </Button>
                </div>
                <CardDescription>{tool.description}</CardDescription>
              </CardHeader>
            </Card>
          ))}
        </div>
      </div>
    </HomeLayout>
  );
}
