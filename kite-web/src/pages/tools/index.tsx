import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { MailPlusIcon, WebhookIcon } from "lucide-react";
import Link from "next/link";

export default function ToolsPage() {
  return (
    <HomeLayout title="Tools">
      <div className="py-10 px-5 max-w-4xl mx-auto">
        <div className="flex flex-col space-y-1.5 mb-5">
          <h1 className="text-2xl font-semibold leading-none tracking-tight">
            Discord Tools
          </h1>
          <p className="text-sm text-muted-foreground">
            A collection of useful tools around Kite and the Discord ecosystem.
          </p>
        </div>

        <div className="space-y-5">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-start">
                <CardTitle className="flex space-x-3">
                  <MailPlusIcon className="w-6 h-6 text-muted-foreground" />
                  <div>Message Creator</div>
                </CardTitle>
                <Button variant="outline" asChild>
                  <Link href="/tools/message-creator">Open tool</Link>
                </Button>
              </div>
              <CardDescription>
                Create good looking Discord messages and send them through
                webhooks.
              </CardDescription>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader>
              <div className="flex justify-between items-start">
                <CardTitle className="flex space-x-3">
                  <WebhookIcon className="w-6 h-6 text-muted-foreground" />
                  <div>Webhook Info</div>
                </CardTitle>
                <Button variant="outline" asChild>
                  <Link href="/tools/webhook-info">Open tool</Link>
                </Button>
              </div>
              <CardDescription>
                Get information about Discord webhooks from the webhook URL.
              </CardDescription>
            </CardHeader>
          </Card>
        </div>
      </div>
    </HomeLayout>
  );
}
