import BaseLayout from "@/components/common/BaseLayout";
import { Button } from "@/components/ui/button";
import Link from "next/link";

export default function NotFoundPage() {
  return (
    <BaseLayout title="Page Not Found">
      <div className="min-h-[100dvh] flex flex-col justify-center items-center">
        <div className="max-w-xl text-center text-foreground">
          <div className="text-7xl md:text-9xl font-bold mb-2 ">404</div>
          <h1 className="text-2xl md:text-4xl mb-8">page not found</h1>

          <Button asChild variant="outline">
            <Link href="/">Go Back</Link>
          </Button>
        </div>
      </div>
    </BaseLayout>
  );
}
