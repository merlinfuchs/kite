import { LucideIcon } from "lucide-react";
import { ReactNode } from "react";

export default function AppEmptyPlaceholder({
  icon: Icon,
  title,
  description,
  action,
  footer,
}: {
  icon?: LucideIcon;
  title: string;
  description: string;
  action?: ReactNode;
  /** Shown below the action, for a link to templates or the docs. */
  footer?: ReactNode;
}) {
  return (
    <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-sm h-full min-h-96 px-6 py-10">
      <div className="flex flex-col items-center text-center max-w-md">
        {Icon ? (
          <div className="mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/15 text-primary">
            <Icon className="h-7 w-7" />
          </div>
        ) : null}
        <h3 className="text-xl md:text-2xl font-bold tracking-tight">
          {title}
        </h3>
        <p className="text-sm text-muted-foreground mt-1 mb-5">{description}</p>
        {action}
        {footer ? (
          <div className="mt-4 text-sm text-muted-foreground">{footer}</div>
        ) : null}
      </div>
    </div>
  );
}
