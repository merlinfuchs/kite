import { bundleCapabilities } from "@/lib/flow/bundleCapabilities";
import { Template, templateFlows } from "@/lib/flow/templates";
import { TriangleAlertIcon } from "lucide-react";
import { useMemo } from "react";

// BundleCapabilities is what the blocks of a bundle a user made can do. Its
// flows run with the bot and credits of whoever adds it.
export default function BundleCapabilities({
  template,
}: {
  template: Template;
}) {
  const capabilities = useMemo(
    () => bundleCapabilities(templateFlows(template).map((f) => f.flow)),
    [template]
  );

  return (
    <section>
      <h2 className="font-semibold mb-1">What it can do</h2>
      <p className="text-sm text-muted-foreground mb-3 max-w-2xl">
        It runs with your bot and your credits, so check what it does before
        adding it.
      </p>
      {capabilities.length ? (
        <ul className="space-y-2">
          {capabilities.map((c) => (
            <li key={c} className="flex gap-2 text-sm">
              <TriangleAlertIcon className="h-4 w-4 flex-none mt-0.5 text-amber-500" />
              {c}
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-sm">
          Nothing that moderates members, deletes things, changes roles,
          contacts other websites or uses AI.
        </p>
      )}
    </section>
  );
}
