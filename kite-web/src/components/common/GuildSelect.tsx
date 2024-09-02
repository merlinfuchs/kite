import { Input } from "../ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../ui/select";

export default function GuildSelect({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  /* return (
    <Select>
      <SelectTrigger className="w-full">
        <SelectValue placeholder="Server" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="light">Light</SelectItem>
        <SelectItem value="dark">Dark</SelectItem>
        <SelectItem value="system">System</SelectItem>
      </SelectContent>
    </Select>
  ); */

  return (
    <Input
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder="Server ID"
    />
  );
}
