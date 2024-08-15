import HomeLayout from "@/components/home/HomeLayout";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ChevronLeftIcon } from "lucide-react";
import Link from "next/link";
import { ReactNode } from "react";

export default function ToolLayout({
  children,
  title,
  description,
  className,
}: {
  children: ReactNode;
  title: string;
  description: string;
  className?: string;
}) {
  return (
    <HomeLayout title={title}>
      <div className={cn("py-10 px-5 max-w-4xl mx-auto", className)}>
        <div className="mb-14">
          <Button variant="outline" asChild>
            <Link href="/tools">
              <ChevronLeftIcon className="h-5 w-5 mr-2 -ml-1" />
              Other Tools
            </Link>
          </Button>
        </div>
        <div className="flex flex-col space-y-2 mb-5">
          <h1 className="text-3xl font-semibold leading-none tracking-tight">
            {title}
          </h1>
          <p className="text-sm text-muted-foreground">{description}</p>
        </div>
        <div>{children}</div>
      </div>
    </HomeLayout>
  );
}
