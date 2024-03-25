import { ReactNode } from "react";
import { Button } from "../ui/button";

interface Props {
  title: string;
  description: string;
  action?: ReactNode;
}

export default function AppGuildPageEmpty({
  title,
  description,
  action,
}: Props) {
  return (
    <div className="flex flex-1 items-center justify-center rounded-lg border border-dashed shadow-sm">
      <div className="flex flex-col items-center gap-1 text-center">
        <h3 className="text-2xl font-bold tracking-tight">{title}</h3>
        <p className="text-sm text-muted-foreground">{description}</p>
        {!!action && <div className="mt-4">{action}</div>}
      </div>
    </div>
  );
}
