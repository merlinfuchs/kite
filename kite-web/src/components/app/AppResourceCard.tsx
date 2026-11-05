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
  limit,
  actionTitle,
  actionHref,
}: {
  title: string;
  count: number;
  // The plan's limit, 0 or undefined when there is none.
  limit?: number;
  actionTitle: string;
  actionHref: Partial<Url> | string;
}) {
  return (
    <Card>
      <CardHeader className="pb-4">
        <CardDescription>{title}</CardDescription>
        <CardTitle className="text-4xl">
          {count}
          {!!limit && (
            <span className="text-xl font-normal text-muted-foreground">
              {" "}
              / {limit}
            </span>
          )}
        </CardTitle>
      </CardHeader>
      <CardFooter>
        <Button size="sm" variant="outline" asChild>
          <Link href={actionHref}>{actionTitle}</Link>
        </Button>
      </CardFooter>
    </Card>
  );
}
