import { Checkbox } from "../ui/checkbox";
import { Label } from "../ui/label";

export function FlowTestField({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      <Label className="text-sm">{label}</Label>
      {children}
    </div>
  );
}

// Picks some of the options, like a member does in a select menu.
export function FlowTestCheckboxList({
  options,
  value,
  onChange,
}: {
  options: { label: string; value: string }[];
  value: string[];
  onChange: (value: string[]) => void;
}) {
  return (
    <div className="flex flex-col gap-1.5">
      {options.map((o) => (
        <label key={o.value} className="flex items-center gap-2 text-sm">
          <Checkbox
            checked={value.includes(o.value)}
            onCheckedChange={(checked) =>
              onChange(
                checked
                  ? [...value, o.value]
                  : value.filter((v) => v !== o.value)
              )
            }
          />
          {o.label}
        </label>
      ))}
    </div>
  );
}
