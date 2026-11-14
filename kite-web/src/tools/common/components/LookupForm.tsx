import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { LoaderCircleIcon, SearchIcon } from "lucide-react";
import { FormEvent, useState } from "react";

// The input of the lookup tools. They only look up when the form is sent, so
// typing an ID doesn't make a request for every digit.
export default function LookupForm({
  id,
  label,
  placeholder,
  loading,
  error,
  onLookup,
}: {
  id: string;
  label: string;
  placeholder: string;
  loading: boolean;
  error?: string;
  onLookup: (value: string) => void;
}) {
  const [value, setValue] = useState("");

  function submit(e: FormEvent) {
    e.preventDefault();
    if (value.trim()) onLookup(value.trim());
  }

  return (
    <form onSubmit={submit} className="space-y-2">
      <Label htmlFor={id}>{label}</Label>
      <div className="flex gap-2">
        <Input
          id={id}
          placeholder={placeholder}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          className="font-mono"
        />
        <Button type="submit" disabled={loading || !value.trim()}>
          {loading ? (
            <LoaderCircleIcon className="w-4 h-4 mr-2 animate-spin" />
          ) : (
            <SearchIcon className="w-4 h-4 mr-2" />
          )}
          Look up
        </Button>
      </div>
      {error && <p className="text-sm text-destructive">{error}</p>}
    </form>
  );
}
