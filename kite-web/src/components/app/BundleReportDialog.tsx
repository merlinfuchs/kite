import { ReactNode, useState } from "react";
import { toast } from "sonner";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../ui/dialog";
import { Button } from "../ui/button";
import { Label } from "../ui/label";
import { RadioGroup, RadioGroupItem } from "../ui/radio-group";
import { Textarea } from "../ui/textarea";
import LoadingButton from "../common/LoadingButton";
import { useBundleReportCreateMutation } from "@/lib/api/mutations";
import { loginUrlTo } from "@/lib/api/client";
import { useUserQuery } from "@/lib/api/queries";
import { useRouter } from "next/router";

const categories = [
  {
    value: "harmful",
    label: "It harms servers, like banning members or deleting channels",
  },
  { value: "scam", label: "It's a scam or collects data it shouldn't" },
  {
    value: "inappropriate",
    label: "It's hateful, sexual or otherwise inappropriate",
  },
  { value: "spam", label: "It's spam or doesn't do what it says" },
  { value: "other", label: "Something else" },
];

// BundleReportDialog reports a bundle of another user that breaks the rules.
// Kite looks at the reports and can take the bundle down.
export default function BundleReportDialog({
  bundleId,
  name,
  children,
}: {
  bundleId: string;
  name: string;
  children: ReactNode;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent>
        {open && (
          <BundleReportForm
            bundleId={bundleId}
            name={name}
            onDone={() => setOpen(false)}
          />
        )}
      </DialogContent>
    </Dialog>
  );
}

function BundleReportForm({
  bundleId,
  name,
  onDone,
}: {
  bundleId: string;
  name: string;
  onDone: () => void;
}) {
  const router = useRouter();
  const { data: user } = useUserQuery();
  const reportMutation = useBundleReportCreateMutation();
  const [category, setCategory] = useState("");
  const [details, setDetails] = useState("");

  async function submit() {
    const res = await reportMutation.mutateAsync({
      bundleId,
      req: { category, details: details.trim() },
    });
    if (!res.success) {
      toast.error(
        `Failed to report ${name}: ${res.error.message} (${res.error.code})`
      );
      return;
    }
    toast.success(`Reported ${name}, thanks for telling us`);
    onDone();
  }

  return (
    <>
      <DialogHeader>
        <DialogTitle>Report {name}</DialogTitle>
        <DialogDescription>
          Kite looks at every report and takes down templates that break the
          rules. Whoever made it doesn&apos;t see who reported it.
        </DialogDescription>
      </DialogHeader>

      {user && !user.success ? (
        <Button asChild>
          <a href={loginUrlTo(router.asPath)}>Log in to report it</a>
        </Button>
      ) : (
        <>
          <RadioGroup value={category} onValueChange={setCategory}>
            {categories.map((c) => (
              <div key={c.value} className="flex items-center gap-2">
                <RadioGroupItem value={c.value} id={`report-${c.value}`} />
                <Label htmlFor={`report-${c.value}`} className="font-normal">
                  {c.label}
                </Label>
              </div>
            ))}
          </RadioGroup>
          <Textarea
            value={details}
            maxLength={1000}
            rows={4}
            placeholder="What's wrong with it? Name the command or listener if you can."
            onChange={(e) => setDetails(e.target.value)}
          />
          <DialogFooter className="gap-2">
            <DialogClose asChild>
              <Button variant="outline">Cancel</Button>
            </DialogClose>
            <LoadingButton
              onClick={submit}
              loading={reportMutation.isPending}
              disabled={!category || !details.trim()}
            >
              Report
            </LoadingButton>
          </DialogFooter>
        </>
      )}
    </>
  );
}
