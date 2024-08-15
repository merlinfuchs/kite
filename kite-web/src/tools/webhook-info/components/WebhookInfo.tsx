import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import userAvatar from "@/tools/common/utils/discordCdn";

export interface WebhookData {
  id: string;
  name: string;
  avatar: string | null;
  type: number;
  token: string;
  channel_id: string;
  guild_id: string;
  application_id: string | null;
}

export default function WebhookInfo({ data }: { data: WebhookData }) {
  return (
    <Card>
      <div className="flex items-center">
        <img
          src={userAvatar(data)}
          className="rounded-full h-24 w-24 float float-end mt-5 ml-5"
          alt="Webhook avatar"
        />
        <CardHeader className="mt-5">
          <CardTitle>{data.name}</CardTitle>
          <CardDescription>{data.id}</CardDescription>
        </CardHeader>
      </div>

      <CardContent className="grid sm:grid-cols-2 mt-8 text-sm gap-5">
        <div>
          <div className="font-bold">Guild ID</div>
          <div className="text-muted-foreground">{data.guild_id}</div>
        </div>
        <div>
          <div className="font-bold">Channel ID</div>
          <div className="text-muted-foreground">{data.channel_id}</div>
        </div>
        <div>
          <div className="font-bold">Application ID</div>
          <div className="text-muted-foreground">
            {data.application_id || "-"}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
