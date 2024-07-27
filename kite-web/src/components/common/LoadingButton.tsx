import { LoaderCircleIcon } from "lucide-react";
import { Button, ButtonProps } from "../ui/button";
import { cn } from "@/lib/utils";

interface Props extends ButtonProps {
  loading: boolean;
}

export default function LoadingButton({ loading, children, ...props }: Props) {
  return (
    <Button
      {...props}
      disabled={loading}
      className={cn("flex items-center", props.className)}
    >
      {loading && <LoaderCircleIcon className="animate-spin h-5 w-5 mr-2" />}
      {children}
    </Button>
  );
}
