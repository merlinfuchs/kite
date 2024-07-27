import Link from "next/link";
import { Button } from "../ui/button";
import {
  Card,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../ui/card";
import { Url } from "url";

export default function AppResourceCard({
  title,
  count,
  actionTitle,
  actionHref,
}: {
  title: string;
  count: number;
  actionTitle: string;
  actionHref: Partial<Url> | string;
}) {
  return (
    <Card>
      <CardHeader className="pb-4">
        <CardDescription>{title}</CardDescription>
        <CardTitle className="text-4xl">{count}</CardTitle>
      </CardHeader>
      <CardFooter>
        <Button size="sm" variant="outline" asChild>
          <Link href={actionHref}>{actionTitle}</Link>
        </Button>
      </CardFooter>
    </Card>
  );
}
