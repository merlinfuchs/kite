import { ReactNode } from "react";

export default function AppEmptyPlaceholder({
  title,
  description,
  action,
}: {
  title: string;
  description: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-sm h-full min-h-96">
      <div className="flex flex-col items-center gap-1 text-center">
        <h3 className="text-xl md:text-2xl font-bold tracking-tight">
          {title}
        </h3>
        <p className="text-sm text-muted-foreground mb-4">{description}</p>
        {action}
      </div>
    </div>
  );
}
