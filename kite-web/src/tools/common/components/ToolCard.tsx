import { cn } from "@/lib/utils";
import { Tool } from "@/tools/registry";
import { ArrowUpRightIcon } from "lucide-react";
import Link from "next/link";

export function ToolIcon({
  tool,
  className,
}: {
  tool: Tool;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "flex-none rounded-lg flex items-center justify-center",
        tool.color,
        className ?? "w-10 h-10"
      )}
    >
      <tool.icon className="w-1/2 h-1/2" />
    </div>
  );
}

export default function ToolCard({ tool }: { tool: Tool }) {
  return (
    <Link
      href={tool.href}
      target={tool.external ? "_blank" : undefined}
      className="flex gap-4 rounded-xl border bg-card p-5 transition-all hover:border-primary/50 hover:shadow-md"
    >
      <ToolIcon tool={tool} className="w-12 h-12" />
      <div>
        <div className="font-semibold flex items-center gap-1">
          {tool.title}
          {tool.external && (
            <ArrowUpRightIcon className="w-4 h-4 text-muted-foreground" />
          )}
        </div>
        <div className="text-sm text-muted-foreground mt-1">
          {tool.description}
        </div>
      </div>
    </Link>
  );
}
